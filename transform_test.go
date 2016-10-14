package pipedream

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestTransformEmpty(t *testing.T) {
	t.Parallel()

	if err := os.MkdirAll(filepath.Join(testTmp, "transforms", "js"), 0775); err != nil {
		t.Error(err)
	}

	inFile := filepath.Join(testTmp, "transforms", "css", "transform_empty.css")
	outFile := filepath.Join(testTmp, "tranforms_out", "css", "transform_empty.min.css")

	if err := ioutil.WriteFile(inFile, []byte(testTransformFile), 0664); err != nil {
		t.Fatal(err)
	}

	var p Pipedream

	p.JS.Minifier = Command{
		Cmd:    "cat",
		Args:   []string{"$infile"},
		Stdout: true,
	}

	p.transform("css", inFile)

	b, err := ioutil.ReadFile(outFile)
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
	filename, err := in.ToFile("")
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

func TestInputBufferToFile(t *testing.T) {
	t.Parallel()

	buf := (*inputBuffer)(bytes.NewBuffer([]byte(testTransformFile)))
	filename, err := buf.ToFile("")
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
