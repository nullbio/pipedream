package pipedream

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StaticHandler serves static assets from the out path encouraging browser
// caching. StaticHandler disables directory listings. It serves a gzipped
// file if requested through Accept-Encoding.
type StaticHandler struct {
	*Pipedream
	log *log.Logger
}

// StaticHandler returns a StaticHandler object with a logger
func (p *Pipedream) StaticHandler(log *log.Logger) http.Handler {
	return StaticHandler{
		Pipedream: p,
		log:       log,
	}
}

func (s StaticHandler) logf(format string, v ...interface{}) {
	if s.log != nil {
		s.log.Printf(format, v...)
	}
}

// ServeHTTP serves static assets from the out path.
func (s StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.Replace(strings.Replace(r.URL.Path, "..", "", -1), string(os.PathSeparator)+".", "", -1)

	fileDets, ok := s.Manifest.Files[urlPath]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !s.NoCompress {
		acceptEncodings := r.Header.Get("Accept-Encoding")
		encodings := strings.Split(acceptEncodings, ",")
		useGzip := false
		for _, e := range encodings {
			e = strings.TrimSpace(e)
			if strings.HasPrefix(e, "gzip") || strings.HasPrefix(e, "*") {
				useGzip = true
				break
			}
		}

		if useGzip {
			urlPath += ".gz"
			w.Header().Set("Content-Encoding", "gzip")
		}
	}

	fileLoc := filepath.Join(s.Out, urlPath)
	fileStat, err := os.Stat(fileLoc)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		s.logf("failed to stat file %s: %v", fileLoc, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fileStat.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	file, err := os.Open(fileLoc)
	if err != nil {
		s.logf("failed to open file %s: %v", fileLoc, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Md5", fileDets.Digest)
	w.Header().Set("ETag", `"`+fileDets.Digest+`"`)
	http.ServeContent(w, r, urlPath, fileDets.MTime, file)

	_ = file.Close()
}

// DynamicHandler automatically recompiles assets that have changed
// since it last recompiled them. It makes an effort to disable
// browser-side caching.
type DynamicHandler struct {
	*Pipedream
	log *log.Logger
}

// DynamicHandler returns a DynamicHandler object with a logger
func (p *Pipedream) DynamicHandler(log *log.Logger) http.Handler {
	return DynamicHandler{
		Pipedream: p,
		log:       log,
	}
}

func (d DynamicHandler) logf(format string, v ...interface{}) {
	if d.log != nil {
		d.log.Printf(format, v...)
	}
}

func (d DynamicHandler) validURL(reqURL string) ([]string, *Exes, bool) {
	// Return not found for directories, only files are permitted
	if reqURL[len(reqURL)-1] == '/' {
		return nil, nil, false
	}

	reqURL = strings.TrimPrefix(reqURL, "/")
	chunks := strings.Split(reqURL, "/")

	ln := len(chunks)
	if ln < 3 && ln > 0 && chunks[0] != "assets" {
		return nil, nil, false
	}

	typ := chunks[1]
	exes, ok := d.exes(typ)
	return chunks, &exes, ok
}

type fileInfo struct {
	// path to file in input folder. uses ReadDir to locate it if the
	// extensions on fileName are not an exact match.
	inPath string

	// path to file in output folder. uses input file path if serving
	// a non-compiled asset.
	outPath string
}

func (d DynamicHandler) getFileInfo(chunks []string) (fileInfo, error) {
	info := fileInfo{}

	fileName := chunks[len(chunks)-1]

	// in path without assets folder and filename
	inRelPath := strings.Join(chunks[1:len(chunks)-1], string(os.PathSeparator))

	// full in path without filename
	inPath := filepath.Join(d.In, inRelPath)

	fnames, err := ioutil.ReadDir(inPath)
	if err != nil && !os.IsNotExist(err) {
		d.logf("failed to list dir %s: %v", inPath, err)
	}
	if err != nil {
		return info, err
	}

	matchFileName := findFile(fnames, fileName)

	if matchFileName == "" {
		return info, os.ErrNotExist
	}

	info.inPath = filepath.Join(inPath, matchFileName)
	info.outPath = filepath.Join(d.Out, inRelPath, fileName)

	return info, nil
}

// findFile finds the requested file in the in directory.
// If it cannot find a full match it will do a prefix match.
func findFile(fnames []os.FileInfo, fileName string) string {
	a := strings.Split(fileName, ".")

OuterFor:
	for _, fname := range fnames {
		if fname.IsDir() {
			continue
		}

		b := strings.Split(fname.Name(), ".")
		if len(b) < len(a) {
			continue
		}
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				continue OuterFor
			}
		}

		return fname.Name()
	}

	return ""
}

// ServeHTTP maintains a cache and recompiles files that are dirty.
func (d DynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var inFileInfo, outFileInfo os.FileInfo
	reqURL := r.URL.Path

	// Return not found for directories, only files are permitted
	if reqURL[len(reqURL)-1] == '/' {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reqURL = strings.TrimPrefix(reqURL, "/")
	chunks := strings.Split(reqURL, "/")

	ln := len(chunks)
	if ln < 3 && ln > 0 && chunks[0] != "assets" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	typ := chunks[1]
	exes, ok := d.exes(typ)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fileInfo, err := d.getFileInfo(chunks)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("%v", fileInfo)

	// No compilers or minifiers for this asset, serve asset directly
	if len(exes.Compilers) == 0 && len(exes.Minifier.Cmd) == 0 {
		fileInfo.outPath = fileInfo.inPath
		goto ServeFile
	}

	inFileInfo, err = os.Stat(fileInfo.inPath)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		d.logf("failed to stat %s: %v", fileInfo.inPath, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outFileInfo, err = os.Stat(fileInfo.outPath)
	if err != nil && !os.IsNotExist(err) {
		d.logf("failed to stat %s: %v", fileInfo.outPath, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// If file exists, and if out file is newer, serve out file
	if err == nil && outFileInfo.ModTime().After(inFileInfo.ModTime()) {
		goto ServeFile
	}

	_, err = d.transform(typ, fileInfo.inPath)
	if err != nil {
		d.logf("failed to transform %s: %v", fileInfo.inPath, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

ServeFile:
	file, err := os.Open(fileInfo.outPath)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		d.logf("failed to open %s: %v", fileInfo.outPath, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, r.URL.Path, time.Now(), file)
	_ = file.Close()
}
