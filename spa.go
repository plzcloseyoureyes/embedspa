package embedspa

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
)

type EtagFunc func(filename string, fileStat fs.FileInfo) string
type EmbedSPAHandler struct {
	Fsys           fs.FS
	HttpFS         http.FileSystem
	FileServer     http.Handler
	IndexPath      string
	UrlStripPrefix string
	customETAG     EtagFunc
}

func NewEmbedSPAHandler(Fsys fs.FS) *EmbedSPAHandler {
	h := &EmbedSPAHandler{
		Fsys: Fsys,
	}
	h.HttpFS = http.FS(h.Fsys)
	h.FileServer = http.FileServer(h.HttpFS)
	return h
}
func (h *EmbedSPAHandler) SetCustomETAG(CustomETAG EtagFunc) *EmbedSPAHandler {
	h.customETAG = CustomETAG
	return h
}
func (h *EmbedSPAHandler) SetIndexPath(IndexPath string) *EmbedSPAHandler {
	h.IndexPath = IndexPath
	return h
}
func (h *EmbedSPAHandler) StripPrefixURL(UrlStripPrefix string) *EmbedSPAHandler {
	h.UrlStripPrefix = UrlStripPrefix
	return h
}

// ServeHTTP inspects the URL path to locate a file within the embedded static dir
// If a file is found, it will be served. If not, the file located at the index path will be served.
func (h *EmbedSPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	filename := path.Clean(r.RequestURI)
	stripFilename := strings.TrimPrefix(filename, h.UrlStripPrefix)
	stripFilename = strings.TrimPrefix(stripFilename, "/")
	if stripFilename == "" {
		stripFilename = h.IndexPath
	}
	var etag string
	file, err := h.Fsys.Open(stripFilename)
	notFound := false
	if err != nil {
    // File not found, will render index.html or IndexPath.
		notFound = true
		file, _ = h.Fsys.Open(h.IndexPath)
	}
	fileStat, err := file.Stat()
	if h.customETAG != nil {
		etag = h.customETAG(filename, fileStat) // Custom ETAG func wrap
	} else {
		if err != nil {
			// Do something
		}
		fileSize := fileStat.Size()
		fileModTime := fileStat.ModTime().UTC().String()

		// Create a hash of the combined string
		hash := md5.New()
		hash.Write([]byte(fmt.Sprintf("%d-%s", fileSize, fileModTime)))
		etag = hex.EncodeToString(hash.Sum(nil))
	}
	if headers.Get(`If-None-Match`) == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Cache-Control", "max-age=604800")
	w.Header().Set("ETag", etag)
	expires := time.Now().Add(7 * 24 * time.Hour).Format("Mon, 02 Jan 2006 15:04:05 GMT")
	w.Header().Set("Expires", expires)
	w.Header().Set("Vary", "Accept-Encoding")
	if notFound {
		http.ServeFileFS(w, r, h.Fsys, h.IndexPath)
	} else {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, h.UrlStripPrefix)
		r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, h.UrlStripPrefix)
		h.FileServer.ServeHTTP(w, r)
	}
}

func (h *EmbedSPAHandler) GIN(ctx *gin.Context) {
	h.ServeHTTP(ctx.Writer, ctx.Request)
}
