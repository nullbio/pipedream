package pipedream

import (
	"testing"

	"github.com/BurntSushi/toml"
)

var testConfig = `
in = "/in"
out = "/out"

[js.compilers.ts]
cmd = "ts"
args = ["--outFile", "$outfile", "$infile"]

[js.minifier]
cmd = "babel"
args = ["$infile"]
usestdout = true

[css.compilers.scss]
cmd = "node-sass"
args = ["$infile"]
usestdout = true
`

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cfg := &Pipedream{}

	_, err := toml.Decode(testConfig, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.In != "/in" {
		t.Error("in was wrong:", cfg.In)
	}
	if cfg.Out != "/out" {
		t.Error("out was wrong:", cfg.Out)
	}

	ts, ok := cfg.JS.Compilers["ts"]
	if !ok {
		t.Error("there should be a ts compiler")
	}
	min := cfg.JS.Minifier

	if ts.Cmd != "ts" {
		t.Error("command was wrong:", ts.Cmd)
	}
	for i, a := range []string{"--outFile", "$outfile", "$infile"} {
		if a != ts.Args[i] {
			t.Errorf("argument %d was wrong:", ts.Args[i])
		}
	}

	if min.Cmd != "babel" {
		t.Error("command was wrong:", ts.Cmd)
	}
	for i, a := range []string{"$infile"} {
		if a != min.Args[i] {
			t.Errorf("argument %d was wrong:", min.Args[i])
		}
	}
}
