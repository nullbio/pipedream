package pipedream

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStaticHandler(t *testing.T) {
	t.Parallel()

	outFile1 := filepath.Join(testTmp, "static", "assets", "js", "transform_file-a1b2c3.js")
	outFile2 := filepath.Join(testTmp, "static", "assets", "css", "transform_file-a1b2c3.css.gz")

	if err := os.MkdirAll(filepath.Join(testTmp, "static", "assets", "js"), 0775); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll(filepath.Join(testTmp, "static", "assets", "css"), 0775); err != nil {
		t.Error(err)
	}

	if err := ioutil.WriteFile(outFile1, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(outFile2, []byte(testTransformFileGZ), 0664); err != nil {
		t.Fatal(err)
	}

	var p Pipedream
	p.Out = filepath.Join(testTmp, "static")
	p.Manifest.Files = map[string]FileInfo{
		"/assets/js/transform_file-a1b2c3.js": FileInfo{
			MTime:  time.Now(),
			Size:   uint64(len(testTransformFile)),
			Digest: "a1b2c3",
		},
		"/assets/css/transform_file-a1b2c3.css": FileInfo{
			MTime:  time.Now(),
			Size:   uint64(len(testTransformFile)),
			Digest: "a1b2c3",
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/transform_file-a1b2c3.js", nil)
		p.StaticHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		md5 := w.Header().Get("Content-Md5")
		etag := w.Header().Get("ETag")

		if md5 != "a1b2c3" {
			t.Error("content-md5 was wrong:", md5)
		}

		if etag != `"a1b2c3"` {
			t.Error("etag was wrong:", etag)
		}

		if bs := w.Body.String(); bs != testTransformFile {
			t.Errorf("body mismatch, got:\n%s", bs)
		}
	})

	t.Run("Gzip", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/css/transform_file-a1b2c3.css", nil)
		r.Header.Set("Accept-Encoding", "gzip;q=1.0, identity; q=0.5, *;q=0")
		p.StaticHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		cEnc := w.Header().Get("Content-Encoding")
		if cEnc != "gzip" {
			t.Error("wanted gzip, got:", cEnc)
		}

		if bs := w.Body.String(); bs != testTransformFileGZ {
			t.Errorf("body mismatch, got:\n%s", bs)
		}

	})

	t.Run("GzipStar", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/css/transform_file-a1b2c3.css", nil)
		r.Header.Set("Accept-Encoding", "identity; q=0.5, *;q=0")
		p.StaticHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		cEnc := w.Header().Get("Content-Encoding")
		if cEnc != "gzip" {
			t.Error("wanted gzip, got:", cEnc)
		}

		if bs := w.Body.String(); bs != testTransformFileGZ {
			t.Errorf("body mismatch, got:\n%s", bs)
		}
	})

	t.Run("PreventFolderTraversal", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/../assets/css/transform_file-a1b2c3.css", nil)
		p.StaticHandler(nil).ServeHTTP(w, r)

		if w.Code == http.StatusOK {
			t.Fatal("did not want status ok")
		}
	})

	t.Run("NoServeDirs", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/assets/css/", nil)
		p.StaticHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatal("wanted 404, got:", w.Code)
		}
	})
}

func TestDynamicHandler(t *testing.T) {
	t.Parallel()

	if err := os.MkdirAll(filepath.Join(testTmp, "dynamic", "assets", "js"), 0775); err != nil {
		t.Error(err)
	}

	var p Pipedream
	p.In = filepath.Join(testTmp, "dynamic", "assets")
	p.Out = filepath.Join(testTmp, "dynamic", "cached")
	p.NoHash = true

	t.Run("Success", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{
			"cat": Command{
				Cmd:    "cat",
				Stdin:  true,
				Stdout: true,
			},
			"tee": Command{
				Cmd:   "tee",
				Args:  []string{"$outfile"},
				Stdin: true,
			},
		}

		p.JS.Minifier = Command{
			Cmd:    "cat",
			Args:   []string{"$infile"},
			Stdout: true,
		}

		inFile := filepath.Join(testTmp, "dynamic", "assets", "js", "transform_file.js.tee.cat")
		outFile := filepath.Join(testTmp, "dynamic", "cached", "assets", "js", "transform_file.js")
		if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/transform_file.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		if bs := w.Body.String(); bs != testTransformFile {
			t.Errorf("body mismatch, got:\n%s", bs)
		}

		_, err := os.Stat(outFile)
		if err != nil {
			t.Errorf("Expected output file at %s", outFile)
		}
	})

	t.Run("ServeNoTransform", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{}
		p.JS.Minifier = Command{}

		inFile := filepath.Join(testTmp, "dynamic", "assets", "js", "transform_file1.js")
		outFile := filepath.Join(testTmp, "dynamic", "cached", "assets", "js", "transform_file1.js")
		if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/transform_file1.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		if bs := w.Body.String(); bs != testTransformFile {
			t.Errorf("body mismatch, got:\n%s", bs)
		}

		_, err := os.Stat(outFile)
		if err == nil {
			t.Errorf("Expected output file to be missing")
		}
	})

	t.Run("ServeCached", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{
			"cat": Command{
				Cmd:    "cat",
				Stdin:  true,
				Stdout: true,
			},
			"tee": Command{
				Cmd:   "tee",
				Args:  []string{"$outfile"},
				Stdin: true,
			},
		}

		p.JS.Minifier = Command{
			Cmd:    "cat",
			Args:   []string{"$infile"},
			Stdout: true,
		}

		inFile := filepath.Join(testTmp, "dynamic", "assets", "js", "transform_file2.js.tee.cat")
		outFile := filepath.Join(testTmp, "dynamic", "cached", "assets", "js", "transform_file2.js")
		if err := ioutil.WriteFile(inFile, []byte(`old file`), 0664); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 10)
		if err := ioutil.WriteFile(outFile, []byte(testTransformFile), 0664); err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/transform_file2.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		if bs := w.Body.String(); bs != testTransformFile {
			t.Errorf("body mismatch, got:\n%s", bs)
		}

		_, err := os.Stat(outFile)
		if err != nil {
			t.Errorf("Expected output file at %s", outFile)
		}
	})

	t.Run("RefreshCached", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{
			"cat": Command{
				Cmd:    "cat",
				Stdin:  true,
				Stdout: true,
			},
			"tee": Command{
				Cmd:   "tee",
				Args:  []string{"$outfile"},
				Stdin: true,
			},
		}

		p.JS.Minifier = Command{
			Cmd:    "cat",
			Args:   []string{"$infile"},
			Stdout: true,
		}

		inFile := filepath.Join(testTmp, "dynamic", "assets", "js", "transform_file3.js.tee.cat")
		outFile := filepath.Join(testTmp, "dynamic", "cached", "assets", "js", "transform_file3.js")
		if err := ioutil.WriteFile(outFile, []byte(`old file`), 0664); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 10)
		if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/transform_file3.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("wanted status ok, got:", w.Code)
		}

		if bs := w.Body.String(); bs != testTransformFile {
			t.Errorf("body mismatch, got:\n%s", bs)
		}

		_, err := os.Stat(outFile)
		if err != nil {
			t.Errorf("Expected output file at %s", outFile)
		}
	})

	t.Run("BadTypeNotFound", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{}
		p.JS.Minifier = Command{}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/bad/transform_file.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatal("wanted status not found, got:", w.Code)
		}
	})

	t.Run("BadPathNotFound", func(t *testing.T) {
		p.JS.Compilers = map[string]Command{}
		p.JS.Minifier = Command{}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/assets/js/", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatal("wanted status not found, got:", w.Code)
		}

		r = httptest.NewRequest("GET", "/assets/js/thing.js", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatal("wanted status not found, got:", w.Code)
		}

		r = httptest.NewRequest("GET", "/assets/js/thing.js/", nil)
		p.DynamicHandler(nil).ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatal("wanted status not found, got:", w.Code)
		}
	})
}
