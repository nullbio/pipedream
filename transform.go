package pipedream

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// transform takes a type of file (subfolder of assets directory: js, css, etc)
// and a full path to the file to transform
func (p Pipedream) transform(typ, file string) (string, error) {
	fn, err := mkFileNaming(p.In, p.Out, typ, file)
	if err != nil {
		return "", err
	}

	var out piper = inputFile(fn.AbsPath)
	out, err = p.runPipeline(typ, fn.Extensions, out)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(fn.AbsOutPath, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create output directory")
	}

	outputters := make([]io.Writer, 0, 2)
	finalOutput, err := os.Create(fn.Tmp)
	if err != nil {
		return "", errors.Wrap(err, "failed to create intermediate output file")
	}
	outputters = append(outputters, finalOutput)

	var fingerprint hash.Hash
	if !p.NoHash {
		fingerprint = md5.New()
		outputters = append(outputters, fingerprint)
	}

	var compressedOutput io.WriteCloser
	var compressor io.WriteCloser
	if !p.NoCompress {
		compressedOutput, err = os.Create(fn.Tmp + ".gz")
		if err != nil {
			return "", errors.Wrap(err, "failed to create intermediate output file")
		}

		compressor, err = gzip.NewWriterLevel(compressedOutput, gzip.BestSpeed)
		if err != nil {
			return "", errors.Wrap(err, "failed to create gzip writer")
		}

		outputters = append(outputters, compressor)
	}

	writer := io.MultiWriter(outputters...)

	var reader io.ReadCloser

	switch o := out.(type) {
	case *inputBuffer:
		reader = ioutil.NopCloser((*bytes.Buffer)(o))
	case inputFile:
		reader, err = os.Open(string(o))
		if err != nil {
			return "", errors.Wrap(err, "failed to open pipeline's source file")
		}
	default:
		panic("unreachable code")
	}

	if _, err = io.Copy(writer, reader); err != nil {
		return "", errors.Wrap(err, "failed to write to multiwriter")
	}

	if err = reader.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close the reader")
	}

	if err = finalOutput.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close final output")
	}

	if !p.NoHash {
		fn.Filename = fmt.Sprintf("%s-%x.%s", fn.Filename, fingerprint.Sum(nil), fn.Extension)
	} else {
		fn.Filename = fmt.Sprintf("%s.%s", fn.Filename, fn.Extension)
	}

	if !p.NoCompress {
		if err = compressor.Close(); err != nil {
			return "", errors.Wrap(err, "failed to flush compressor")
		}
		if err = compressedOutput.Close(); err != nil {
			return "", errors.Wrap(err, "failed to close compressed output")
		}

		if err = os.Rename(fn.Tmp+".gz", fn.Filename+".gz"); err != nil {
			return "", errors.Wrap(err, "failed to rename gzip'd output to final destination")
		}
	}

	if err = os.Rename(fn.Tmp, fn.Filename); err != nil {
		return "", errors.Wrap(err, "failed to rename to final destination")
	}

	return fn.Filename, nil
}

type fileNaming struct {
	AbsPath    string   // /tmp/assets/js/myapp/app.js.ts.erb
	RelPath    string   // js/myapp/app.js
	AbsOutPath string   // /tmp/assets/js/myapp/
	Filename   string   // app
	Extension  string   // js
	Extensions []string // [ts, erb]
	Tmp        string   // /tmp/assets/js/myapp/app-209320932030293.js
}

func mkFileNaming(inPath, outPath, typ, absPath string) (fileNaming, error) {
	fn := fileNaming{}
	fn.AbsPath = absPath

	filename := filepath.Base(absPath)
	fragments := strings.Split(filename, ".")
	if len(fragments) > 2 {
		fn.Extensions = fragments[2:]
	}
	if len(fragments) > 1 {
		fn.Extension = fragments[1]
	}
	fn.Filename = fragments[0]

	var err error
	fn.RelPath, err = filepath.Rel(filepath.Join(inPath, typ), absPath)
	if err != nil {
		return fn, errors.Wrap(err, "failed to find relative path")
	}

	fn.AbsOutPath = filepath.Join(outPath, filepath.Dir(fn.RelPath))

	randChunk := strconv.FormatInt(time.Now().UnixNano(), 10)
	fn.Tmp = filepath.Join(fn.AbsOutPath, fn.Filename+randChunk+fn.Extension)

	return fn, nil
}

func (p Pipedream) runPipeline(typ string, exts []string, out piper) (piper, error) {
	var err error

	pipeline := make([]transformer, 0)
	if !p.NoCompile {
		for i := len(exts) - 1; i >= 0; i-- {
			pipeline = append(pipeline, p.compiler(typ, exts[i]))
		}
	}

	if !p.NoMinify {
		pipeline = append(pipeline, p.minifier(typ))
	}

	for _, t := range pipeline {
		out, err = t(typ, out)
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute pipeline")
		}
	}

	return out, nil
}

type transformer func(typ string, in piper) (piper, error)

func (p Pipedream) compiler(typ string, extension string) transformer {
	var compiler Command
	var ok bool
	switch typ {
	case "js":
		compiler, ok = p.JS.Compilers[extension]
	case "css":
		compiler, ok = p.CSS.Compilers[extension]
	case "audio":
		compiler, ok = p.Audio.Compilers[extension]
	case "videos":
		compiler, ok = p.Videos.Compilers[extension]
	case "fonts":
		compiler, ok = p.Fonts.Compilers[extension]
	case "img":
		compiler, ok = p.Img.Compilers[extension]
	}

	if !ok {
		return nil
	}

	return mkTransformer(compiler)
}

func (p Pipedream) minifier(typ string) transformer {
	var minifier Command
	switch typ {
	case "js":
		minifier = p.JS.Minifier
	case "css":
		minifier = p.CSS.Minifier
	}

	if len(minifier.Cmd) == 0 {
		return nil
	}

	return mkTransformer(minifier)
}

func mkTransformer(c Command) transformer {
	return func(typ string, in piper) (piper, error) {
		out, err := runCmd(in, c)
		if err != nil {
			return nil, err
		}

		return out, nil
	}
}

func runCmd(in piper, c Command) (piper, error) {
	var err error
	var out piper
	var srcFile, dstFile string

	args := append([]string{}, c.Args...)
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "$infile":
			srcFile, err = in.ToFile()
			if err != nil {
				return nil, err
			}
			args[i] = srcFile
		case "$outfile":
			dstFile = randomTmpFileName()
			args[i] = dstFile
		}
	}

	cmd := exec.Command(c.Cmd, args...)
	fmt.Println("Command:", c.Cmd, "args:", args)
	if c.Stdin {
		reader, err := in.ToPipe()
		if err != nil {
			return nil, errors.Wrap(err, "failed to open command pipe")
		}

		cmd.Stdin = reader
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "cmd: %s args: %v\nstderr: %s\nstdout: %s\n",
			c.Cmd,
			args,
			stderr.Bytes(),
			stdout.Bytes(),
		)
	}

	if c.Stdout {
		out = (*inputBuffer)(stdout)
	} else {
		out = inputFile(dstFile)
	}
	return out, nil
}

func randomTmpFileName() string {
	return filepath.Join(
		os.TempDir(),
		"pipedream"+strconv.FormatInt(time.Now().UnixNano(), 10),
	)
}

type piper interface {
	ToPipe() (io.Reader, error)
	ToFile() (string, error)
}

type inputFile string

func (i inputFile) ToPipe() (io.Reader, error) {
	b, err := ioutil.ReadFile(string(i))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write inputfile to pipe")
	}

	return bytes.NewReader(b), nil
}

func (i inputFile) ToFile() (string, error) {
	return string(i), nil
}

type inputBuffer bytes.Buffer

func (i *inputBuffer) ToPipe() (io.Reader, error) {
	return (*bytes.Buffer)(i), nil
}

func (i *inputBuffer) ToFile() (string, error) {
	buf := (*bytes.Buffer)(i)
	return writeDstFile(buf)
}

func writeDstFile(src io.Reader) (string, error) {
	dst, err := ioutil.TempFile("", "pipedream")
	if err != nil {
		return "", errors.Wrapf(err, "failed to open temp file for writeDst")
	}
	defer dst.Close()

	dstFile := dst.Name()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", errors.Wrapf(err, "failed to copy to dstFile %s", dstFile)
	}
	return dstFile, nil
}
