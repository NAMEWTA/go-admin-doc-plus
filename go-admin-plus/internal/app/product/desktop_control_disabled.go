//go:build !desktop_native_e2e

package product

import (
	desktophost "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/host/desktop"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/session"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func desktopPrivateRoute(*database.Database, *session.Service) *desktophost.PrivateRoute { return nil }
