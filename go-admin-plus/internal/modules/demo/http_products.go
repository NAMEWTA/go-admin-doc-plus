package demo

import (
	"context"
	"errors"

	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo/transport"
)

func (server *HTTPServer) ListDemoProducts(ctx context.Context, request transport.ListDemoProductsRequestObject) (transport.ListDemoProductsResponseObject, error) {
	query := ListQuery{Page: 1, PageSize: 20, Sort: "updatedAt", Direction: "descending"}
	if request.Params.Search != nil {
		query.Search = *request.Params.Search
	}
	if request.Params.Page != nil {
		query.Page = *request.Params.Page
	}
	if request.Params.PageSize != nil {
		query.PageSize = *request.Params.PageSize
	}
	if request.Params.Sort != nil {
		query.Sort = string(*request.Params.Sort)
	}
	if request.Params.Direction != nil {
		query.Direction = string(*request.Params.Direction)
	}
	value, err := server.service.List(ctx, requestValue(ctx).actorID, query)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.ListDemoProducts400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.ListDemoProducts403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationResponse(ctx)}, nil
		default:
			return transport.ListDemoProducts500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalResponse(ctx)}, nil
		}
	}
	rows := make([]transport.Product, 0, len(value.Rows))
	for _, product := range value.Rows {
		rows = append(rows, transportProduct(product))
	}
	csrf, cookie := headers(ctx)
	return transport.ListDemoProducts200JSONResponse{Body: transport.ProductPage{Rows: rows, Total: value.Total}, Headers: transport.ListDemoProducts200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) CreateDemoProduct(ctx context.Context, request transport.CreateDemoProductRequestObject) (transport.CreateDemoProductResponseObject, error) {
	if request.Body == nil {
		return transport.CreateDemoProduct400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
	}
	value, err := server.service.Create(ctx, requestValue(ctx).actorID, ProductInput{SKU: request.Body.Sku, Name: request.Body.Name, Description: request.Body.Description, PriceCents: int64(request.Body.PriceCents), Status: string(request.Body.Status)})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.CreateDemoProduct400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.CreateDemoProduct403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationResponse(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.CreateDemoProduct409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictResponse(ctx)}, nil
		default:
			return transport.CreateDemoProduct500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalResponse(ctx)}, nil
		}
	}
	csrf, cookie := headers(ctx)
	return transport.CreateDemoProduct201JSONResponse{Body: transportProduct(value), Headers: transport.CreateDemoProduct201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) GetDemoProduct(ctx context.Context, request transport.GetDemoProductRequestObject) (transport.GetDemoProductResponseObject, error) {
	value, err := server.service.Get(ctx, requestValue(ctx).actorID, request.ProductId.String())
	if err != nil {
		switch {
		case errors.Is(err, ErrDenied):
			return transport.GetDemoProduct403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationResponse(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.GetDemoProduct404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundResponse(ctx)}, nil
		default:
			return transport.GetDemoProduct500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalResponse(ctx)}, nil
		}
	}
	csrf, cookie := headers(ctx)
	return transport.GetDemoProduct200JSONResponse{Body: transportProduct(value), Headers: transport.GetDemoProduct200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) UpdateDemoProduct(ctx context.Context, request transport.UpdateDemoProductRequestObject) (transport.UpdateDemoProductResponseObject, error) {
	if request.Body == nil {
		return transport.UpdateDemoProduct400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
	}
	value, err := server.service.Update(ctx, requestValue(ctx).actorID, request.ProductId.String(), int64(request.Body.Revision), ProductInput{SKU: request.Body.Sku, Name: request.Body.Name, Description: request.Body.Description, PriceCents: int64(request.Body.PriceCents), Status: string(request.Body.Status)})
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.UpdateDemoProduct400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.UpdateDemoProduct403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationResponse(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.UpdateDemoProduct404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundResponse(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.UpdateDemoProduct409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictResponse(ctx)}, nil
		default:
			return transport.UpdateDemoProduct500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalResponse(ctx)}, nil
		}
	}
	csrf, cookie := headers(ctx)
	return transport.UpdateDemoProduct200JSONResponse{Body: transportProduct(value), Headers: transport.UpdateDemoProduct200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) DeleteDemoProducts(ctx context.Context, request transport.DeleteDemoProductsRequestObject) (transport.DeleteDemoProductsResponseObject, error) {
	if request.Body == nil {
		return transport.DeleteDemoProducts400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
	}
	targets := make([]DeleteTarget, 0, len(request.Body.Products))
	for _, target := range request.Body.Products {
		targets = append(targets, DeleteTarget{ID: target.Id.String(), Revision: int64(target.Revision)})
	}
	err := server.service.Delete(ctx, requestValue(ctx).actorID, targets)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			return transport.DeleteDemoProducts400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationResponse(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.DeleteDemoProducts403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationResponse(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.DeleteDemoProducts404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundResponse(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.DeleteDemoProducts409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictResponse(ctx)}, nil
		default:
			return transport.DeleteDemoProducts500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalResponse(ctx)}, nil
		}
	}
	csrf, cookie := headers(ctx)
	return transport.DeleteDemoProducts204Response{Headers: transport.DeleteDemoProducts204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
