package pipedream

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func (p Pipedream) transform(typ, file string) (string, error) {
	return "", nil
	// [0] = filename
	// [1] = fileext
	// [2..] = transformers
	chunks := strings.Split(file, ".")

	transforms := chunks[2:]
	if p.NoCompile || len(transforms) == 0 {
		// DON'T
	}

	if p.NoMinify {
		return "", nil
	}

	// minify
	return "", nil
}

func (p Pipedream) minify(typ, file string) (string, error) {
	return "", nil
}

func (p Pipedream) runCmd(c Command) error {
	return nil
}

type piper interface {
	ToPipe() (io.Reader, error)
	ToFile(string) (string, error)
}

type inputFile string

func (i inputFile) ToPipe() (io.Reader, error) {
	b, err := ioutil.ReadFile(string(i))
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func (i inputFile) ToFile(dstFile string) (string, error) {
	srcFile := string(i)
	if len(dstFile) != 0 && srcFile == dstFile {
		return srcFile, nil
	}

	src, err := os.Open(srcFile)
	if err != nil {
		return "", err
	}
	defer src.Close()

	return writeDstFile(dstFile, src)
}

type inputBuffer bytes.Buffer

func (i *inputBuffer) ToPipe() (io.Reader, error) {
	return (*bytes.Buffer)(i), nil
}

func (i *inputBuffer) ToFile(dstFile string) (string, error) {
	buf := (*bytes.Buffer)(i)
	return writeDstFile(dstFile, buf)
}

func writeDstFile(dstFile string, src io.Reader) (string, error) {
	var dst *os.File
	var err error

	if len(dstFile) != 0 {
		dst, err = os.Create(dstFile)
	} else {
		dst, err = ioutil.TempFile("", "pipedream")
		dstFile = dst.Name()
	}
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return dstFile, err
}
