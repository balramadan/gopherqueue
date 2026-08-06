package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distEmbedFS embed.FS

// DistFS returns an fs.FS containing the embedded Vue Web Dashboard static assets.
func DistFS() (fs.FS, error) {
	return fs.Sub(distEmbedFS, "dist")
}
