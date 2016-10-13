package pipedream

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func (p Pipedream) transform(typ, file string) (string, error) {
	// [0] = filename
	// [1] = fileext
	// [2..] = transformers
	chunks := strings.Split(file, ".")

	transforms = chunks[2:]
	if p.NoCompile || len(transforms) == 0 {
		// DON'T
	}

	if p.NoMinify {
		return nil
	}

	// minify
}

type inputer interface {
	ToPipe() (io.Reader, error)
	ToFile(string) (string, error)
}

type inputFile string

func (i inputFile) ToPipe() (io.Reader, error) {
	b, err := ioutil.ReadFile(i)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), err
}

func (i inputFile) ToFile(dest string) (string, error) {
	if len(dest) == 0 || i == dest {
		return i, err
	}

	src, err := os.Open(i)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

type inputBuffer bytes.Buffer

func (i *inputBuffer) ToPipe() (io.Reader, error) {
	return *bytes.Buffer(i), nil
}

func (i *inputBuffer) ToFile(dest string) (string, error) {
	var f *os.File
	var err error

	if len(dest) != 0 {
		f, err := os.Create(dest)
	} else {
		f, err := ioutil.TempFile("", "pipedream")
	}
	if err != nil {
		return "", err
	}

	buf := *bytes.Buffer(i)
	if _, err = io.Copy(f, buf); err != nil {
		return "", err
	}

	fname := f.Name()
	return fname, f.Close()
}

func (p Pipedream) minify(typ, file string) (string, error) {
}

func (p Pipedream) runCmd(c Command) error {
}
