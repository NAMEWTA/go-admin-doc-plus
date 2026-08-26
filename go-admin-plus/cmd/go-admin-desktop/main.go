//go:build !bindings

package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"go-admin/common/global"
	desktophost "go-admin/internal/host/desktop"
	desktopplatform "go-admin/internal/platform/desktop"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "go-admin desktop failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	assets, err := adminAssets()
	if err != nil {
		return fmt.Errorf("load embedded Admin assets: %w", err)
	}
	dataRoot, err := desktopDataRoot()
	if err != nil {
		return err
	}
	host, err := desktophost.New(
		desktophost.Config{Assets: assets, Version: global.Version},
		desktophost.NewRuntimeBuilder(desktophost.RuntimeConfig{
			DataRoot: dataRoot,
			Name:     "go-admin-desktop",
			Mode:     "dev",
		}),
		runWindow,
	)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return host.Run(ctx)
}

func desktopDataRoot() (string, error) {
	return desktopplatform.ResolveDataRoot(os.Getenv("GO_ADMIN_DESKTOP_DATA_ROOT"))
}

func runWindow(parent context.Context, bridge *desktophost.Bridge, assets fs.FS) error {
	stopped := make(chan struct{})
	var stopOnce sync.Once
	defer stopOnce.Do(func() { close(stopped) })
	return wails.Run(&options.App{
		Title:     "Go Admin Plus",
		Width:     1360,
		Height:    860,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 255},
		OnStartup: func(wailsCtx context.Context) {
			startDesktopE2E(wailsCtx, bridge)
			go func() {
				select {
				case <-parent.Done():
					wailsRuntime.Quit(wailsCtx)
				case <-stopped:
				}
			}()
		},
		OnDomReady: func(wailsCtx context.Context) {
			if script := desktopE2EScript(); script != "" {
				wailsRuntime.WindowExecJS(wailsCtx, script)
			}
		},
		OnShutdown: func(context.Context) {
			stopOnce.Do(func() { close(stopped) })
		},
		Bind: []interface{}{bridge},
	})
}
