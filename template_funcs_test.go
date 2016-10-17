package pipedream

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatePaths(t *testing.T) {
	t.Parallel()

	var p Pipedream
	p.NoHash = true

	if got := p.JSPath("test/javascript.js"); got != "/assets/js/test/javascript.js" {
		t.Error("path was wrong:", got)
	}
	if got := p.CSSPath("stylesheet.css"); got != "/assets/css/stylesheet.css" {
		t.Error("path was wrong:", got)
	}
	if got := p.ImgPath("image.png"); got != "/assets/img/image.png" {
		t.Error("path was wrong:", got)
	}
	if got := p.VideoPath("video.mp4"); got != "/assets/videos/video.mp4" {
		t.Error("path was wrong:", got)
	}
	if got := p.AudioPath("audio.ogg"); got != "/assets/audio/audio.ogg" {
		t.Error("path was wrong:", got)
	}
	if got := p.FontPath("font.ttf"); got != "/assets/fonts/font.ttf" {
		t.Error("path was wrong:", got)
	}
}

func TestTemplatePathsCDN(t *testing.T) {
	t.Parallel()

	var p Pipedream
	p.NoHash = true
	p.CDNURL = "https://cdn.com/pubhelpers"

	if got := p.JSPath("test/javascript.js"); got != "https://cdn.com/pubhelpers/assets/js/test/javascript.js" {
		t.Error("path was wrong:", got)
	}
	if got := p.CSSPath("stylesheet.css"); got != "https://cdn.com/pubhelpers/assets/css/stylesheet.css" {
		t.Error("path was wrong:", got)
	}
	if got := p.ImgPath("image.png"); got != "https://cdn.com/pubhelpers/assets/img/image.png" {
		t.Error("path was wrong:", got)
	}
	if got := p.VideoPath("video.mp4"); got != "https://cdn.com/pubhelpers/assets/videos/video.mp4" {
		t.Error("path was wrong:", got)
	}
	if got := p.AudioPath("audio.ogg"); got != "https://cdn.com/pubhelpers/assets/audio/audio.ogg" {
		t.Error("path was wrong:", got)
	}
	if got := p.FontPath("font.ttf"); got != "https://cdn.com/pubhelpers/assets/fonts/font.ttf" {
		t.Error("path was wrong:", got)
	}
}

func TestTemplatePathsFingerprint(t *testing.T) {
	t.Parallel()

	var p Pipedream
	p.Out = filepath.Join(testTmp, "fingerprintpaths")

	assetDir := filepath.Join(p.Out, "assets")

	if err := os.MkdirAll(assetDir, 0775); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(assetDir, "manifest.json"), []byte(testManifest), 0664); err != nil {
		t.Fatal(err)
	}

	if err := p.LoadManifest(); err != nil {
		t.Fatal(err)
	}

	if got := p.JSPath("test/javascript.js"); got != "/assets/js/test/javascript-abc.js" {
		t.Error("path was wrong:", got)
	}
	if got := p.CSSPath("stylesheet.css"); got != "/assets/css/stylesheet-def.css" {
		t.Error("path was wrong:", got)
	}
	if got := p.ImgPath("image.png"); got != "/assets/img/image-ghi.png" {
		t.Error("path was wrong:", got)
	}

	p.CDNURL = "https://assets.com/hack"

	if got := p.VideoPath("video.mp4"); got != "https://assets.com/hack/assets/videos/video-jkl.mp4" {
		t.Error("path was wrong:", got)
	}
	if got := p.AudioPath("audio.ogg"); got != "https://assets.com/hack/assets/audio/audio-mno.ogg" {
		t.Error("path was wrong:", got)
	}
	if got := p.FontPath("font.ttf"); got != "https://assets.com/hack/assets/fonts/font-pqr.ttf" {
		t.Error("path was wrong:", got)
	}
}

var testManifest = `
{
	"files": {
		"/assets/js/test/javascript-abc.js": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "abc"
		},
		"/assets/css/stylesheet-def.css": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "def"
		},
		"/assets/img/image-ghi.png": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "png"
		},
		"/assets/videos/video-jkl.mp4": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "mp4"
		},
		"/assets/audio/audio-mno.ogg": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "ogg"
		},
		"/assets/fonts/font-pqr.ttf": {
			"mtime": "2016-06-18T14:18:44-07:00",
			"size": 158986,
			"digest": "ttf"
		}
	},
	"assets": {
		"js/test/javascript.js": "/assets/js/test/javascript-abc.js",
		"css/stylesheet.css": "/assets/css/stylesheet-def.css",
		"img/image.png": "/assets/img/image-ghi.png",
		"videos/video.mp4": "/assets/videos/video-jkl.mp4",
		"audio/audio.ogg": "/assets/audio/audio-mno.ogg",
		"fonts/font.ttf": "/assets/fonts/font-pqr.ttf"
	}
}`
