package generator

import (
	"context"
	"errors"

	transport "go-admin/internal/modules/generator/transport"
)

func (server *HTTPServer) GetGeneratorConfig(ctx context.Context, request transport.GetGeneratorConfigRequestObject) (transport.GetGeneratorConfigResponseObject, error) {
	draft, digest, err := server.generator.Config(ctx, requestValue(ctx).actorID, request.ModuleName)
	if err != nil {
		switch {
		case errors.Is(err, ErrDenied):
			return transport.GetGeneratorConfig403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.GetGeneratorConfig404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		default:
			return transport.GetGeneratorConfig500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.GetGeneratorConfig200JSONResponse{Body: transport.SavedGenerationConfig{Draft: transportDraft(draft), PreviewDigest: digest}, Headers: transport.GetGeneratorConfig200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func transportDraft(value Draft) transport.GenerationDraft {
	columns := make([]transport.ColumnDraft, 0, len(value.Columns))
	for _, column := range value.Columns {
		columns = append(columns, transport.ColumnDraft{Name: column.Name, Field: column.Field, Include: column.Include, Searchable: column.Searchable, Sortable: column.Sortable})
	}
	return transport.GenerationDraft{Module: value.Module, Entity: value.Entity, Plural: value.Plural, Table: transport.TableReference{Schema: value.Table.Schema, Name: value.Table.Name}, Columns: columns}
}

func (server *HTTPServer) ListGeneratorTables(ctx context.Context, _ transport.ListGeneratorTablesRequestObject) (transport.ListGeneratorTablesResponseObject, error) {
	values, err := server.generator.Tables(ctx, requestValue(ctx).actorID)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			return transport.ListGeneratorTables403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		}
		return transport.ListGeneratorTables500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
	}
	rows := make([]transport.TableReference, 0, len(values))
	for _, value := range values {
		rows = append(rows, transport.TableReference{Schema: value.Schema, Name: value.Name})
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.ListGeneratorTables200JSONResponse{Body: rows, Headers: transport.ListGeneratorTables200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) GetGeneratorTable(ctx context.Context, request transport.GetGeneratorTableRequestObject) (transport.GetGeneratorTableResponseObject, error) {
	value, err := server.generator.Describe(ctx, requestValue(ctx).actorID, TableRef{Schema: request.SchemaName, Name: request.TableName})
	if err != nil {
		switch {
		case errors.Is(err, ErrDenied):
			return transport.GetGeneratorTable403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.GetGeneratorTable404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		default:
			return transport.GetGeneratorTable500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	columns := make([]transport.ColumnMetadata, 0, len(value.Columns))
	for _, column := range value.Columns {
		columns = append(columns, transport.ColumnMetadata{Name: column.Name, DatabaseType: column.DatabaseType, Kind: transport.ColumnKind(column.Kind), Nullable: column.Nullable, PrimaryKey: column.PrimaryKey, Ordinal: column.Ordinal})
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.GetGeneratorTable200JSONResponse{Body: transport.TableMetadata{Table: transport.TableReference{Schema: value.Ref.Schema, Name: value.Ref.Name}, Columns: columns}, Headers: transport.GetGeneratorTable200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) PreviewGeneratorModule(ctx context.Context, request transport.PreviewGeneratorModuleRequestObject) (transport.PreviewGeneratorModuleResponseObject, error) {
	if request.Body == nil {
		return transport.PreviewGeneratorModule400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	draft := Draft{Module: request.Body.Module, Entity: request.Body.Entity, Plural: request.Body.Plural, Table: TableRef{Schema: request.Body.Table.Schema, Name: request.Body.Table.Name}}
	for _, column := range request.Body.Columns {
		draft.Columns = append(draft.Columns, ColumnDraft{Name: column.Name, Field: column.Field, Include: column.Include, Searchable: column.Searchable, Sortable: column.Sortable})
	}
	value, err := server.generator.Preview(ctx, requestValue(ctx).actorID, draft)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalid):
			return transport.PreviewGeneratorModule400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.PreviewGeneratorModule403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrNotFound):
			return transport.PreviewGeneratorModule404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil
		default:
			return transport.PreviewGeneratorModule500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	files := make([]transport.PreviewFile, 0, len(value.Files))
	for _, file := range value.Files {
		files = append(files, transport.PreviewFile{Path: file.Path, Content: file.Content, Sha256: file.SHA256})
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.PreviewGeneratorModule200JSONResponse{Body: transport.GenerationPreview{Token: value.Token, Digest: value.Digest, Module: value.Module, CreatedAt: value.CreatedAt, ExpiresAt: value.ExpiresAt, Files: files}, Headers: transport.PreviewGeneratorModule200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}

func (server *HTTPServer) WriteGeneratorModule(ctx context.Context, request transport.WriteGeneratorModuleRequestObject) (transport.WriteGeneratorModuleResponseObject, error) {
	if request.Body == nil {
		return transport.WriteGeneratorModule400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
	}
	value, err := server.generator.Write(ctx, requestValue(ctx).actorID, request.Body.PreviewToken, bool(request.Body.Confirmed))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalid), errors.Is(err, ErrPreviewStale):
			return transport.WriteGeneratorModule400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil
		case errors.Is(err, ErrDenied):
			return transport.WriteGeneratorModule403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil
		case errors.Is(err, ErrConflict):
			return transport.WriteGeneratorModule409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil
		case errors.Is(err, ErrGateFailed):
			return transport.WriteGeneratorModule422ApplicationProblemPlusJSONResponse{GateProblemApplicationProblemPlusJSONResponse: gateProblem(ctx)}, nil
		default:
			return transport.WriteGeneratorModule500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil
		}
	}
	csrf, cookie := responseHeaders(ctx)
	return transport.WriteGeneratorModule201JSONResponse{Body: transport.WriteResult{Token: value.Token, Directory: value.Directory, Files: value.Files}, Headers: transport.WriteGeneratorModule201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
