package pipedream

import "github.com/BurntSushi/toml"

// Pipedream is the config for pipedream
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

// Exes holds the compilers and minifiers for each file type
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

// exes returns the exe for typ
func (p Pipedream) exes(typ string) (exes Exes, ok bool) {
	switch typ {
	case "js":
		return p.JS, true
	case "css":
		return p.CSS, true
	case "img":
		return p.Img, true
	case "audio":
		return p.Audio, true
	case "fonts":
		return p.Fonts, true
	case "videos":
		return p.Videos, true
	default:
		return exes, false
	}
}
