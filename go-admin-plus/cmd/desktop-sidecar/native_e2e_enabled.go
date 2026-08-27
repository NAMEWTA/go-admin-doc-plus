//go:build desktop_native_e2e

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go-admin/internal/modules/demo"
	"go-admin/internal/modules/iam/authorization"
)

type nativeE2EAction struct {
	Action string `json:"action"`
}

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
	if err == nil {
		ctx, cancel := context.WithTimeout(request.Context(), requestTimeout)
		defer cancel()
		err = runtime.applyNativeE2EAction(ctx, action)
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (runtime *sidecarRuntime) applyNativeE2EAction(ctx context.Context, action string) error {
	if runtime.database == nil || runtime.sessions == nil {
		return errors.New("desktop test control unavailable")
	}
	switch action {
	case "scope-self", "scope-all":
		scope := strings.TrimPrefix(action, "scope-")
		result, err := runtime.database.Bun().ExecContext(ctx, `UPDATE iam_roles SET data_scope = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, scope, "role-system-admin")
		if err != nil {
			return errors.New("desktop test control failed")
		}
		count, err := result.RowsAffected()
		if err != nil || count != 1 {
			return errors.New("desktop test control failed")
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
