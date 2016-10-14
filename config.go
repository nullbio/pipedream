package pipedream

import "github.com/BurntSushi/toml"

// Config for pipedream
type Pipedream struct {
	In  string `toml:"in"`
	Out string `toml:"out"`

	NoCompile  bool `toml:"no_compile"`
	NoMinify   bool `toml:"no_minify"`
	NoHash     bool `toml:"no_hash"`
	NoCompress bool `toml:"no_compress"`

	JS     Exes `toml:"js"`
	CSS    Exes `toml:"css"`
	Img    Exes `toml:"img"`
	Audio  Exes `toml:"audio"`
	Fonts  Exes `toml:"fonts"`
	Videos Exes `toml:"videos"`
}

type Exes struct {
	Compilers map[string]Command `toml:"compilers"`
	Minifier  Command            `toml:"minifier"`
}

// Command is an executable that can be run to produce a file
type Command struct {
	Cmd    string   `toml:"cmd"`
	Args   []string `toml:"args"`
	Stdout bool     `toml:"stdout"`
	Stdin  bool     `toml:"stdin"`
}

// New loads a configuration
func New(file string) (Pipedream, error) {
	var pipedream Pipedream
	_, err := toml.DecodeFile(file, &pipedream)

	return pipedream, err
}
