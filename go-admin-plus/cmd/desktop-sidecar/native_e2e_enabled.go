//go:build desktop_native_e2e

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

type nativeE2EAction struct {
	Action string `json:"action"`
}

var (
	errNativeE2EDependencies  = errors.New("desktop test control dependencies unavailable")
	errNativeE2EScopeSentinel = errors.New("desktop test control scope sentinel failed")
	errNativeE2EScopeUpdate   = errors.New("desktop test control scope update failed")
	errNativeE2EScopeRows     = errors.New("desktop test control scope rows failed")
)

func (runtime *sidecarRuntime) registerNativeE2EControl(mux *http.ServeMux) {
	mux.Handle("POST /__desktop/test-control", runtime.requireControl(http.HandlerFunc(runtime.nativeE2EControl)))
}

func decodeNativeE2EAction(request *http.Request) (string, error) {
	if request.Header.Get("Content-Type") != "application/json" {
		return "", errors.New("invalid desktop test control request")
	}
	decoder := json.NewDecoder(io.LimitReader(request.Body, 129))
	decoder.DisallowUnknownFields()
	var value nativeE2EAction
	if err := decoder.Decode(&value); err != nil {
		return "", errors.New("invalid desktop test control request")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return "", errors.New("invalid desktop test control request")
	}
	switch value.Action {
	case "scope-self", "scope-all", "permissions-off", "permissions-on", "session-revoke":
		return value.Action, nil
	default:
		return "", errors.New("invalid desktop test control request")
	}
}

func (runtime *sidecarRuntime) nativeE2EControl(writer http.ResponseWriter, request *http.Request) {
	action, err := decodeNativeE2EAction(request)
	if err != nil {
		writer.WriteHeader(http.StatusGatewayTimeout)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), requestTimeout)
	defer cancel()
	err = runtime.applyNativeE2EAction(ctx, action)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, errNativeE2EDependencies):
			status = http.StatusHTTPVersionNotSupported
		case errors.Is(err, errNativeE2EScopeSentinel):
			status = http.StatusNotImplemented
		case errors.Is(err, errNativeE2EScopeUpdate):
			status = http.StatusBadGateway
		case errors.Is(err, errNativeE2EScopeRows):
			status = http.StatusServiceUnavailable
		}
		writer.WriteHeader(status)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (runtime *sidecarRuntime) applyNativeE2EAction(ctx context.Context, action string) error {
	if runtime.database == nil || runtime.sessions == nil {
		return errNativeE2EDependencies
	}
	switch action {
	case "scope-self", "scope-all":
		scope := strings.TrimPrefix(action, "scope-")
		if action == "scope-self" {
			if _, err := runtime.database.Bun().ExecContext(ctx, `INSERT INTO demo_products
				(id, owner_account_id, sku, name, name_key, description, price_cents, status, revision, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT(id) DO NOTHING`, "00000000-0000-4000-8000-000000000001", "account-native-other",
				"E2E-FOREIGN", "Foreign product", "foreign product", "native scope sentinel", 1, "active", 1); err != nil {
				return errNativeE2EScopeSentinel
			}
		}
		result, err := runtime.database.Bun().ExecContext(ctx, `UPDATE iam_roles SET data_scope = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, scope, "role-system-admin")
		if err != nil {
			return errNativeE2EScopeUpdate
		}
		count, err := result.RowsAffected()
		if err != nil || count != 1 {
			return errNativeE2EScopeRows
		}
		return nil
	case "permissions-off":
		result, err := runtime.database.Bun().ExecContext(ctx, `DELETE FROM iam_role_permissions WHERE role_id = ? AND permission_code IN (?, ?, ?)`, "role-system-admin", demo.PermissionProductsRead, demo.PermissionProductsWrite, demo.PermissionProductsDelete)
		if err != nil {
			return errors.New("desktop test control failed")
		}
		count, err := result.RowsAffected()
		if err != nil || count != 3 {
			return errors.New("desktop test control failed")
		}
		return nil
	case "permissions-on":
		registry, err := authorization.NewCapabilityRegistry(runtime.database)
		if err != nil {
			return errors.New("desktop test control failed")
		}
		if err := demo.RegisterCapabilities(ctx, registry); err != nil {
			return errors.New("desktop test control failed")
		}
		return nil
	case "session-revoke":
		if err := runtime.sessions.RevokeAccount(ctx, "account-desktop-e2e"); err != nil {
			return errors.New("desktop test control failed")
		}
		return nil
	default:
		return errors.New("desktop test control unavailable")
	}
}
