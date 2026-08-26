//go:build !desktop_e2e

package main

import (
	"context"

	desktophost "go-admin/internal/host/desktop"
)

func startDesktopE2E(context.Context, *desktophost.Bridge) {}

func desktopE2EScript() string { return "" }
