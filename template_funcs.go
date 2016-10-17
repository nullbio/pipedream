package pipedream

import "fmt"

// JSPath returns the path for a given js asset.
func (p Pipedream) JSPath(file string) string {
	return p.lookupPath(typeJS, file)
}

// CSSPath returns the path for a given css asset.
func (p Pipedream) CSSPath(file string) string {
	return p.lookupPath(typeCSS, file)
}

// ImgPath returns the path for a given img asset.
func (p Pipedream) ImgPath(file string) string {
	return p.lookupPath(typeImg, file)
}

// VideoPath returns the path for a given video asset.
func (p Pipedream) VideoPath(file string) string {
	return p.lookupPath(typeVideos, file)
}

// AudioPath returns the path for a given audio asset.
func (p Pipedream) AudioPath(file string) string {
	return p.lookupPath(typeAudio, file)
}

// FontPath returns the path for a given font asset.
func (p Pipedream) FontPath(file string) string {
	return p.lookupPath(typeFonts, file)
}

func (p Pipedream) lookupPath(typ, file string) string {
	if p.NoHash {
		return fmt.Sprintf("%s/assets/%s/%s", p.CDNURL, typ, file)
	}

	key := fmt.Sprintf("%s/%s", typ, file)
	asset, ok := p.Manifest.Assets[key]

	if !ok {
		panic(fmt.Sprintf("asset %s requested but was not in manifest, did you rememeber to precompile assets?", key))
	}

	return asset
}
