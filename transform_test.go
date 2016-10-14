package pipedream

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

var testTransformFile = `
dreaming of pipes
`

func TestTransform(t *testing.T) {
	/*t.Parallel()

	if err := os.MkdirAll(filepath.Join(testTmp, "transforms", "js"), 0775); err != nil {
		t.Error(err)
	}

	inFile := filepath.Join(testTmp, "transforms", "javascripts", "transform_file.js.tee.cat")
	outFile := filepath.Join(testTmp, "tranforms_out", "javascripts", "transform_file.min.js")

	if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}

	var p Pipedream

	p.JS.Compilers["cat"] = Command{
		Cmd:    "cat",
		Stdin:  true,
		Stdout: true,
	}

	p.JS.Compilers["tee"] = Command{
		Cmd:   "tee",
		Args:  []string{"$outfile"},
		Stdin: true,
	}

	p.JS.Minifier = Command{
		Cmd:    "cat",
		Args:   []string{"$infile"},
		Stdout: true,
	}

	p.transform("js", inFile)

	b, err := ioutil.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != testTransformFile {
		t.Errorf("file was wrong:\n%x\n%s", b, b)
	}*/
}

func TestMkFileNaming(t *testing.T) {
	t.Parallel()

	inPath := "/in_stuff/assets.folder"
	outPath := "/out_stuff/assets.folder"
	typ := "css"
	absPath := "/in_stuff/assets.folder/css/my.things/file.thing.css.scss.erb"
	absOutPath := "/out_stuff/assets.folder/css/my.things"
	filename := "file.thing"
	extension := "css"
	extensions := []string{"scss", "erb"}
	outfile := regexp.MustCompile(`^(?i)/out_stuff/assets.folder/css/my.things/file\.thing-[0-9]+\.css$`)

	r, err := mkFileNaming(
		inPath,
		outPath,
		typ,
		absPath,
	)

	if err != nil {
		t.Error(err)
	}

	if r.AbsPath != absPath {
		t.Errorf("AbsPatch mismatch\nwant: %s\ngot: %s", absPath, r.AbsPath)
	}

	if r.AbsOutPath != absOutPath {
		t.Errorf("AbsOutPath mismatch\nwant: %s\ngot: %s", absOutPath, r.AbsOutPath)
	}

	if r.Filename != filename {
		t.Errorf("Filename mismatch\nwant: %s\ngot: %s", filename, r.Filename)
	}

	if r.Extension != extension {
		t.Errorf("Extension mismatch\nwant: %s\ngot: %s", extension, r.Extension)
	}

	if !outfile.MatchString(r.OutFile) {
		t.Errorf("OutFile mismatch, got: %s", r.OutFile)
	}

	if !reflect.DeepEqual(extensions, r.Extensions) {
		t.Errorf("Extensions mismatch\nwant: %v\ngot: %v", extensions, r.Extensions)
	}
}

func TestTransformMinifyOnly(t *testing.T) {
	t.Parallel()

	if err := os.MkdirAll(filepath.Join(testTmp, "transforms", "js"), 0775); err != nil {
		t.Error(err)
	}

	inFile := filepath.Join(testTmp, "transforms", "css", "transform_empty.css")
	outFile := filepath.Join(testTmp, "transforms_out", "css", "transform_empty.css")
	outFileRgx := regexp.MustCompile(`^` + testTmp + `/transforms_out/css/transform_empty-[0-9a-z]+\.css$`)

	if err := os.MkdirAll(filepath.Dir(inFile), 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(outFile), 0775); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}

	var p Pipedream
	p.In = filepath.Join(testTmp, "transforms")
	p.Out = filepath.Join(testTmp, "transforms_out")
	p.NoCompress = true

	p.CSS.Minifier = Command{
		Cmd:    "cat",
		Args:   []string{"$infile"},
		Stdout: true,
	}

	out, err := p.transform("css", inFile)
	if err != nil {
		t.Fatal(err)
	}
	if !outFileRgx.MatchString(out) {
		t.Errorf("output file path did not match regexp:\n%s\n%s", outFileRgx.String(), out)
	}

	b, err := ioutil.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != testTransformFile {
		t.Errorf("file was wrong:\n%x\n%s", b, b)
	}
}

func TestInputFileToPipe(t *testing.T) {
	t.Parallel()

	inFile := filepath.Join(testTmp, "inputfiletopipe")
	if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}

	in := inputFile(inFile)
	out, err := in.ToPipe()
	if err != nil {
		t.Fatal(err)
	}

	if str, err := ioutil.ReadAll(out); err != nil {
		t.Error(err)
	} else if string(str) != testTransformFile {
		t.Error("file output was wrong:\n", string(str))
	}
}

func TestInputFileToFile(t *testing.T) {
	t.Parallel()

	inFile := filepath.Join(testTmp, "inputfiletofile")
	if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}

	in := inputFile(inFile)
	filename, err := in.ToFile()
	if err != nil {
		t.Fatal(err)
	}

	if filename != inFile {
		t.Error("filename should not change:", filename)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != testTransformFile {
		t.Error("file output was wrong:\n", string(b))
	}
}

func TestInputBufferToFile(t *testing.T) {
	t.Parallel()

	buf := (*inputBuffer)(bytes.NewBuffer([]byte(testTransformFile)))
	filename, err := buf.ToFile()
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != testTransformFile {
		t.Error("file output was wrong:\n", string(b))
	}
}

func TestInputBufferToPipe(t *testing.T) {
	t.Parallel()

	buf := (*inputBuffer)(bytes.NewBuffer([]byte(testTransformFile)))
	pipe, err := buf.ToPipe()
	if err != nil {
		t.Fatal(err)
	}

	if str, err := ioutil.ReadAll(pipe); err != nil {
		t.Error(err)
	} else if string(str) != testTransformFile {
		t.Error("file output was wrong:\n", string(str))
	}
}
