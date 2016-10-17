package pipedream

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler serves static assets from the out path.
func (p *Pipedream) StaticHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.Replace(strings.Replace(r.URL.Path, "..", "", -1), os.PathSeparator+".", "", -1)

	fileInfo, ok := p.Manifest.Files[urlPath]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	acceptEncodings := r.Header.Get("Accept-Encoding")
	encodings := strings.Split(acceptEncodings)
	useGzip := false
	for _, e := range encodings {
		if strings.HasPrefix(e, "gzip") {
			useGzip = true
			break
		}
	}

	urlPath += ".gz"
	fileLoc := filepath.Join(p.Out, urlPath)
	file, err := os.Open(fileLoc)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Md5", fileInfo.Digest)
	w.Header().Set("ETag", `"`+fileInfo.Digest+`"`)
	http.ServeContent(w, r, urlPath, fileInfo.MTime)

	_ = file.Close()
}

// DynamicHandler maintains a cache and recompiles files that are dirty.
func (p *Pipedream) DynamicHandler(w http.ResponseWriter, r *http.Request) {
}
