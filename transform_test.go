package pipedream

import (
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
