package organization

import (
	"context"
	"errors"

	transport "go-admin/internal/modules/organization/transport"
)

func (s *HTTPServer) ListOrganizationDepartments(ctx context.Context, _ transport.ListOrganizationDepartmentsRequestObject) (transport.ListOrganizationDepartmentsResponseObject, error) {
	values, err := s.service.ListDepartments(ctx, requestHTTP(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.ListOrganizationDepartments403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListOrganizationDepartments500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	body := make([]transport.Department, 0, len(values))
	for _, value := range values {
		body = append(body, transportDepartment(value))
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListOrganizationDepartments200JSONResponse{Body: body, Headers: transport.ListOrganizationDepartments200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) CreateOrganizationDepartment(ctx context.Context, request transport.CreateOrganizationDepartmentRequestObject) (transport.CreateOrganizationDepartmentResponseObject, error) {
	if request.Body == nil {
		return transport.CreateOrganizationDepartment400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.CreateDepartment(ctx, requestHTTP(ctx).actorID, DepartmentInput{Key: request.Body.Key, Name: request.Body.Name, ParentID: request.Body.ParentId, SortOrder: request.Body.SortOrder})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.CreateOrganizationDepartment400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.CreateOrganizationDepartment403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.CreateOrganizationDepartment404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.CreateOrganizationDepartment409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.CreateOrganizationDepartment500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.CreateOrganizationDepartment201JSONResponse{Body: transportDepartment(value), Headers: transport.CreateOrganizationDepartment201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) UpdateOrganizationDepartment(ctx context.Context, request transport.UpdateOrganizationDepartmentRequestObject) (transport.UpdateOrganizationDepartmentResponseObject, error) {
	if request.Body == nil {
		return transport.UpdateOrganizationDepartment400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := s.service.UpdateDepartment(ctx, requestHTTP(ctx).actorID, request.DepartmentId, DepartmentInput{Key: request.Body.Key, Name: request.Body.Name, ParentID: request.Body.ParentId, SortOrder: request.Body.SortOrder})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.UpdateOrganizationDepartment400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.UpdateOrganizationDepartment403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.UpdateOrganizationDepartment404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.UpdateOrganizationDepartment409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.UpdateOrganizationDepartment500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.UpdateOrganizationDepartment200JSONResponse{Body: transportDepartment(value), Headers: transport.UpdateOrganizationDepartment200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) DeleteOrganizationDepartment(ctx context.Context, request transport.DeleteOrganizationDepartmentRequestObject) (transport.DeleteOrganizationDepartmentResponseObject, error) {
	err := s.service.DeleteDepartment(ctx, requestHTTP(ctx).actorID, request.DepartmentId)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.DeleteOrganizationDepartment404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.DeleteOrganizationDepartment403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.DeleteOrganizationDepartment404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.DeleteOrganizationDepartment409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		default:
			return transport.DeleteOrganizationDepartment500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.DeleteOrganizationDepartment204Response{Headers: transport.DeleteOrganizationDepartment204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
