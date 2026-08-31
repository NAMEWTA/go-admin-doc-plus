package administration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/administration/transport"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func (s *HTTPServer) GetIamCapabilityManifest(ctx context.Context, _ transport.GetIamCapabilityManifestRequestObject) (transport.GetIamCapabilityManifestResponseObject, error) {
	value, err := s.service.Manifest(ctx, requestHTTP(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.GetIamCapabilityManifest403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.GetIamCapabilityManifest500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.GetIamCapabilityManifest200JSONResponse{Body: transportManifest(value), Headers: transport.GetIamCapabilityManifest200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) ListIamRoles(ctx context.Context, _ transport.ListIamRolesRequestObject) (transport.ListIamRolesResponseObject, error) {
	values, err := s.service.ListRoles(ctx, requestHTTP(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.ListIamRoles403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListIamRoles500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	if err := s.projectEffectiveRoleScopes(ctx, values); err != nil {
		return transport.ListIamRoles500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	body := make([]transport.Role, 0, len(values))
	for _, value := range values {
		body = append(body, transportRole(value))
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListIamRoles200JSONResponse{Body: body, Headers: transport.ListIamRoles200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) CreateIamRole(ctx context.Context, request transport.CreateIamRoleRequestObject) (transport.CreateIamRoleResponseObject, error) {
	if request.Body == nil {
		return transport.CreateIamRole400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.CreateRole(ctx, requestHTTP(ctx).actorID, request.Body.Key, request.Body.Name, authorization.Scope(request.Body.DataScope))
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.CreateIamRole400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.CreateIamRole403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.CreateIamRole409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.CreateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	if s.dataScopes != nil {
		err = s.dataScopes.SetRoleDataScope(ctx, requestHTTP(ctx).actorID, value.ID, RoleDataScope{Scope: value.Scope})
		if err != nil {
			if cleanupErr := s.removeUninitializedRole(context.WithoutCancel(ctx), value.ID); cleanupErr != nil {
				return transport.CreateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
			}
			switch {
			case errors.Is(err, ErrDenied):
				return transport.CreateIamRole403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
			case errors.Is(err, ErrConflict):
				return transport.CreateIamRole409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
			default:
				return transport.CreateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
			}
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.CreateIamRole201JSONResponse{Body: transportRole(value), Headers: transport.CreateIamRole201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) projectEffectiveRoleScopes(ctx context.Context, values []Role) error {
	if s.dataScopes == nil || len(values) == 0 {
		return nil
	}
	arguments := make([]any, 0, len(values))
	index := make(map[string]int, len(values))
	for position := range values {
		arguments = append(arguments, values[position].ID)
		index[values[position].ID] = position
	}
	return s.dataScopes.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT role_id, scope FROM iam_role_data_scopes WHERE role_id IN (`+placeholders(len(values))+`)`, arguments...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var roleID string
			var scope authorization.Scope
			if err := rows.Scan(&roleID, &scope); err != nil {
				return err
			}
			position, ok := index[roleID]
			if !ok || !validDataScope(scope) {
				return fmt.Errorf("invalid role data scope projection")
			}
			values[position].Scope = scope
		}
		return rows.Err()
	})
}

func (s *HTTPServer) removeUninitializedRole(ctx context.Context, roleID string) error {
	return s.dataScopes.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM iam_roles WHERE id = ?`, roleID)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("created role cleanup affected %d rows", count)
		}
		return nil
	})
}

func (s *HTTPServer) UpdateIamRole(ctx context.Context, request transport.UpdateIamRoleRequestObject) (transport.UpdateIamRoleResponseObject, error) {
	if request.Body == nil {
		return transport.UpdateIamRole400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	err := s.service.UpdateRole(ctx, requestHTTP(ctx).actorID, Role{ID: request.RoleId, Key: request.Body.Key, Name: request.Body.Name, Scope: authorization.Scope(request.Body.DataScope), Enabled: request.Body.Enabled})
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.UpdateIamRole400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.UpdateIamRole403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return transport.UpdateIamRole404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.UpdateIamRole409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.UpdateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	if s.dataScopes != nil {
		syncScope, syncErr := s.shouldSyncBaseRoleScope(ctx, request.RoleId)
		if syncErr != nil {
			return transport.UpdateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
		if syncScope {
			syncErr = s.dataScopes.SetRoleDataScope(ctx, requestHTTP(ctx).actorID, request.RoleId, RoleDataScope{Scope: authorization.Scope(request.Body.DataScope)})
			switch {
			case syncErr == nil:
			case errors.Is(syncErr, ErrDenied):
				return transport.UpdateIamRole403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
			case errors.Is(syncErr, ErrNotFound):
				return transport.UpdateIamRole404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
			case errors.Is(syncErr, ErrConflict):
				return transport.UpdateIamRole409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
			default:
				return transport.UpdateIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
			}
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.UpdateIamRole204Response{Headers: transport.UpdateIamRole204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) shouldSyncBaseRoleScope(ctx context.Context, roleID string) (bool, error) {
	var scope authorization.Scope
	err := s.dataScopes.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT scope FROM iam_role_data_scopes WHERE role_id = ?`, roleID).Scan(&scope)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	switch scope {
	case authorization.ScopeAll, authorization.ScopeSelf:
		return true, nil
	case authorization.ScopeOrganization, authorization.ScopeOrganizationTree, authorization.ScopeCustom:
		return false, nil
	default:
		return false, fmt.Errorf("invalid role data scope projection")
	}
}

func (s *HTTPServer) DeleteIamRole(ctx context.Context, request transport.DeleteIamRoleRequestObject) (transport.DeleteIamRoleResponseObject, error) {
	err := s.service.DeleteRole(ctx, requestHTTP(ctx).actorID, request.RoleId)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.DeleteIamRole403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return transport.DeleteIamRole404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.DeleteIamRole409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.DeleteIamRole500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.DeleteIamRole204Response{Headers: transport.DeleteIamRole204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) SetIamRoleGrants(ctx context.Context, request transport.SetIamRoleGrantsRequestObject) (transport.SetIamRoleGrantsResponseObject, error) {
	if request.Body == nil {
		return transport.SetIamRoleGrants400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	err := s.service.SetRoleGrants(ctx, requestHTTP(ctx).actorID, request.RoleId, request.Body.PermissionCodes, request.Body.MenuIds)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.SetIamRoleGrants400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.SetIamRoleGrants403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return transport.SetIamRoleGrants404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.SetIamRoleGrants409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.SetIamRoleGrants500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.SetIamRoleGrants204Response{Headers: transport.SetIamRoleGrants204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) ListIamMenus(ctx context.Context, _ transport.ListIamMenusRequestObject) (transport.ListIamMenusResponseObject, error) {
	values, err := s.service.ListMenus(ctx, requestHTTP(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.ListIamMenus403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListIamMenus500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	body := make([]transport.Menu, 0, len(values))
	for _, value := range values {
		body = append(body, transportMenu(value))
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListIamMenus200JSONResponse{Body: body, Headers: transport.ListIamMenus200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) CreateIamMenu(ctx context.Context, request transport.CreateIamMenuRequestObject) (transport.CreateIamMenuResponseObject, error) {
	if request.Body == nil {
		return transport.CreateIamMenu400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.CreateMenu(ctx, requestHTTP(ctx).actorID, Menu{Key: request.Body.Key, Label: request.Body.Label, Path: request.Body.Path, PermissionCode: request.Body.PermissionCode, SortOrder: request.Body.SortOrder})
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.CreateIamMenu400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.CreateIamMenu403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.CreateIamMenu409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.CreateIamMenu500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.CreateIamMenu201JSONResponse{Body: transportMenu(value), Headers: transport.CreateIamMenu201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) UpdateIamMenu(ctx context.Context, request transport.UpdateIamMenuRequestObject) (transport.UpdateIamMenuResponseObject, error) {
	if request.Body == nil {
		return transport.UpdateIamMenu400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	err := s.service.UpdateMenu(ctx, requestHTTP(ctx).actorID, Menu{ID: request.MenuId, Key: request.Body.Key, Label: request.Body.Label, Path: request.Body.Path, PermissionCode: request.Body.PermissionCode, SortOrder: request.Body.SortOrder})
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.UpdateIamMenu400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.UpdateIamMenu403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return transport.UpdateIamMenu404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.UpdateIamMenu409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.UpdateIamMenu500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.UpdateIamMenu204Response{Headers: transport.UpdateIamMenu204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) DeleteIamMenu(ctx context.Context, request transport.DeleteIamMenuRequestObject) (transport.DeleteIamMenuResponseObject, error) {
	err := s.service.DeleteMenu(ctx, requestHTTP(ctx).actorID, request.MenuId)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.DeleteIamMenu403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return transport.DeleteIamMenu404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		}
		if errors.Is(err, ErrConflict) {
			return transport.DeleteIamMenu409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		}
		return transport.DeleteIamMenu500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.DeleteIamMenu204Response{Headers: transport.DeleteIamMenu204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) ListIamPermissions(ctx context.Context, _ transport.ListIamPermissionsRequestObject) (transport.ListIamPermissionsResponseObject, error) {
	values, err := s.service.ListPermissions(ctx, requestHTTP(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.ListIamPermissions403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListIamPermissions500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	body := make([]transport.Permission, 0, len(values))
	for _, value := range values {
		body = append(body, transport.Permission{Code: value.Code, Name: value.Name})
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListIamPermissions200JSONResponse{Body: body, Headers: transport.ListIamPermissions200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
