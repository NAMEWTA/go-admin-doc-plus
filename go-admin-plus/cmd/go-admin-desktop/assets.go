package main

import (
	"embed"
	"io/fs"
)

// Wails copies the shared frontend dist here for the duration of a build.
//
//go:embed all:frontend/dist
var embeddedAssets embed.FS

func adminAssets() (fs.FS, error) {
	return fs.Sub(embeddedAssets, "frontend/dist")
}
