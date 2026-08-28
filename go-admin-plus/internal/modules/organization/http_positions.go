package organization

import (
	"context"
	"errors"

	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization/transport"
)

func (s *HTTPServer) ListOrganizationPositions(ctx context.Context, request transport.ListOrganizationPositionsRequestObject) (transport.ListOrganizationPositionsResponseObject, error) {
	search, page, pageSize := "", 1, 20
	if request.Params.Search != nil {
		search = *request.Params.Search
	}
	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}
	value, err := s.service.ListPositions(ctx, requestHTTP(ctx).actorID, search, page, pageSize)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return transport.ListOrganizationPositions400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		}
		if errors.Is(err, ErrDenied) {
			return transport.ListOrganizationPositions403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListOrganizationPositions500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	rows := make([]transport.Position, 0, len(value.Rows))
	for _, item := range value.Rows {
		rows = append(rows, transportPosition(item))
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListOrganizationPositions200JSONResponse{Body: transport.PositionPage{Rows: rows, Total: value.Total}, Headers: transport.ListOrganizationPositions200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) CreateOrganizationPosition(ctx context.Context, request transport.CreateOrganizationPositionRequestObject) (transport.CreateOrganizationPositionResponseObject, error) {
	if request.Body == nil {
		return transport.CreateOrganizationPosition400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.CreatePosition(ctx, requestHTTP(ctx).actorID, PositionInput{Key: request.Body.Key, Name: request.Body.Name, DepartmentID: request.Body.DepartmentId, Enabled: request.Body.Enabled})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.CreateOrganizationPosition400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.CreateOrganizationPosition403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.CreateOrganizationPosition404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.CreateOrganizationPosition409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.CreateOrganizationPosition500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.CreateOrganizationPosition201JSONResponse{Body: transportPosition(value), Headers: transport.CreateOrganizationPosition201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) UpdateOrganizationPosition(ctx context.Context, request transport.UpdateOrganizationPositionRequestObject) (transport.UpdateOrganizationPositionResponseObject, error) {
	if request.Body == nil {
		return transport.UpdateOrganizationPosition400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.UpdatePosition(ctx, requestHTTP(ctx).actorID, request.PositionId, PositionInput{Key: request.Body.Key, Name: request.Body.Name, DepartmentID: request.Body.DepartmentId, Enabled: request.Body.Enabled})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.UpdateOrganizationPosition400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.UpdateOrganizationPosition403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.UpdateOrganizationPosition404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.UpdateOrganizationPosition409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.UpdateOrganizationPosition500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.UpdateOrganizationPosition200JSONResponse{Body: transportPosition(value), Headers: transport.UpdateOrganizationPosition200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) DeleteOrganizationPosition(ctx context.Context, request transport.DeleteOrganizationPositionRequestObject) (transport.DeleteOrganizationPositionResponseObject, error) {
	err := s.service.DeletePosition(ctx, requestHTTP(ctx).actorID, request.PositionId)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.DeleteOrganizationPosition404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.DeleteOrganizationPosition403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.DeleteOrganizationPosition404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.DeleteOrganizationPosition409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.DeleteOrganizationPosition500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.DeleteOrganizationPosition204Response{Headers: transport.DeleteOrganizationPosition204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
