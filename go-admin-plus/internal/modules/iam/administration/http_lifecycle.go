package administration

import (
	"context"
	"errors"

	"github.com/google/uuid"

	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/administration/transport"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

func (s *HTTPServer) SetIamUserOrganization(ctx context.Context, request transport.SetIamUserOrganizationRequestObject) (transport.SetIamUserOrganizationResponseObject, error) {
	if request.Body == nil {
		return transport.SetIamUserOrganization400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	if s.dataScopes == nil {
		return transport.SetIamUserOrganization500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	positions := make([]string, len(request.Body.PositionIds))
	for index, id := range request.Body.PositionIds {
		positions[index] = id
	}
	var primary *string
	if request.Body.PrimaryDepartmentId != nil {
		value := string(*request.Body.PrimaryDepartmentId)
		primary = &value
	}
	err := s.dataScopes.SetAccountOrganization(ctx, requestHTTP(ctx).actorID, request.UserId, AccountOrganization{
		PrimaryDepartmentID: primary,
		PositionIDs:         positions,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.SetIamUserOrganization400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.SetIamUserOrganization403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.SetIamUserOrganization404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.SetIamUserOrganization409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.SetIamUserOrganization500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.SetIamUserOrganization204Response{Headers: transport.SetIamUserOrganization204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) SetIamRoleDataScope(ctx context.Context, request transport.SetIamRoleDataScopeRequestObject) (transport.SetIamRoleDataScopeResponseObject, error) {
	if request.Body == nil {
		return transport.SetIamRoleDataScope400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	if s.dataScopes == nil {
		return transport.SetIamRoleDataScope500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	departments := make([]string, len(request.Body.DepartmentIds))
	for index, id := range request.Body.DepartmentIds {
		departments[index] = id
	}
	err := s.dataScopes.SetRoleDataScope(ctx, requestHTTP(ctx).actorID, request.RoleId, RoleDataScope{
		Scope:         authorization.Scope(request.Body.Scope),
		DepartmentIDs: departments,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.SetIamRoleDataScope400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.SetIamRoleDataScope403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.SetIamRoleDataScope404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.SetIamRoleDataScope409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.SetIamRoleDataScope500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.SetIamRoleDataScope204Response{Headers: transport.SetIamRoleDataScope204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) StartIamUserDeletion(ctx context.Context, request transport.StartIamUserDeletionRequestObject) (transport.StartIamUserDeletionResponseObject, error) {
	if request.Body == nil {
		return transport.StartIamUserDeletion400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	if s.deletions == nil {
		return transport.StartIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	target := ""
	if request.Body.TransferTargetId != nil {
		target = string(*request.Body.TransferTargetId)
	}
	purgeConfirmed := request.Body.PurgeConfirmed != nil && *request.Body.PurgeConfirmed
	value, err := s.deletions.StartDeletion(ctx, requestHTTP(ctx).actorID, StartDeletion{
		AccountID:        request.UserId,
		Strategy:         DeletionStrategy(request.Body.Strategy),
		TransferTargetID: target,
		PurgeConfirmed:   purgeConfirmed,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.StartIamUserDeletion400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.StartIamUserDeletion403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.StartIamUserDeletion409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.StartIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	body, ok := transportDeletion(value)
	if !ok {
		return transport.StartIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.StartIamUserDeletion202JSONResponse{Body: body, Headers: transport.StartIamUserDeletion202ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) GetIamUserDeletion(ctx context.Context, request transport.GetIamUserDeletionRequestObject) (transport.GetIamUserDeletionResponseObject, error) {
	if s.deletions == nil {
		return transport.GetIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	value, err := s.deletions.GetDeletion(ctx, requestHTTP(ctx).actorID, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.GetIamUserDeletion400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.GetIamUserDeletion403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.GetIamUserDeletion404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		default:
			return transport.GetIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	body, ok := transportDeletion(value)
	if !ok {
		return transport.GetIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.GetIamUserDeletion200JSONResponse{Body: body, Headers: transport.GetIamUserDeletion200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) CancelIamUserDeletion(ctx context.Context, request transport.CancelIamUserDeletionRequestObject) (transport.CancelIamUserDeletionResponseObject, error) {
	if s.deletions == nil {
		return transport.CancelIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	err := s.deletions.CancelDeletion(ctx, requestHTTP(ctx).actorID, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.CancelIamUserDeletion400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.CancelIamUserDeletion403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.CancelIamUserDeletion404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.CancelIamUserDeletion409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.CancelIamUserDeletion500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.CancelIamUserDeletion204Response{Headers: transport.CancelIamUserDeletion204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func transportDeletion(value Deletion) (transport.AccountDeletion, bool) {
	id, err := uuid.Parse(value.ID)
	if err != nil {
		return transport.AccountDeletion{}, false
	}
	result := transport.AccountDeletion{
		Id: id, AccountId: value.AccountID, Strategy: transport.AccountDeletionStrategy(value.Strategy),
		Status: transport.AccountDeletionStatus(value.Status), AuditReference: value.AuditReference,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
	if value.TransferTargetID != "" {
		target := transport.Identifier(value.TransferTargetID)
		result.TransferTargetId = &target
	}
	return result, true
}
