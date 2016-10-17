package pipedream

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	typeJS     = "js"
	typeCSS    = "css"
	typeImg    = "img"
	typeAudio  = "audio"
	typeVideos = "videos"
	typeFonts  = "fonts"
)

// Pipedream is the config for pipedream
type Pipedream struct {
	In  string `toml:"in"`
	Out string `toml:"out"`

	CDNURL string `toml:"cdn_url"`

	NoCompile  bool `toml:"no_compile"`
	NoMinify   bool `toml:"no_minify"`
	NoHash     bool `toml:"no_hash"`
	NoCompress bool `toml:"no_compress"`

	JS     Exes `toml:"js"`
	CSS    Exes `toml:"css"`
	Img    Exes `toml:"img"`
	Audio  Exes `toml:"audio"`
	Videos Exes `toml:"videos"`
	Fonts  Exes `toml:"fonts"`

	Manifest Manifest `toml:"-"`
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

// Manifest for compiled assets
type Manifest struct {
	Assets map[string]string `json:"assets"`
}

// New loads a configuration
func New(file string) (Pipedream, error) {
	var pipedream Pipedream
	_, err := toml.DecodeFile(file, &pipedream)

	pipedream.CDNURL = strings.TrimRight(pipedream.CDNURL, "/")

	return pipedream, err
}

// LoadManifest loads the manifest in p.OutPath/assets/manifest.json
func (p *Pipedream) LoadManifest() error {
	b, err := ioutil.ReadFile(filepath.Join(p.Out, "assets", "manifest.json"))
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, &p.Manifest); err != nil {
		return err
	}

	return nil
}

// exes returns the exe for typ
func (p *Pipedream) exes(typ string) (exes Exes, ok bool) {
	switch typ {
	case typeJS:
		return p.JS, true
	case typeCSS:
		return p.CSS, true
	case typeImg:
		return p.Img, true
	case typeAudio:
		return p.Audio, true
	case typeVideos:
		return p.Videos, true
	case typeFonts:
		return p.Fonts, true
	default:
		return exes, false
	}
}
