package settings

import (
	"context"
	"errors"

	transport "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/settings/transport"
)

func (s *HTTPServer) ListSettingValues(ctx context.Context, r transport.ListSettingValuesRequestObject) (transport.ListSettingValuesResponseObject, error) {
	v, e := s.service.ListSettings(ctx, requestValue(ctx).actorID, Category(r.Params.Category), query(r.Params.Search, r.Params.Page, r.Params.PageSize))
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.ListSettingValues400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.ListSettingValues403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListSettingValues500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	rows := make([]transport.SettingValue, 0, len(v.Rows))
	for _, x := range v.Rows {
		rows = append(rows, settingTransport(x))
	}
	csrf, cookie := headers(ctx)
	return transport.ListSettingValues200JSONResponse{Body: transport.SettingValuePage{Rows: rows, Total: v.Total}, Headers: transport.ListSettingValues200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) CreateSettingValue(ctx context.Context, r transport.CreateSettingValueRequestObject) (transport.CreateSettingValueResponseObject, error) {
	if r.Body == nil {
		return transport.CreateSettingValue400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.CreateSetting(ctx, requestValue(ctx).actorID, SettingInput{Category: Category(r.Body.Category), Key: r.Body.Key, Label: r.Body.Label, Value: r.Body.Value, Description: r.Body.Description, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.CreateSettingValue400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.CreateSettingValue403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.CreateSettingValue409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.CreateSettingValue500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.CreateSettingValue201JSONResponse{Body: settingTransport(v), Headers: transport.CreateSettingValue201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) UpdateSettingValue(ctx context.Context, r transport.UpdateSettingValueRequestObject) (transport.UpdateSettingValueResponseObject, error) {
	if r.Body == nil {
		return transport.UpdateSettingValue400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.UpdateSetting(ctx, requestValue(ctx).actorID, r.SettingId.String(), int64(r.Body.Revision), SettingInput{Category: Category(r.Body.Category), Key: r.Body.Key, Label: r.Body.Label, Value: r.Body.Value, Description: r.Body.Description, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.UpdateSettingValue400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.UpdateSettingValue403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.UpdateSettingValue404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.UpdateSettingValue409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.UpdateSettingValue500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.UpdateSettingValue200JSONResponse{Body: settingTransport(v), Headers: transport.UpdateSettingValue200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) DeleteSettingValue(ctx context.Context, r transport.DeleteSettingValueRequestObject) (transport.DeleteSettingValueResponseObject, error) {
	e := s.service.DeleteSetting(ctx, requestValue(ctx).actorID, r.SettingId.String(), int64(r.Params.Revision))
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.DeleteSettingValue400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.DeleteSettingValue403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.DeleteSettingValue404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.DeleteSettingValue409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.DeleteSettingValue500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.DeleteSettingValue204Response{Headers: transport.DeleteSettingValue204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) ListDictionaryTypes(ctx context.Context, r transport.ListDictionaryTypesRequestObject) (transport.ListDictionaryTypesResponseObject, error) {
	v, e := s.service.ListDictionaries(ctx, requestValue(ctx).actorID, query(r.Params.Search, r.Params.Page, r.Params.PageSize))
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.ListDictionaryTypes400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.ListDictionaryTypes403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListDictionaryTypes500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	rows := make([]transport.DictionaryType, 0, len(v.Rows))
	for _, x := range v.Rows {
		rows = append(rows, dictionaryTransport(x))
	}
	csrf, cookie := headers(ctx)
	return transport.ListDictionaryTypes200JSONResponse{Body: transport.DictionaryTypePage{Rows: rows, Total: v.Total}, Headers: transport.ListDictionaryTypes200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) CreateDictionaryType(ctx context.Context, r transport.CreateDictionaryTypeRequestObject) (transport.CreateDictionaryTypeResponseObject, error) {
	if r.Body == nil {
		return transport.CreateDictionaryType400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.CreateDictionary(ctx, requestValue(ctx).actorID, DictionaryInput{Key: r.Body.Key, Name: r.Body.Name, Description: r.Body.Description, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.CreateDictionaryType400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.CreateDictionaryType403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.CreateDictionaryType409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.CreateDictionaryType500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.CreateDictionaryType201JSONResponse{Body: dictionaryTransport(v), Headers: transport.CreateDictionaryType201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) UpdateDictionaryType(ctx context.Context, r transport.UpdateDictionaryTypeRequestObject) (transport.UpdateDictionaryTypeResponseObject, error) {
	if r.Body == nil {
		return transport.UpdateDictionaryType400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.UpdateDictionary(ctx, requestValue(ctx).actorID, r.DictionaryId.String(), int64(r.Body.Revision), DictionaryInput{Key: r.Body.Key, Name: r.Body.Name, Description: r.Body.Description, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.UpdateDictionaryType400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.UpdateDictionaryType403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.UpdateDictionaryType404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.UpdateDictionaryType409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.UpdateDictionaryType500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.UpdateDictionaryType200JSONResponse{Body: dictionaryTransport(v), Headers: transport.UpdateDictionaryType200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) DeleteDictionaryType(ctx context.Context, r transport.DeleteDictionaryTypeRequestObject) (transport.DeleteDictionaryTypeResponseObject, error) {
	e := s.service.DeleteDictionary(ctx, requestValue(ctx).actorID, r.DictionaryId.String(), int64(r.Params.Revision))
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.DeleteDictionaryType400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.DeleteDictionaryType403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.DeleteDictionaryType404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.DeleteDictionaryType409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.DeleteDictionaryType500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.DeleteDictionaryType204Response{Headers: transport.DeleteDictionaryType204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (s *HTTPServer) ListDictionaryItems(ctx context.Context, r transport.ListDictionaryItemsRequestObject) (transport.ListDictionaryItemsResponseObject, error) {
	v, e := s.service.ListItems(ctx, requestValue(ctx).actorID, r.DictionaryId.String(), query(r.Params.Search, r.Params.Page, r.Params.PageSize))
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.ListDictionaryItems400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.ListDictionaryItems403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.ListDictionaryItems404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		return transport.ListDictionaryItems500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	rows := make([]transport.DictionaryItem, 0, len(v.Rows))
	for _, x := range v.Rows {
		rows = append(rows, itemTransport(x))
	}
	csrf, cookie := headers(ctx)
	return transport.ListDictionaryItems200JSONResponse{Body: transport.DictionaryItemPage{Rows: rows, Total: v.Total}, Headers: transport.ListDictionaryItems200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) CreateDictionaryItem(ctx context.Context, r transport.CreateDictionaryItemRequestObject) (transport.CreateDictionaryItemResponseObject, error) {
	if r.Body == nil {
		return transport.CreateDictionaryItem400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.CreateItem(ctx, requestValue(ctx).actorID, r.DictionaryId.String(), DictionaryItemInput{Value: r.Body.Value, Label: r.Body.Label, SortOrder: r.Body.SortOrder, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.CreateDictionaryItem400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.CreateDictionaryItem403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.CreateDictionaryItem404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.CreateDictionaryItem409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.CreateDictionaryItem500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.CreateDictionaryItem201JSONResponse{Body: itemTransport(v), Headers: transport.CreateDictionaryItem201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) UpdateDictionaryItem(ctx context.Context, r transport.UpdateDictionaryItemRequestObject) (transport.UpdateDictionaryItemResponseObject, error) {
	if r.Body == nil {
		return transport.UpdateDictionaryItem400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
	}
	v, e := s.service.UpdateItem(ctx, requestValue(ctx).actorID, r.ItemId.String(), int64(r.Body.Revision), DictionaryItemInput{Value: r.Body.Value, Label: r.Body.Label, SortOrder: r.Body.SortOrder, Enabled: r.Body.Enabled})
	if e != nil {
		if errors.Is(e, ErrValidation) || errors.Is(e, ErrSensitive) {
			return transport.UpdateDictionaryItem400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.UpdateDictionaryItem403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.UpdateDictionaryItem404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.UpdateDictionaryItem409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.UpdateDictionaryItem500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.UpdateDictionaryItem200JSONResponse{Body: itemTransport(v), Headers: transport.UpdateDictionaryItem200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) DeleteDictionaryItem(ctx context.Context, r transport.DeleteDictionaryItemRequestObject) (transport.DeleteDictionaryItemResponseObject, error) {
	e := s.service.DeleteItem(ctx, requestValue(ctx).actorID, r.ItemId.String(), int64(r.Params.Revision))
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.DeleteDictionaryItem400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.DeleteDictionaryItem403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.DeleteDictionaryItem404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		if errors.Is(e, ErrConflict) {
			return transport.DeleteDictionaryItem409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflict(ctx)}, nil
		}
		return transport.DeleteDictionaryItem500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	csrf, cookie := headers(ctx)
	return transport.DeleteDictionaryItem204Response{Headers: transport.DeleteDictionaryItem204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (s *HTTPServer) ListDictionaryOptions(ctx context.Context, r transport.ListDictionaryOptionsRequestObject) (transport.ListDictionaryOptionsResponseObject, error) {
	v, e := s.service.Options(ctx, requestValue(ctx).actorID, r.DictionaryKey)
	if e != nil {
		if errors.Is(e, ErrValidation) {
			return transport.ListDictionaryOptions400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validation(ctx)}, nil
		}
		if errors.Is(e, ErrDenied) {
			return transport.ListDictionaryOptions403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		if errors.Is(e, ErrNotFound) {
			return transport.ListDictionaryOptions404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFound(ctx)}, nil
		}
		return transport.ListDictionaryOptions500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internal(ctx)}, nil
	}
	rows := make([]transport.DictionaryOption, 0, len(v))
	for _, x := range v {
		rows = append(rows, transport.DictionaryOption{Value: x.Value, Label: x.Label})
	}
	csrf, cookie := headers(ctx)
	return transport.ListDictionaryOptions200JSONResponse{Body: rows, Headers: transport.ListDictionaryOptions200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
