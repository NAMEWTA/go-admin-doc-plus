//go:build bindings

package main

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"

	desktophost "go-admin/internal/host/desktop"
)

// Wails executes a bindings-tagged binary before frontend assets exist.
func main() {
	_ = wails.Run(&options.App{Bind: []interface{}{&desktophost.Bridge{}}})
}
