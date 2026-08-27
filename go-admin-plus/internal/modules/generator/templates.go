package generator

import (
	"encoding/json"
	"fmt"
	"go/format"
	"sort"
	"strings"
)

func renderBase(model Model) ([]PreviewFile, error) {
	files := map[string]string{
		fmt.Sprintf("contracts/openapi/modules/%s.yaml", model.Module):                                                                                 generatedOpenAPI(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/model.go", model.Module):                                                                        generatedModel(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/mapping.go", model.Module):                                                                      generatedMapping(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/repository.go", model.Module):                                                                   generatedRepository(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/service.go", model.Module):                                                                      generatedService(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/iam_adapters.go", model.Module):                                                                 generatedIAMAdapters(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/http.go", model.Module):                                                                         generatedHTTP(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/http_records.go", model.Module):                                                                 generatedHTTPOperations(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/service_test.go", model.Module):                                                                 generatedServiceTest(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/service_postgres_test.go", model.Module):                                                        generatedPostgresServiceTest(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/permissions.go", model.Module):                                                                  generatedPermissions(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/migrations/0010-%s/provider.go", model.Module, model.Plural):                                    generatedMigrationProvider(model),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/migrations/0010-%s/postgres/7000000000000_%s.sql", model.Module, model.Plural, model.TableName): generatedMigration(model, true),
		fmt.Sprintf("go-admin-plus/internal/modules/%s/migrations/0010-%s/sqlite/7000000000000_%s.sql", model.Module, model.Plural, model.TableName):   generatedMigration(model, false),
		fmt.Sprintf("go-admin-plus-ui/packages/domains/%s/package.json", model.Module):                                                                 generatedDomainManifest(model),
		fmt.Sprintf("go-admin-plus-ui/packages/domains/%s/src/index.ts", model.Module):                                                                 generatedDomain(model),
		fmt.Sprintf("go-admin-plus-ui/packages/domains/%s/src/model.spec.ts", model.Module):                                                            generatedDomainTest(model),
		fmt.Sprintf("go-admin-plus-ui/packages/domains/%s/src/tsconfig.json", model.Module):                                                            generatedDomainTSConfig(),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/package.json", model.Module):                                                             generatedWebManifest(model),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/vitest.config.ts", model.Module):                                                         generatedWebVitestConfig(),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/src/index.ts", model.Module):                                                             generatedWebDomain(model),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/src/web-client.ts", model.Module):                                                        generatedWebClient(model),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/src/controller.spec.ts", model.Module):                                                   generatedWebTest(model),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/src/tsconfig.json", model.Module):                                                        generatedWebTSConfig(),
		fmt.Sprintf("go-admin-plus-ui/packages/web-domains/%s/src/%sPage.vue", model.Module, model.Entity):                                             generatedVuePage(model),
	}
	result := make([]PreviewFile, 0, len(files))
	for path, content := range files {
		if strings.HasSuffix(path, ".go") {
			formatted, err := format.Source([]byte(content))
			if err != nil {
				return nil, ErrInternal
			}
			content = string(formatted)
		}
		result = append(result, PreviewFile{Path: path, Content: content})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func generatedModel(model Model) string {
	needsTime := false
	for _, column := range model.Columns {
		needsTime = needsTime || column.Kind == KindTime
	}
	var out strings.Builder
	fmt.Fprintf(&out, "// Package %s is generated from the canonical OpenAPI and database metadata model.\npackage %s\n\n", model.Module, model.Module)
	if needsTime {
		out.WriteString("import (\n\t\"errors\"\n\t\"time\"\n)\n\n")
	} else {
		out.WriteString("import \"errors\"\n\n")
	}
	out.WriteString("var (\n\tErrDenied = errors.New(\"authorization denied\")\n\tErrAuthentication = errors.New(\"authentication required\")\n\tErrCSRF = errors.New(\"csrf rejected\")\n\tErrInvalid = errors.New(\"request invalid\")\n\tErrNotFound = errors.New(\"record not found\")\n\tErrConflict = errors.New(\"record conflict\")\n)\n\n")
	out.WriteString("var ErrInternal = errors.New(\"operation failed\")\n\n")
	fmt.Fprintf(&out, "type %s struct {\n", model.Entity)
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "\t%s %s\n", column.Field, column.GoType)
	}
	out.WriteString("}\n")
	fmt.Fprintf(&out, "\ntype %sInput struct {\n", model.Entity)
	for _, column := range model.Columns {
		if !standardColumn(column.Name) {
			fmt.Fprintf(&out, "\t%s %s\n", column.Field, column.GoType)
		}
	}
	out.WriteString("}\n")
	fmt.Fprintf(&out, "\ntype Page struct { Rows []%s; Total int }\n", model.Entity)
	return out.String()
}

func standardColumn(name string) bool {
	return name == "id" || name == "revision" || name == "created_at" || name == "updated_at"
}

func generatedMapping(model Model) string {
	var out strings.Builder
	fmt.Fprintf(&out, "package %s\n\n", model.Module)
	for _, column := range model.Columns {
		if column.Kind == KindTime {
			out.WriteString("import \"time\"\n\n")
			break
		}
	}
	fmt.Fprintf(&out, "type %sRecord struct {\n", model.EntityVariable)
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "\t%s %s\n", column.Field, column.GoType)
	}
	out.WriteString("}\n\n")
	fmt.Fprintf(&out, "func map%sRecord(value %sRecord) %s {\n\treturn %s{\n", model.Entity, model.EntityVariable, model.Entity, model.Entity)
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "\t\t%s: value.%s,\n", column.Field, column.Field)
	}
	out.WriteString("\t}\n}\n\n")
	fmt.Fprintf(&out, "func map%s(value %s) %sRecord {\n\treturn %sRecord{\n", model.Entity, model.Entity, model.EntityVariable, model.EntityVariable)
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "\t\t%s: value.%s,\n", column.Field, column.Field)
	}
	out.WriteString("\t}\n}\n")
	return out.String()
}

func generatedRepository(model Model) string {
	columnNames := make([]string, 0, len(model.Columns))
	placeholders := make([]string, 0, len(model.Columns))
	mutableColumns := make([]NormalizedColumn, 0, len(model.Columns)-1)
	sqliteSearchClauses := make([]string, 0)
	postgresSearchClauses := make([]string, 0)
	sortCases := make([]string, 0)
	for _, column := range model.Columns {
		columnNames = append(columnNames, quoteIdentifier(column.Name))
		placeholders = append(placeholders, "?")
		if !column.PrimaryKey {
			mutableColumns = append(mutableColumns, column)
		}
		if column.Searchable {
			sqliteSearchClauses = append(sqliteSearchClauses, "instr(CAST("+quoteIdentifier(column.Name)+" AS TEXT) COLLATE BINARY, ?) > 0")
			postgresSearchClauses = append(postgresSearchClauses, "strpos(CAST("+quoteIdentifier(column.Name)+" AS TEXT) COLLATE \"C\", ?) > 0")
		}
		if column.Sortable {
			sortCases = append(sortCases, fmt.Sprintf("case %q: order = repository.sortExpression(%q, %t)", lowerFirst(column.Field), quoteIdentifier(column.Name), column.Kind == KindString))
		}
	}
	var out strings.Builder
	fmt.Fprintf(&out, "package %s\n\nimport (\n\t\"context\"\n\t\"database/sql\"\n\t\"errors\"\n\n\t\"go-admin/internal/platform/database\"\n)\n\n", model.Module)
	out.WriteString("type Repository struct{ dialect database.Dialect }\n\nfunc (Repository) bind(query string) string { return query }\nfunc (repository Repository) sortExpression(column string, textual bool) string { if !textual { return column }; if repository.dialect == database.DialectPostgres { return column + ` COLLATE \"C\"` }; return column + \" COLLATE BINARY\" }\n\n")
	fmt.Fprintf(&out, "func (repository Repository) Create(ctx context.Context, tx database.Tx, value %s) error {\n\trecord := map%s(value)\n", model.Entity, model.Entity)
	fmt.Fprintf(&out, "\t_, err := tx.ExecContext(ctx, repository.bind(%q)", "INSERT INTO "+quoteIdentifier(model.TableName)+" ("+strings.Join(columnNames, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")")
	for _, column := range model.Columns {
		fmt.Fprintf(&out, ", record.%s", column.Field)
	}
	out.WriteString(")\n\treturn err\n}\n\n")
	fmt.Fprintf(&out, "func (repository Repository) Get(ctx context.Context, tx database.Tx, id %s) (%s, error) {\n\tvar record %sRecord\n", model.PrimaryKey.GoType, model.Entity, model.EntityVariable)
	fmt.Fprintf(&out, "\terr := tx.QueryRowContext(ctx, repository.bind(%q), id).Scan(", "SELECT "+strings.Join(columnNames, ", ")+" FROM "+quoteIdentifier(model.TableName)+" WHERE "+quoteIdentifier(model.PrimaryKey.Name)+" = ?")
	for index, column := range model.Columns {
		if index > 0 {
			out.WriteString(", ")
		}
		fmt.Fprintf(&out, "&record.%s", column.Field)
	}
	out.WriteString(")\n\tif errors.Is(err, sql.ErrNoRows) { return " + model.Entity + "{}, ErrNotFound }\n\tif err != nil { return " + model.Entity + "{}, err }\n\treturn map" + model.Entity + "Record(record), nil\n}\n\n")
	fmt.Fprintf(&out, "func (repository Repository) List(ctx context.Context, tx database.Tx, search, sortKey string, descending bool, limit, offset int) ([]%s, error) {\n", model.Entity)
	fmt.Fprintf(&out, "\torder := repository.sortExpression(%q, %t); switch sortKey { %s }; direction := \"ASC\"; if descending { direction = \"DESC\" }; query := %q; args := []any{}\n", quoteIdentifier(model.PrimaryKey.Name), model.PrimaryKey.Kind == KindString, strings.Join(sortCases, "; "), "SELECT "+strings.Join(columnNames, ", ")+" FROM "+quoteIdentifier(model.TableName))
	if len(sqliteSearchClauses) > 0 {
		fmt.Fprintf(&out, "\tif search != \"\" { clause := %q; if repository.dialect == database.DialectPostgres { clause = %q }; query += clause; for range %d { args = append(args, search) } }\n", " WHERE ("+strings.Join(sqliteSearchClauses, " OR ")+")", " WHERE ("+strings.Join(postgresSearchClauses, " OR ")+")", len(sqliteSearchClauses))
	}
	fmt.Fprintf(&out, "\tquery += \" ORDER BY \" + order + \" \" + direction + \", \" + repository.sortExpression(%q, %t) + \" ASC LIMIT ? OFFSET ?\"; args = append(args, limit, offset)\n", quoteIdentifier(model.PrimaryKey.Name), model.PrimaryKey.Kind == KindString)
	out.WriteString("\trows, err := tx.QueryContext(ctx, repository.bind(query), args...)\n")
	out.WriteString("\tif err != nil { return nil, err }\n\tdefer rows.Close()\n\tresult := make([]" + model.Entity + ", 0)\n\tfor rows.Next() {\n\t\tvar record " + model.EntityVariable + "Record\n\t\tif err := rows.Scan(")
	for index, column := range model.Columns {
		if index > 0 {
			out.WriteString(", ")
		}
		fmt.Fprintf(&out, "&record.%s", column.Field)
	}
	out.WriteString("); err != nil { return nil, err }\n\t\tresult = append(result, map" + model.Entity + "Record(record))\n\t}\n\treturn result, rows.Err()\n}\n\n")
	fmt.Fprintf(&out, "func (repository Repository) Count(ctx context.Context, tx database.Tx, search string) (int, error) {\n\tquery := %q; args := []any{}\n", "SELECT COUNT(*) FROM "+quoteIdentifier(model.TableName))
	if len(sqliteSearchClauses) > 0 {
		fmt.Fprintf(&out, "\tif search != \"\" { clause := %q; if repository.dialect == database.DialectPostgres { clause = %q }; query += clause; for range %d { args = append(args, search) } }\n", " WHERE ("+strings.Join(sqliteSearchClauses, " OR ")+")", " WHERE ("+strings.Join(postgresSearchClauses, " OR ")+")", len(sqliteSearchClauses))
	}
	out.WriteString("\tvar count int; err := tx.QueryRowContext(ctx, repository.bind(query), args...).Scan(&count); return count, err\n}\n\n")
	fmt.Fprintf(&out, "func (repository Repository) Update(ctx context.Context, tx database.Tx, value %s, expectedRevision int64) error {\n\trecord := map%s(value)\n", model.Entity, model.Entity)
	assignments := make([]string, 0, len(mutableColumns))
	for _, column := range mutableColumns {
		assignments = append(assignments, quoteIdentifier(column.Name)+" = ?")
	}
	fmt.Fprintf(&out, "\tresult, err := tx.ExecContext(ctx, repository.bind(%q)", "UPDATE "+quoteIdentifier(model.TableName)+" SET "+strings.Join(assignments, ", ")+" WHERE "+quoteIdentifier(model.PrimaryKey.Name)+" = ? AND "+quoteIdentifier("revision")+" = ?")
	for _, column := range mutableColumns {
		fmt.Fprintf(&out, ", record.%s", column.Field)
	}
	fmt.Fprintf(&out, ", record.%s, expectedRevision)\n", model.PrimaryKey.Field)
	out.WriteString("\tif err != nil { return err }\n\tcount, err := result.RowsAffected()\n\tif err != nil { return err }\n\tif count != 1 { return ErrConflict }\n\treturn nil\n}\n\n")
	fmt.Fprintf(&out, "func (repository Repository) Delete(ctx context.Context, tx database.Tx, id %s, expectedRevision int64) error {\n", model.PrimaryKey.GoType)
	fmt.Fprintf(&out, "\tresult, err := tx.ExecContext(ctx, repository.bind(%q), id, expectedRevision)\n", "DELETE FROM "+quoteIdentifier(model.TableName)+" WHERE "+quoteIdentifier(model.PrimaryKey.Name)+" = ? AND "+quoteIdentifier("revision")+" = ?")
	out.WriteString("\tif err != nil { return err }\n\tcount, err := result.RowsAffected()\n\tif err != nil { return err }\n\tif count != 1 { return ErrConflict }\n\treturn nil\n}\n")
	return out.String()
}

func generatedService(model Model) string {
	assignments := strings.Builder{}
	validation := strings.Builder{}
	sortCases := strings.Builder{}
	fmt.Fprintf(&sortCases, "case %q:\n", lowerFirst(model.PrimaryKey.Field))
	for _, column := range model.Columns {
		if !standardColumn(column.Name) {
			fmt.Fprintf(&assignments, "\t\t%s: input.%s,\n", column.Field, column.Field)
			if column.Kind == KindString {
				if column.Nullable {
					fmt.Fprintf(&validation, "\tif input.%s != nil && len([]rune(*input.%s)) > 500 { return ErrInvalid }\n", column.Field, column.Field)
				} else {
					fmt.Fprintf(&validation, "\tif strings.TrimSpace(input.%s) == \"\" || len([]rune(input.%s)) > 500 { return ErrInvalid }\n", column.Field, column.Field)
				}
			}
			if column.Kind == KindBytes {
				fmt.Fprintf(&validation, "\tif len(input.%s) > 1_000_000 { return ErrInvalid }\n", column.Field)
			}
		}
		if column.Sortable {
			fmt.Fprintf(&sortCases, "case %q:\n", lowerFirst(column.Field))
		}
	}
	return fmt.Sprintf(`package %s

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-admin/internal/platform/database"
)

type Database interface { WithinTx(context.Context, func(context.Context, database.Tx) error) error; Dialect() database.Dialect }
type Authorizer interface { RequireInTx(context.Context, database.Tx, string, string) error; Dialect() database.Dialect }
type Service struct { db Database; auth Authorizer; repository Repository; now func() time.Time }

func NewService(db Database, auth Authorizer) (*Service, error) {
	if db == nil || auth == nil || db.Dialect() != auth.Dialect() { return nil, ErrInvalid }
	return &Service{db: db, auth: auth, repository: Repository{dialect: db.Dialect()}, now: time.Now}, nil
}
func (service *Service) List(ctx context.Context, actorID string, page, pageSize int, search, sortKey, direction string) (Page, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || len([]rune(search)) > 100 || (direction != "ascending" && direction != "descending") { return Page{}, ErrInvalid }
	switch sortKey { %s default: return Page{}, ErrInvalid }
	var result Page
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := service.auth.RequireInTx(ctx, tx, actorID, PermissionRead); err != nil { return err }
		rows, err := service.repository.List(ctx, tx, strings.TrimSpace(search), sortKey, direction == "descending", pageSize, (page-1)*pageSize); if err != nil { return err }
		total, err := service.repository.Count(ctx, tx, strings.TrimSpace(search)); if err != nil { return err }
		result = Page{Rows: rows, Total: total}; return nil
	})
	return result, classifyServiceError(ctx, err)
}
func (service *Service) Get(ctx context.Context, actorID, id string) (%s, error) {
	if _, err := uuid.Parse(id); err != nil { return %s{}, ErrInvalid }
	var value %s
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error { if err := service.auth.RequireInTx(ctx, tx, actorID, PermissionRead); err != nil { return err }; var err error; value, err = service.repository.Get(ctx, tx, id); return err })
	return value, classifyServiceError(ctx, err)
}
func (service *Service) Create(ctx context.Context, actorID string, input %sInput) (%s, error) {
	if err := validateInput(input); err != nil { return %s{}, err }
	now := service.now().UTC(); value := %s{ID: uuid.NewString(), Revision: 1, CreatedAt: now, UpdatedAt: now,
%s	}
	var created %s
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error { if err := service.auth.RequireInTx(ctx, tx, actorID, PermissionWrite); err != nil { return err }; if err := service.repository.Create(ctx, tx, value); err != nil { return err }; created = value; return nil })
	return created, classifyServiceError(ctx, err)
}
func (service *Service) Update(ctx context.Context, actorID, id string, expectedRevision int64, input %sInput) (%s, error) {
	if _, err := uuid.Parse(id); err != nil || expectedRevision < 1 { return %s{}, ErrInvalid }; if err := validateInput(input); err != nil { return %s{}, err }
	var updated %s
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := service.auth.RequireInTx(ctx, tx, actorID, PermissionWrite); err != nil { return err }
		current, err := service.repository.Get(ctx, tx, id); if err != nil { return err }; if current.Revision != expectedRevision { return ErrConflict }
		updated = %s{ID: current.ID, Revision: current.Revision+1, CreatedAt: current.CreatedAt, UpdatedAt: service.now().UTC(),
%s		}
		return service.repository.Update(ctx, tx, updated, expectedRevision)
	})
	return updated, classifyServiceError(ctx, err)
}
func (service *Service) Delete(ctx context.Context, actorID, id string, revision int64) error {
	if _, err := uuid.Parse(id); err != nil || revision < 1 { return ErrInvalid }
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error { if err := service.auth.RequireInTx(ctx, tx, actorID, PermissionDelete); err != nil { return err }; if _, err := service.repository.Get(ctx, tx, id); err != nil { return err }; return service.repository.Delete(ctx, tx, id, revision) })
	return classifyServiceError(ctx, err)
}
func classifyServiceError(ctx context.Context, err error) error {
	if err == nil || errors.Is(err, ErrDenied) || errors.Is(err, ErrInvalid) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) { return err }
	if context.Cause(ctx) != nil { return context.Cause(ctx) }
	if errors.Is(err, sql.ErrNoRows) { return ErrNotFound }
	if uniqueConflict(err) { return ErrConflict }
	return ErrInternal
}
func validateInput(input %sInput) error {
%s	return nil
}
func uniqueConflict(err error) bool {
	var state interface{ SQLState() string }; if errors.As(err, &state) && state.SQLState() == "23505" { return true }
	var coded interface{ Code() int }; return errors.As(err, &coded) && (coded.Code() == 1555 || coded.Code() == 2067)
}
	`, model.Module, sortCases.String(),
		model.Entity, model.Entity, model.Entity,
		model.Entity, model.Entity, model.Entity, model.Entity, assignments.String(), model.Entity,
		model.Entity, model.Entity, model.Entity, model.Entity, model.Entity, model.Entity, assignments.String(),
		model.Entity, validation.String())
}

func generatedIAMAdapters(model Model) string {
	return fmt.Sprintf(`package %s

import (
	"context"
	"errors"
	"net/http"

	"go-admin/internal/modules/iam/authorization"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/database"
)
type IAMAuthorizationAdapter struct { service *authorization.Service; dialect database.Dialect }
func NewIAMAuthorizationAdapter(db Database) (*IAMAuthorizationAdapter, error) { if db == nil { return nil, ErrInvalid }; return &IAMAuthorizationAdapter{service: authorization.NewService(db), dialect: db.Dialect()}, nil }
func (adapter *IAMAuthorizationAdapter) Dialect() database.Dialect { return adapter.dialect }
func (adapter *IAMAuthorizationAdapter) RequireInTx(ctx context.Context, tx database.Tx, actorID, permission string) error { decision, err := adapter.service.RequireInTx(ctx, tx, actorID, permission); if errors.Is(err, authorization.ErrDenied) || (err == nil && decision.Scope != authorization.ScopeAll) { return ErrDenied }; return err }
type SessionRequestService interface { AuthorizeRequest(context.Context, string, string, bool) (session.Issued, error) }
type IAMSessionRequestAdapter struct { service SessionRequestService }
func NewIAMSessionRequestAdapter(service SessionRequestService) (*IAMSessionRequestAdapter, error) { if service == nil { return nil, ErrInvalid }; return &IAMSessionRequestAdapter{service: service}, nil }
func (*IAMSessionRequestAdapter) CookieName() string { return session.CookieName }
func (adapter *IAMSessionRequestAdapter) AuthorizeRequest(ctx context.Context, token, csrf string, mutation bool) (RequestIdentity, error) { grant, err := adapter.service.AuthorizeRequest(ctx, token, csrf, mutation); if err != nil { if errors.Is(err, session.ErrCSRF) { return RequestIdentity{}, ErrCSRF }; if errors.Is(err, session.ErrAuthentication) { return RequestIdentity{}, ErrAuthentication }; return RequestIdentity{}, err }; identity := RequestIdentity{ActorID: grant.Profile.ID, CSRF: grant.CSRF}; if grant.Rotated && grant.Token != "" { cookie := (&http.Cookie{Name: session.CookieName, Value: grant.Token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}).String(); identity.ReplacementCookie = &cookie }; return identity, nil }
`, model.Module)
}

func generatedHTTP(model Model) string {
	var mapping strings.Builder
	fmt.Fprintf(&mapping, "func transport%s(value %s) transport.%s { return transport.%s{\n", model.Entity, model.Entity, model.Entity, model.Entity)
	for _, column := range model.Columns {
		fmt.Fprintf(&mapping, "\t%s: %s,\n", transportGoField(column.Field), transportExpression(column, "value."+column.Field))
	}
	mapping.WriteString("} }\n")
	return fmt.Sprintf(`package %s

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"
	"go-admin/internal/contracts"
	transport "go-admin/internal/modules/%s/transport"
)
//go:embed transport/openapi.json
var openAPIDocument []byte
var tracePattern = regexp.MustCompile("^[a-f0-9]{16,64}$")
type RequestIdentity struct { ActorID, CSRF string; ReplacementCookie *string }
type RequestAuthenticator interface { CookieName() string; AuthorizeRequest(context.Context, string, string, bool) (RequestIdentity, error) }
type requestContextKey struct{}
type requestContext struct { actorID, csrf, trace string; cookie *string }
type HTTPServer struct { service *Service }
func NewHTTPHandler(service *Service, authenticator RequestAuthenticator, traceID contracts.TraceIDProvider) (http.Handler, error) {
	if service == nil || authenticator == nil || authenticator.CookieName() == "" || traceID == nil { return nil, ErrInvalid }
	server := &HTTPServer{service: service}
	strict := transport.NewStrictHandlerWithOptions(server, nil, transport.StrictHTTPServerOptions{RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) { writeProblem(w, makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400)) }, ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) { writeProblem(w, makeProblem(transport.Internal, "INTERNAL_ERROR", "Internal server error", traceID(r), 500)) }})
	router := transport.HandlerWithOptions(strict, transport.ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) { writeProblem(w, makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", traceID(r), 400)) }})
	validated, err := contracts.NewRequestValidator(openAPIDocument, router, traceID, contracts.RequestValidatorOptions{MaxBodyBytes: contracts.DefaultMaxRequestBodyBytes, AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}); if err != nil { return nil, err }
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Cache-Control", "no-store"); cookie, _ := r.Cookie(authenticator.CookieName()); token := ""; if cookie != nil { token = cookie.Value }; identity, err := authenticator.AuthorizeRequest(r.Context(), token, r.Header.Get("X-CSRF-Token"), r.Method != http.MethodGet && r.Method != http.MethodHead); if err != nil { if errors.Is(err, ErrCSRF) { writeProblem(w, makeProblem(transport.Authorization, "CSRF_REJECTED", "Request authorization failed", traceID(r), 403)) } else { writeProblem(w, makeProblem(transport.Authentication, "SESSION_REQUIRED", "Authentication required", traceID(r), 401)) }; return }; value := requestContext{actorID: identity.ActorID, csrf: identity.CSRF, cookie: identity.ReplacementCookie, trace: traceID(r)}; validated.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestContextKey{}, value))) }), nil
}

func requestValue(ctx context.Context) requestContext { value, _ := ctx.Value(requestContextKey{}).(requestContext); return value }
func makeProblem(category transport.ProblemCategory, code, title, trace string, status int) transport.Problem { if !tracePattern.MatchString(trace) { trace = "0000000000000000" }; return transport.Problem{Type: "urn:go-admin-plus:problem:"+strings.ToLower(strings.ReplaceAll(code, "_", "-")), Title: title, Status: status, Category: category, Code: code, TraceId: trace} }
func writeProblem(w http.ResponseWriter, value transport.Problem) { w.Header().Set("Content-Type", "application/problem+json"); w.WriteHeader(value.Status); _ = json.NewEncoder(w).Encode(value) }
func responseHeaders(ctx context.Context) (string, *string) { value := requestValue(ctx); return value.csrf, value.cookie }
func validationProblem(ctx context.Context) transport.ValidationProblemApplicationProblemPlusJSONResponse { csrf, cookie := responseHeaders(ctx); return transport.ValidationProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Validation, "REQUEST_INVALID", "Request validation failed", requestValue(ctx).trace, 400), Headers: transport.ValidationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}} }
func authorizationProblem(ctx context.Context) transport.AuthorizationProblemApplicationProblemPlusJSONResponse { csrf, cookie := responseHeaders(ctx); return transport.AuthorizationProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Authorization, "PERMISSION_DENIED", "Request authorization failed", requestValue(ctx).trace, 403), Headers: transport.AuthorizationProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}} }
func notFoundProblem(ctx context.Context) transport.NotFoundProblemApplicationProblemPlusJSONResponse { csrf, cookie := responseHeaders(ctx); return transport.NotFoundProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.NotFound, "RESOURCE_NOT_FOUND", "Resource not found", requestValue(ctx).trace, 404), Headers: transport.NotFoundProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}} }
func conflictProblem(ctx context.Context) transport.ConflictProblemApplicationProblemPlusJSONResponse { csrf, cookie := responseHeaders(ctx); return transport.ConflictProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Conflict, "RESOURCE_CONFLICT", "Resource conflict", requestValue(ctx).trace, 409), Headers: transport.ConflictProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}} }
func internalProblem(ctx context.Context) transport.InternalProblemApplicationProblemPlusJSONResponse { csrf, cookie := responseHeaders(ctx); return transport.InternalProblemApplicationProblemPlusJSONResponse{Body: makeProblem(transport.Internal, "INTERNAL_ERROR", "Internal server error", requestValue(ctx).trace, 500), Headers: transport.InternalProblemResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}} }
%s
`, model.Module, model.Module, mapping.String())
}

func transportGoField(field string) string {
	if field == "ID" {
		return "Id"
	}
	return field
}

func transportExpression(column NormalizedColumn, expression string) string {
	if column.Kind == KindUUID && !column.Nullable {
		return "uuid.MustParse(" + expression + ")"
	}
	if column.Kind == KindUUID && column.Nullable {
		return "func() *uuid.UUID { if " + expression + " == nil { return nil }; value := uuid.MustParse(*" + expression + "); return &value }()"
	}
	return expression
}

func generatedHTTPOperations(model Model) string {
	moduleTitle := strings.ToUpper(model.Module[:1]) + model.Module[1:]
	var createInput, updateInput strings.Builder
	for _, column := range model.Columns {
		if !standardColumn(column.Name) {
			fmt.Fprintf(&createInput, "%s: %s, ", column.Field, domainTransportExpression(column, "request.Body."+column.Field))
			fmt.Fprintf(&updateInput, "%s: %s, ", column.Field, domainTransportExpression(column, "request.Body."+column.Field))
		}
	}
	return fmt.Sprintf(`package %s

import (
	"context"
	"errors"

	transport "go-admin/internal/modules/%s/transport"
)
func (server *HTTPServer) List%s%sRecords(ctx context.Context, request transport.List%s%sRecordsRequestObject) (transport.List%s%sRecordsResponseObject, error) {
	page, pageSize, search, sortKey, direction := 1, 20, "", "id", "ascending"; if request.Params.Page != nil { page = *request.Params.Page }; if request.Params.PageSize != nil { pageSize = *request.Params.PageSize }; if request.Params.Search != nil { search = *request.Params.Search }; if request.Params.Sort != nil { sortKey = *request.Params.Sort }; if request.Params.Direction != nil { direction = string(*request.Params.Direction) }
	value, err := server.service.List(ctx, requestValue(ctx).actorID, page, pageSize, search, sortKey, direction); if err != nil { if errors.Is(err, ErrInvalid) { return transport.List%s%sRecords400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }; if errors.Is(err, ErrDenied) { return transport.List%s%sRecords403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil }; return transport.List%s%sRecords500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil }
	rows := make([]transport.%s, 0, len(value.Rows)); for _, row := range value.Rows { rows = append(rows, transport%s(row)) }; csrf, cookie := responseHeaders(ctx)
	return transport.List%s%sRecords200JSONResponse{Body: transport.%sPage{Rows: rows, Total: value.Total}, Headers: transport.List%s%sRecords200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (server *HTTPServer) Get%s%sRecord(ctx context.Context, request transport.Get%s%sRecordRequestObject) (transport.Get%s%sRecordResponseObject, error) {
	value, err := server.service.Get(ctx, requestValue(ctx).actorID, request.Id.String()); if err != nil { if errors.Is(err, ErrDenied) { return transport.Get%s%sRecord403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil }; if errors.Is(err, ErrNotFound) { return transport.Get%s%sRecord404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil }; return transport.Get%s%sRecord500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil }; csrf, cookie := responseHeaders(ctx); return transport.Get%s%sRecord200JSONResponse{Body: transport%s(value), Headers: transport.Get%s%sRecord200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (server *HTTPServer) Create%s%sRecord(ctx context.Context, request transport.Create%s%sRecordRequestObject) (transport.Create%s%sRecordResponseObject, error) {
	if request.Body == nil { return transport.Create%s%sRecord400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }
	value, err := server.service.Create(ctx, requestValue(ctx).actorID, %sInput{%s}); if err != nil { if errors.Is(err, ErrInvalid) { return transport.Create%s%sRecord400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }; if errors.Is(err, ErrDenied) { return transport.Create%s%sRecord403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil }; if errors.Is(err, ErrConflict) { return transport.Create%s%sRecord409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil }; return transport.Create%s%sRecord500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil }; csrf, cookie := responseHeaders(ctx); return transport.Create%s%sRecord201JSONResponse{Body: transport%s(value), Headers: transport.Create%s%sRecord201ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (server *HTTPServer) Update%s%sRecord(ctx context.Context, request transport.Update%s%sRecordRequestObject) (transport.Update%s%sRecordResponseObject, error) {
	if request.Body == nil { return transport.Update%s%sRecord400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }
	value, err := server.service.Update(ctx, requestValue(ctx).actorID, request.Id.String(), int64(request.Body.Revision), %sInput{%s}); if err != nil { if errors.Is(err, ErrInvalid) { return transport.Update%s%sRecord400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }; if errors.Is(err, ErrDenied) { return transport.Update%s%sRecord403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil }; if errors.Is(err, ErrNotFound) { return transport.Update%s%sRecord404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil }; if errors.Is(err, ErrConflict) { return transport.Update%s%sRecord409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil }; return transport.Update%s%sRecord500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil }; csrf, cookie := responseHeaders(ctx); return transport.Update%s%sRecord200JSONResponse{Body: transport%s(value), Headers: transport.Update%s%sRecord200ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
func (server *HTTPServer) Delete%s%sRecord(ctx context.Context, request transport.Delete%s%sRecordRequestObject) (transport.Delete%s%sRecordResponseObject, error) {
	err := server.service.Delete(ctx, requestValue(ctx).actorID, request.Id.String(), int64(request.Params.Revision)); if err != nil { if errors.Is(err, ErrInvalid) { return transport.Delete%s%sRecord400ApplicationProblemPlusJSONResponse{ValidationProblemApplicationProblemPlusJSONResponse: validationProblem(ctx)}, nil }; if errors.Is(err, ErrDenied) { return transport.Delete%s%sRecord403ApplicationProblemPlusJSONResponse{AuthorizationProblemApplicationProblemPlusJSONResponse: authorizationProblem(ctx)}, nil }; if errors.Is(err, ErrNotFound) { return transport.Delete%s%sRecord404ApplicationProblemPlusJSONResponse{NotFoundProblemApplicationProblemPlusJSONResponse: notFoundProblem(ctx)}, nil }; if errors.Is(err, ErrConflict) { return transport.Delete%s%sRecord409ApplicationProblemPlusJSONResponse{ConflictProblemApplicationProblemPlusJSONResponse: conflictProblem(ctx)}, nil }; return transport.Delete%s%sRecord500ApplicationProblemPlusJSONResponse{InternalProblemApplicationProblemPlusJSONResponse: internalProblem(ctx)}, nil }; csrf, cookie := responseHeaders(ctx); return transport.Delete%s%sRecord204Response{Headers: transport.Delete%s%sRecord204ResponseHeaders{XCSRFToken: csrf, SetCookie: cookie}}, nil
}
`, model.Module, model.Module,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		model.Entity, model.Entity, moduleTitle, model.Entity, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, model.Entity, createInput.String(), moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, model.Entity, updateInput.String(), moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity,
		moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity, moduleTitle, model.Entity)
}

func domainTransportExpression(column NormalizedColumn, expression string) string {
	if column.Kind == KindUUID && !column.Nullable {
		return expression + ".String()"
	}
	if column.Kind == KindUUID && column.Nullable {
		return "func() *string { if " + expression + " == nil { return nil }; value := " + expression + ".String(); return &value }()"
	}
	return expression
}

func generatedServiceTest(model Model) string {
	var inputValues strings.Builder
	searchField := ""
	for _, column := range model.Columns {
		if standardColumn(column.Name) {
			continue
		}
		value := `"value"`
		switch column.Kind {
		case KindInt64:
			value = "int64(7)"
		case KindBoolean:
			value = "true"
		case KindDecimal:
			value = "1.5"
		case KindTime:
			value = "time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)"
		case KindBytes:
			value = `[]byte("value")`
		}
		if column.Nullable {
			value = "pointer(" + value + ")"
		}
		fmt.Fprintf(&inputValues, "%s: %s, ", column.Field, value)
	}
	sortKey := lowerFirst(model.PrimaryKey.Field)
	for _, column := range model.Columns {
		if column.Sortable {
			sortKey = lowerFirst(column.Field)
			break
		}
	}
	search := ""
	for _, column := range model.Columns {
		if column.Searchable && column.Kind == KindString && !column.Nullable {
			search = "val"
			searchField = column.Field
			break
		}
	}
	fixtures, assertions := generatedRuntimeSearchTest(model, searchField, sortKey, "")
	up := strings.Split(generatedMigration(model, false), "-- +goose Down")[0]
	up = strings.TrimSpace(strings.TrimPrefix(up, "-- +goose Up"))
	template := `package {{MODULE}}

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"go-admin/internal/platform/database"
	_ "modernc.org/sqlite"
)

type runtimeTestDatabase struct{ db *bun.DB; dialect database.Dialect }
func (value runtimeTestDatabase) Dialect() database.Dialect { return value.dialect }
func (value runtimeTestDatabase) WithinTx(ctx context.Context, fn func(context.Context, database.Tx) error) error { return value.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error { return fn(ctx, tx) }) }
type runtimeTestAuthorizer struct{ allowed bool; dialect database.Dialect }
func (value *runtimeTestAuthorizer) Dialect() database.Dialect { return value.dialect }
func (value *runtimeTestAuthorizer) RequireInTx(context.Context, database.Tx, string, string) error { if !value.allowed { return ErrDenied }; return nil }
func pointer[T any](value T) *T { return &value }

func TestGeneratedSQLiteAuthorizedCRUDSearchSortRevisionAndRevoke(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:"); if err != nil { t.Fatal(err) }; defer sqlDB.Close()
	db := bun.NewDB(sqlDB, sqlitedialect.New()); defer db.Close()
	if _, err := db.Exec({{MIGRATION}}); err != nil { t.Fatal(err) }
	authorizer := &runtimeTestAuthorizer{allowed: true, dialect: database.DialectSQLite}
	service, err := NewService(runtimeTestDatabase{db: db, dialect: database.DialectSQLite}, authorizer); if err != nil { t.Fatal(err) }
	service.now = func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }
	input := {{ENTITY}}Input{ {{INPUT}} }
	created, err := service.Create(context.Background(), "actor-one", input); if err != nil { t.Fatal(err) }
	{{FIXTURES}}
	page, err := service.List(context.Background(), "actor-one", 1, 20, {{SEARCH}}, {{SORT}}, "ascending"); if err != nil || page.Total != 1 || len(page.Rows) != 1 { t.Fatalf("list total=%d rows=%d err=%v", page.Total, len(page.Rows), err) }
	{{ASSERTIONS}}
	updated, err := service.Update(context.Background(), "actor-one", created.ID, created.Revision, input); if err != nil || updated.Revision != 2 { t.Fatalf("update revision=%d err=%v", updated.Revision, err) }
	if _, err := service.Update(context.Background(), "actor-one", created.ID, created.Revision, input); !errors.Is(err, ErrConflict) { t.Fatalf("stale update: %v", err) }
	if _, err := service.Get(context.Background(), "actor-one", "00000000-0000-4000-8000-000000000099"); !errors.Is(err, ErrNotFound) { t.Fatalf("not found: %v", err) }
	authorizer.allowed = false
	if _, err := service.List(context.Background(), "actor-one", 1, 20, "", {{SORT}}, "ascending"); !errors.Is(err, ErrDenied) { t.Fatalf("revoked read: %v", err) }
	authorizer.allowed = true
	if err := service.Delete(context.Background(), "actor-one", updated.ID, updated.Revision); err != nil { t.Fatal(err) }
	if err := service.Delete(context.Background(), "actor-one", updated.ID, updated.Revision); !errors.Is(err, ErrNotFound) { t.Fatalf("deleted not found: %v", err) }
}

func TestServiceFailsClosedWithoutDependencies(t *testing.T) {
	if _, err := NewService(nil, nil); err != ErrInvalid { t.Fatalf("expected invalid dependencies, got %v", err) }
}
`
	return strings.NewReplacer("{{MODULE}}", model.Module, "{{ENTITY}}", model.Entity, "{{INPUT}}", inputValues.String(), "{{MIGRATION}}", fmt.Sprintf("%q", up), "{{SEARCH}}", fmt.Sprintf("%q", search), "{{SORT}}", fmt.Sprintf("%q", sortKey), "{{FIXTURES}}", fixtures, "{{ASSERTIONS}}", assertions).Replace(template)
}

func generatedPostgresServiceTest(model Model) string {
	var inputValues strings.Builder
	searchField := ""
	for _, column := range model.Columns {
		if standardColumn(column.Name) {
			continue
		}
		value := `"value"`
		switch column.Kind {
		case KindInt64:
			value = "int64(7)"
		case KindBoolean:
			value = "true"
		case KindDecimal:
			value = "1.5"
		case KindTime:
			value = "time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)"
		case KindBytes:
			value = `[]byte("value")`
		}
		if column.Nullable {
			value = "pointer(" + value + ")"
		}
		fmt.Fprintf(&inputValues, "%s: %s, ", column.Field, value)
	}
	sortKey, search := lowerFirst(model.PrimaryKey.Field), ""
	for _, column := range model.Columns {
		if column.Sortable {
			sortKey = lowerFirst(column.Field)
			break
		}
	}
	for _, column := range model.Columns {
		if column.Searchable && column.Kind == KindString && !column.Nullable {
			search = "val"
			searchField = column.Field
			break
		}
	}
	fixtures, assertions := generatedRuntimeSearchTest(model, searchField, sortKey, "postgres ")
	up := strings.Split(generatedMigration(model, true), "-- +goose Down")[0]
	up = strings.TrimSpace(strings.TrimPrefix(up, "-- +goose Up"))
	template := `package {{MODULE}}

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "os"
  "sort"
  "testing"
  "time"
  "github.com/uptrace/bun"
  "github.com/uptrace/bun/dialect/pgdialect"
  "go-admin/internal/platform/database"
  _ "github.com/jackc/pgx/v5/stdlib"
)

func TestGeneratedPostgresAuthorizedCRUDSearchSortRevisionAndRevoke(t *testing.T) {
  dsn := os.Getenv("GO_ADMIN_GENERATOR_POSTGRES_DSN")
  if dsn == "" { t.Skip("GO_ADMIN_GENERATOR_POSTGRES_DSN is required") }
  sqlDB, err := sql.Open("pgx", dsn); if err != nil { t.Fatal("postgres open failed") }; defer sqlDB.Close(); sqlDB.SetMaxOpenConns(1)
  db := bun.NewDB(sqlDB, pgdialect.New()); defer db.Close()
  schema := fmt.Sprintf("generated_{{MODULE}}_%d", time.Now().UnixNano())
  if _, err := db.Exec("CREATE SCHEMA \"" + schema + "\""); err != nil { t.Fatal("postgres schema create failed") }
  defer func() { _, _ = db.Exec("DROP SCHEMA \"" + schema + "\" CASCADE") }()
  if _, err := db.Exec("SET search_path TO \"" + schema + "\""); err != nil { t.Fatal("postgres search path failed") }
  if _, err := db.Exec({{MIGRATION}}); err != nil { t.Fatal("postgres migration failed") }
  authorizer := &runtimeTestAuthorizer{allowed: true, dialect: database.DialectPostgres}
  service, err := NewService(runtimeTestDatabase{db: db, dialect: database.DialectPostgres}, authorizer); if err != nil { t.Fatal(err) }
  service.now = func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }
  input := {{ENTITY}}Input{ {{INPUT}} }
  created, err := service.Create(context.Background(), "actor-one", input); if err != nil { t.Fatal(err) }
	{{FIXTURES}}
  page, err := service.List(context.Background(), "actor-one", 1, 20, {{SEARCH}}, {{SORT}}, "ascending"); if err != nil || page.Total != 1 { t.Fatalf("postgres list total=%d err=%v", page.Total, err) }
	{{ASSERTIONS}}
  updated, err := service.Update(context.Background(), "actor-one", created.ID, created.Revision, input); if err != nil { t.Fatal(err) }
  if _, err := service.Update(context.Background(), "actor-one", created.ID, created.Revision, input); !errors.Is(err, ErrConflict) { t.Fatalf("postgres stale update: %v", err) }
	if _, err := service.Get(context.Background(), "actor-one", "00000000-0000-4000-8000-000000000099"); !errors.Is(err, ErrNotFound) { t.Fatalf("postgres not found: %v", err) }
  authorizer.allowed = false; if _, err := service.List(context.Background(), "actor-one", 1, 20, "", {{SORT}}, "ascending"); !errors.Is(err, ErrDenied) { t.Fatalf("postgres revoked read: %v", err) }
  authorizer.allowed = true; if err := service.Delete(context.Background(), "actor-one", updated.ID, updated.Revision); err != nil { t.Fatal(err) }
	if err := service.Delete(context.Background(), "actor-one", updated.ID, updated.Revision); !errors.Is(err, ErrNotFound) { t.Fatalf("postgres deleted not found: %v", err) }
	if marker := os.Getenv("GO_ADMIN_GENERATOR_POSTGRES_MARKER"); marker != "" { if err := os.WriteFile(marker, []byte("pass"), 0600); err != nil { t.Fatal("postgres marker failed") } }
}
`
	return strings.NewReplacer("{{MODULE}}", model.Module, "{{ENTITY}}", model.Entity, "{{INPUT}}", inputValues.String(), "{{MIGRATION}}", fmt.Sprintf("%q", up), "{{SEARCH}}", fmt.Sprintf("%q", search), "{{SORT}}", fmt.Sprintf("%q", sortKey), "{{FIXTURES}}", fixtures, "{{ASSERTIONS}}", assertions).Replace(template)
}

func generatedRuntimeSearchTest(model Model, searchField, sortKey, label string) (string, string) {
	if searchField == "" || sortKey != lowerFirst(searchField) {
		return "", ""
	}
	fixtures := fmt.Sprintf(`fixtureNames := []string{"literal%%mark", "literal_mark", "é", "e\\u0301", "same", "same"}
	fixtureRows := make([]%s, 0, len(fixtureNames))
	for _, fixtureName := range fixtureNames { fixtureInput := input; fixtureInput.%s = fixtureName; row, createErr := service.Create(context.Background(), "actor-one", fixtureInput); if createErr != nil { t.Fatal(createErr) }; fixtureRows = append(fixtureRows, row) }`, model.Entity, searchField)
	assertions := fmt.Sprintf(`percent, percentErr := service.List(context.Background(), "actor-one", 1, 20, "%%", %q, "ascending"); if percentErr != nil || percent.Total != 1 || len(percent.Rows) != 1 || percent.Rows[0].%s != "literal%%mark" { t.Fatalf(%q, percent.Total, percentErr) }
	underscore, underscoreErr := service.List(context.Background(), "actor-one", 1, 20, "_", %q, "ascending"); if underscoreErr != nil || underscore.Total != 1 || len(underscore.Rows) != 1 || underscore.Rows[0].%s != "literal_mark" { t.Fatalf(%q, underscore.Total, underscoreErr) }
	ordered, orderedErr := service.List(context.Background(), "actor-one", 1, 20, "", %q, "ascending"); if orderedErr != nil || ordered.Total != 7 { t.Fatalf(%q, ordered.Total, orderedErr) }
	expected := append([]%s(nil), ordered.Rows...); sort.Slice(expected, func(i, j int) bool { if expected[i].%s != expected[j].%s { return expected[i].%s < expected[j].%s }; return fmt.Sprint(expected[i].ID) < fmt.Sprint(expected[j].ID) }); for index := range expected { if expected[index].ID != ordered.Rows[index].ID { t.Fatalf(%q, index) } }`,
		sortKey, searchField, label+"literal percent total=%d err=%v", sortKey, searchField, label+"literal underscore total=%d err=%v", sortKey, label+"ordered total=%d err=%v", model.Entity, searchField, searchField, searchField, searchField, label+"stable byte order mismatch at %d")
	return fixtures, assertions
}

func generatedDomainTSConfig() string {
	return `{"compilerOptions":{"allowImportingTsExtensions":true,"erasableSyntaxOnly":true,"lib":["ES2024","DOM","DOM.Iterable"],"module":"ESNext","moduleResolution":"Bundler","noEmit":true,"noUncheckedIndexedAccess":true,"strict":true,"target":"ES2024","types":["vitest/globals"]},"include":["./**/*.ts"]}
`
}
func generatedWebTSConfig() string {
	return `{"compilerOptions":{"allowImportingTsExtensions":true,"erasableSyntaxOnly":true,"lib":["ES2024","DOM","DOM.Iterable"],"module":"ESNext","moduleResolution":"Bundler","noEmit":true,"noUncheckedIndexedAccess":true,"strict":true,"target":"ES2024","types":["vite/client","vitest/globals"]},"include":["./**/*.ts","./**/*.vue"]}
`
}
func generatedWebVitestConfig() string {
	return `import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'
export default defineConfig({ plugins: [vue()], test: { environment: 'node' } })
`
}
func generatedDomainTest(model Model) string {
	return fmt.Sprintf(`import { describe, expect, it } from 'vitest'
import { empty%sInput, permissions, RequestError } from './index'
describe('%s domain contract', () => {
  it('publishes stable permission codes and a fresh input', () => {
    expect(permissions.read).toBe('%s.%s.read')
    expect(empty%sInput()).not.toBe(empty%sInput())
  })
  it('keeps stable failure categories', () => { expect(new RequestError('conflict').category).toBe('conflict') })
})
`, model.Entity, model.Module, model.Module, model.Plural, model.Entity, model.Entity)
}
func generatedWebTest(model Model) string {
	var values strings.Builder
	for _, column := range model.Columns {
		if standardColumn(column.Name) {
			continue
		}
		value := "'value'"
		switch column.Kind {
		case KindInt64, KindDecimal:
			value = "7"
		case KindBoolean:
			value = "true"
		case KindTime:
			value = "'2026-08-27T00:00:00Z'"
		case KindUUID:
			value = "'00000000-0000-4000-8000-000000000001'"
		case KindBytes:
			value = "'dmFsdWU='"
		}
		fmt.Fprintf(&values, "%s: %s, ", lowerFirst(column.Field), value)
	}
	template := `import { describe, expect, it, vi } from 'vitest'
import { create{{ENTITY}}Controller } from './index'
import type { {{ENTITY}}, {{ENTITY}}Client, {{ENTITY}}Input } from '@go-admin/domain-{{MODULE}}'
const row = { id: '00000000-0000-4000-8000-000000000001', revision: 1 } as {{ENTITY}}
const input = { {{INPUT}} } satisfies {{ENTITY}}Input
describe('{{MODULE}} web domain', () => {
  it('loads, creates, updates and confirms delete through shared list state', async () => {
    const client: {{ENTITY}}Client = { list: vi.fn(async () => ({ rows: [row], total: 1 })), get: vi.fn(async () => row), create: vi.fn(async () => row), update: vi.fn(async () => row), delete: vi.fn(async () => undefined) }
    const confirm = vi.fn(async () => true)
    const controller = create{{ENTITY}}Controller(client, confirm, { can: () => true })
    await controller.list.refresh(); expect(controller.list.snapshot().total).toBe(1)
    await controller.save(input); await controller.save({ ...input, id: row.id, revision: row.revision })
    await controller.remove(row)
    expect(client.create).toHaveBeenCalledOnce(); expect(client.update).toHaveBeenCalledOnce(); expect(client.delete).toHaveBeenCalledOnce(); expect(confirm).toHaveBeenCalledOnce()
  })
})
`
	return strings.NewReplacer("{{ENTITY}}", model.Entity, "{{MODULE}}", model.Module, "{{INPUT}}", values.String()).Replace(template)
}
func generatedWebClient(model Model) string {
	template := `import { createContractClient, RequestError, type {{ENTITY}}Client, type RequestFailure } from '@go-admin/domain-{{MODULE}}'
interface Problem { category?: string; code?: string }
const csrfPattern = /^[A-Za-z0-9_-]{43}$/
export const createWebClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): {{ENTITY}}Client => {
  let csrf = ''; let classified: RequestFailure | null = null; let tail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => { const result = tail.then(operation, operation); tail = result.then(() => undefined, () => undefined); return result }
  const contract = createContractClient({ baseUrl, fetch: async input => {
    const headers = new Headers(input.headers); if (csrf && input.method !== 'GET') headers.set('X-CSRF-Token', csrf)
    const response = await fetcher(new Request(input, { credentials: 'include', headers })); const next = response.headers.get('X-CSRF-Token')
    if (next !== null && !csrfPattern.test(next)) { csrf = ''; classified = 'relogin'; throw new RequestError('relogin') }
    const body = response.status >= 400 ? await response.clone().json().catch(() => null) as Problem | null : null
    classified = classify(response.status, body); if (next) csrf = next; else if (classified === 'relogin') csrf = ''; return response
  } })
  const failure = (error: unknown): never => { const category = classified ?? problemCategory(error); classified = null; throw new RequestError(category) }
  const required = <T>(data: T | undefined, error: unknown): T => error === undefined && data !== undefined ? data : failure(error)
  const completed = (error: unknown): void => { if (error !== undefined) failure(error) }
  return {
    list: query => serialized(async () => { const result = await contract.GET('/{{MODULE}}/{{PLURAL}}', { params: { query } }); return required(result.data, result.error) }),
    get: id => serialized(async () => { const result = await contract.GET('/{{MODULE}}/{{PLURAL}}/{id}', { params: { path: { id } } }); return required(result.data, result.error) }),
    create: body => serialized(async () => { const result = await contract.POST('/{{MODULE}}/{{PLURAL}}', { body }); return required(result.data, result.error) }),
    update: (id, body) => serialized(async () => { const result = await contract.PATCH('/{{MODULE}}/{{PLURAL}}/{id}', { params: { path: { id } }, body }); return required(result.data, result.error) }),
    delete: (id, revision) => serialized(async () => { const result = await contract.DELETE('/{{MODULE}}/{{PLURAL}}/{id}', { params: { path: { id }, query: { revision } } }); completed(result.error) }),
  }
}
const classify = (status: number, value: Problem | null): RequestFailure | null => {
  if (status === 401 || value?.code === 'CSRF_REJECTED') return 'relogin'; if (status === 403) return 'forbidden'; if (status === 400 || status === 422) return 'validation'; if (status === 404) return 'not-found'; if (status === 409) return 'conflict'; if (status >= 500) return 'unavailable'; return null
}
const problemCategory = (value: unknown): RequestFailure => typeof value === 'object' && value !== null && 'category' in value
  ? ({ authentication: 'relogin', authorization: 'forbidden', validation: 'validation', 'not-found': 'not-found', conflict: 'conflict' } as const)[String((value as Problem).category) as 'authentication'|'authorization'|'validation'|'not-found'|'conflict'] ?? 'unavailable' : 'unavailable'
`
	return strings.NewReplacer("{{ENTITY}}", model.Entity, "{{MODULE}}", model.Module, "{{PLURAL}}", model.Plural).Replace(template)
}

func generatedPermissions(model Model) string {
	return fmt.Sprintf(`package %s

import (
	"context"

	"go-admin/internal/modules/iam/authorization"
)

const (
	PermissionRead = "%s.%s.read"
	PermissionWrite = "%s.%s.write"
	PermissionDelete = "%s.%s.delete"
)

type CapabilityRegistrar interface { Register(context.Context, authorization.ModuleCapabilities) error }

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	return registrar.Register(ctx, authorization.ModuleCapabilities{
		Permissions: []authorization.PermissionDefinition{
			{Code: PermissionRead, Name: "Read %s"},
			{Code: PermissionWrite, Name: "Manage %s"},
			{Code: PermissionDelete, Name: "Delete %s"},
		},
		Menus: []authorization.MenuDefinition{{ID: "menu-%s-%s", Key: "%s-%s", Label: "%s", Path: "/%s/%s", PermissionCode: PermissionRead, SortOrder: 900}},
	})
}
`, model.Module, model.Module, model.Plural, model.Module, model.Plural, model.Module, model.Plural, model.Plural, model.Plural, model.Plural, model.Module, model.Plural, model.Module, model.Plural, model.Entity, model.Module, model.Plural)
}

func generatedMigrationProvider(model Model) string {
	return fmt.Sprintf(`package %smigration

import (
	"embed"
	"errors"
	"io/fs"

	"go-admin/internal/platform/database"
)

//go:embed postgres/*.sql sqlite/*.sql
var files embed.FS

type Provider struct{}
func (Provider) Module() string { return "%s-%s" }
func (Provider) Migrations(dialect database.Dialect) (fs.FS, error) {
	switch dialect {
	case database.DialectPostgres: return fs.Sub(files, "postgres")
	case database.DialectSQLite: return fs.Sub(files, "sqlite")
	default: return nil, errors.New("%s migration dialect is unsupported")
	}
}
`, model.Module, model.Module, model.Plural, model.Module)
}

func generatedMigration(model Model, postgres bool) string {
	definitions := make([]string, 0, len(model.Columns))
	for _, column := range model.Columns {
		typeName := migrationType(column.Kind, postgres)
		definition := "  " + quoteIdentifier(column.Name) + " " + typeName
		if column.PrimaryKey {
			definition += " PRIMARY KEY"
		}
		if !column.Nullable {
			definition += " NOT NULL"
		}
		definitions = append(definitions, definition)
	}
	return "-- +goose Up\nCREATE TABLE " + quoteIdentifier(model.TableName) + " (\n" + strings.Join(definitions, ",\n") + "\n);\n\n-- +goose Down\nDROP TABLE " + quoteIdentifier(model.TableName) + ";\n"
}

func migrationType(kind ColumnKind, postgres bool) string {
	if postgres {
		switch kind {
		case KindInt64:
			return "BIGINT"
		case KindBoolean:
			return "BOOLEAN"
		case KindDecimal:
			return "DOUBLE PRECISION"
		case KindTime:
			return "TIMESTAMPTZ"
		case KindUUID:
			return "UUID"
		case KindBytes:
			return "BYTEA"
		default:
			return "TEXT"
		}
	}
	switch kind {
	case KindInt64:
		return "INTEGER"
	case KindBoolean:
		return "INTEGER"
	case KindDecimal:
		return "REAL"
	case KindBytes:
		return "BLOB"
	case KindTime:
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func generatedDomainManifest(model Model) string {
	value := map[string]any{"name": "@go-admin/domain-" + model.Module, "version": "0.0.0", "private": true, "type": "module", "exports": map[string]string{".": "./src/index.ts"}, "dependencies": map[string]string{"@go-admin/api-client": "workspace:*"}, "scripts": map[string]string{"test": "vitest run src", "typecheck": "tsc -p src/tsconfig.json"}, "devDependencies": map[string]string{"typescript": "catalog:", "vitest": "catalog:"}}
	encoded, _ := json.MarshalIndent(value, "", "  ")
	return string(encoded) + "\n"
}

func generatedWebManifest(model Model) string {
	value := map[string]any{"name": "@go-admin/web-domain-" + model.Module, "version": "0.0.0", "private": true, "type": "module", "exports": map[string]string{".": "./src/index.ts"}, "dependencies": map[string]string{"@go-admin/domain-" + model.Module: "workspace:*", "@go-admin/ui": "workspace:*", "vue": "catalog:"}, "scripts": map[string]string{"test": "vitest run --config vitest.config.ts", "typecheck": "vue-tsc -p src/tsconfig.json"}, "devDependencies": map[string]string{"@vitejs/plugin-vue": "catalog:", "@vue/compiler-sfc": "catalog:", "typescript": "catalog:", "vite": "catalog:", "vitest": "catalog:", "vue-tsc": "catalog:"}}
	encoded, _ := json.MarshalIndent(value, "", "  ")
	return string(encoded) + "\n"
}

func generatedDomain(model Model) string {
	var empty strings.Builder
	var validation strings.Builder
	for _, column := range model.Columns {
		if standardColumn(column.Name) {
			continue
		}
		value := "''"
		switch column.Kind {
		case KindInt64, KindDecimal:
			value = "0"
		case KindBoolean:
			value = "false"
		}
		if column.Nullable {
			value = "null"
		}
		fmt.Fprintf(&empty, "%s: %s, ", lowerFirst(column.Field), value)
		field := "input." + lowerFirst(column.Field)
		switch column.Kind {
		case KindString:
			if column.Nullable {
				fmt.Fprintf(&validation, "  if (%s !== null && Array.from(%s).length > 500) return false\n", field, field)
			} else {
				fmt.Fprintf(&validation, "  if (%s.trim() === '' || Array.from(%s).length > 500) return false\n", field, field)
			}
		case KindInt64:
			if column.Nullable {
				fmt.Fprintf(&validation, "  if (%s !== null && !Number.isSafeInteger(%s)) return false\n", field, field)
			} else {
				fmt.Fprintf(&validation, "  if (!Number.isSafeInteger(%s)) return false\n", field)
			}
		case KindDecimal:
			if column.Nullable {
				fmt.Fprintf(&validation, "  if (%s !== null && !Number.isFinite(%s)) return false\n", field, field)
			} else {
				fmt.Fprintf(&validation, "  if (!Number.isFinite(%s)) return false\n", field)
			}
		case KindTime:
			if column.Nullable {
				fmt.Fprintf(&validation, "  if (%s !== null && !Number.isFinite(Date.parse(%s))) return false\n", field, field)
			} else {
				fmt.Fprintf(&validation, "  if (!Number.isFinite(Date.parse(%s))) return false\n", field)
			}
		case KindUUID:
			if column.Nullable {
				fmt.Fprintf(&validation, "  if (%s !== null && !uuidPattern.test(%s)) return false\n", field, field)
			} else {
				fmt.Fprintf(&validation, "  if (!uuidPattern.test(%s)) return false\n", field)
			}
		}
	}
	template := `import type { components } from './generated/client'
export { createContractClient } from './generated/client'
export type { operations, paths } from './generated/client'
export type {{ENTITY}} = components['schemas']['{{ENTITY}}']
export type {{ENTITY}}Input = components['schemas']['{{ENTITY}}Input']
export type Update{{ENTITY}}Request = components['schemas']['Update{{ENTITY}}Request']
export type {{ENTITY}}Page = components['schemas']['{{ENTITY}}Page']
export type RequestFailure = 'relogin' | 'forbidden' | 'validation' | 'conflict' | 'not-found' | 'unavailable'
export type PermissionCode = '{{MODULE}}.{{PLURAL}}.read' | '{{MODULE}}.{{PLURAL}}.write' | '{{MODULE}}.{{PLURAL}}.delete'
export const permissions = { read: '{{MODULE}}.{{PLURAL}}.read', write: '{{MODULE}}.{{PLURAL}}.write', delete: '{{MODULE}}.{{PLURAL}}.delete' } as const satisfies Readonly<Record<string, PermissionCode>>
export interface {{ENTITY}}Query { readonly search: string; readonly page: number; readonly pageSize: number; readonly sort: string; readonly direction: 'ascending' | 'descending' }
export interface {{ENTITY}}Client {
  list(query: {{ENTITY}}Query): Promise<{{ENTITY}}Page>; get(id: string): Promise<{{ENTITY}}>;
  create(input: {{ENTITY}}Input): Promise<{{ENTITY}}>; update(id: string, input: Update{{ENTITY}}Request): Promise<{{ENTITY}}>;
  delete(id: string, revision: number): Promise<void>
}
export class RequestError extends Error { readonly category: RequestFailure; constructor(category: RequestFailure) { super(category); this.category = category } }
export const empty{{ENTITY}}Input = (): {{ENTITY}}Input => ({ {{EMPTY}} })
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
export const validate{{ENTITY}}Input = (input: Readonly<{{ENTITY}}Input>): boolean => {
{{VALIDATION}}  return true
}
`
	return strings.NewReplacer("{{ENTITY}}", model.Entity, "{{MODULE}}", model.Module, "{{PLURAL}}", model.Plural, "{{EMPTY}}", empty.String(), "{{VALIDATION}}", validation.String()).Replace(template)
}

func generatedWebDomain(model Model) string {
	template := `import { RequestError, empty{{ENTITY}}Input, permissions, validate{{ENTITY}}Input, type {{ENTITY}}, type {{ENTITY}}Client, type {{ENTITY}}Input, type PermissionCode, type RequestFailure } from '@go-admin/domain-{{MODULE}}'
import { createListController, type ListController } from '@go-admin/ui'
export interface Filters { readonly search: string }
export interface CapabilityPort { can(permission: PermissionCode): boolean }
export interface {{ENTITY}}Controller {
  readonly list: ListController<Filters, {{ENTITY}}, string>; readonly busy: boolean; readonly pendingRepair: boolean; readonly projectionVisible: boolean
  failure(): RequestFailure | null; can(permission: PermissionCode): boolean; empty(): {{ENTITY}}Input
  save(value: {{ENTITY}}Input & { id?: string; revision?: number }): Promise<'completed'|'failed'|'refresh-failed'|'busy'>
  remove(value: {{ENTITY}}): Promise<'completed'|'cancelled'|'failed'|'refresh-failed'|'busy'>; repairProjection(): Promise<'completed'|'refresh-failed'|'busy'>
}
export const create{{ENTITY}}Controller = (client: {{ENTITY}}Client, confirmDelete: () => Promise<boolean>, capabilities: CapabilityPort): {{ENTITY}}Controller => {
  let busy = false; let pending = false; let failure: RequestFailure | null = null; let visible = false
	  const record = (error: unknown) => { failure = error instanceof RequestError ? error.category : 'unavailable'; if (failure === 'relogin' || failure === 'forbidden' || failure === 'unavailable') { visible = false; rawList.clearSelection() } }
	  const rawList = createListController<Filters, {{ENTITY}}, string>({ initialFilters: () => ({ search: '' }), normalizeFilters: value => ({ search: value.search.trim() }), rowKey: row => row.id,
	    load: async request => { if (!capabilities.can(permissions.read)) throw new RequestError('forbidden'); try { const result = await client.list({ search: request.filters.search, page: request.page, pageSize: request.pageSize, sort: request.sort?.key ?? 'updatedAt', direction: request.sort?.direction ?? 'descending' }); failure = null; visible = true; return result } catch (error) { record(error); throw error } } })
	  const list: ListController<Filters, {{ENTITY}}, string> = { snapshot() { const value = rawList.snapshot(); return visible && capabilities.can(permissions.read) ? value : { ...value, rows: [], total: 0, selectedKeys: [] } }, refresh: () => rawList.refresh(), search: value => rawList.search(value), reset: () => rawList.reset(), setPage: value => rawList.setPage(value), setPageSize: value => rawList.setPageSize(value), setSort: value => rawList.setSort(value), select: rows => { if (visible && capabilities.can(permissions.read)) rawList.select(rows) }, clearSelection: () => rawList.clearSelection() }
  const refresh = async () => { try { await list.refresh(); pending = false; return 'completed' as const } catch (error) { record(error); return 'refresh-failed' as const } }
  return { list, get busy() { return busy }, get pendingRepair() { return pending }, get projectionVisible() { return visible && capabilities.can(permissions.read) }, failure: () => failure,
    can: permission => visible && capabilities.can(permissions.read) && capabilities.can(permission), empty: empty{{ENTITY}}Input,
	    async save(value) { if (busy) return 'busy'; if (pending) return 'refresh-failed'; if (!visible || !capabilities.can(permissions.write)) { record(new RequestError('forbidden')); return 'failed' }; const { id: _id, revision: _revision, ...candidate } = value; if (!validate{{ENTITY}}Input(candidate)) { failure = 'validation'; return 'failed' }; busy = true; failure = null; try { try { if (value.id) { if (!Number.isSafeInteger(value.revision) || value.revision! < 1) return 'failed'; const { id, revision, ...input } = value; await client.update(id, { ...input, revision: revision! }) } else await client.create(value) } catch (error) { record(error); return 'failed' }; pending = true; return await refresh() } finally { busy = false } },
	    async remove(value) { if (busy) return 'busy'; if (pending) return 'refresh-failed'; if (!visible || !capabilities.can(permissions.delete)) { record(new RequestError('forbidden')); return 'failed' }; busy = true; try { if (!await confirmDelete()) return 'cancelled'; if (!visible || !capabilities.can(permissions.delete)) { record(new RequestError('forbidden')); return 'failed' }; try { await client.delete(value.id, value.revision) } catch (error) { record(error); return 'failed' }; pending = true; return await refresh() } finally { busy = false } },
    async repairProjection() { if (busy) return 'busy'; if (!pending) return 'completed'; busy = true; try { return await refresh() } finally { busy = false } },
  }
}
export { createWebClient } from './web-client'
export { default as {{ENTITY}}Page } from './{{ENTITY}}Page.vue'
`
	return strings.NewReplacer("{{ENTITY}}", model.Entity, "{{MODULE}}", model.Module).Replace(template)
}

func generatedVuePage(model Model) string {
	var editValues, formFields strings.Builder
	for _, column := range model.Columns {
		if standardColumn(column.Name) {
			continue
		}
		name := lowerFirst(column.Field)
		fmt.Fprintf(&editValues, "%s: row.%s, ", name, name)
		switch column.Kind {
		case KindBoolean:
			fmt.Fprintf(&formFields, "<label><input v-model=\"form.%s\" type=\"checkbox\"> %s</label>", name, column.Field)
		case KindInt64, KindDecimal:
			fmt.Fprintf(&formFields, "<label>%s <input v-model.number=\"form.%s\" name=\"%s\" type=\"number\"></label>", column.Field, name, name)
		default:
			fmt.Fprintf(&formFields, "<label>%s <input v-model=\"form.%s\" name=\"%s\"></label>", column.Field, name, name)
		}
	}
	template := `<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { permissions, type {{ENTITY}}, type {{ENTITY}}Input } from '@go-admin/domain-{{MODULE}}'
import type { {{ENTITY}}Controller } from './index'
const props = defineProps<{ controller: {{ENTITY}}Controller }>()
const revision = ref(0); const search = ref(''); const editing = ref<{{ENTITY}} | null>(null)
const form = reactive<{{ENTITY}}Input & { id?: string; revision?: number }>(props.controller.empty())
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const visible = computed(() => { void revision.value; return props.controller.projectionVisible })
const canWrite = computed(() => props.controller.can(permissions.write)); const canDelete = computed(() => props.controller.can(permissions.delete))
const settle = async (operation: () => Promise<unknown>) => { try { await operation() } finally { revision.value += 1 } }
const reset = () => { editing.value = null; Object.assign(form, props.controller.empty()); delete form.id; delete form.revision }
const edit = (row: {{ENTITY}}) => { editing.value = row; Object.assign(form, { {{EDIT_VALUES}}id: row.id, revision: row.revision }) }
const save = async () => { const result = await props.controller.save({ ...form }); if (result === 'completed') reset(); revision.value += 1 }
const remove = (row: {{ENTITY}}) => settle(() => props.controller.remove(row))
onMounted(() => { void settle(() => props.controller.list.refresh()) })
</script>
<template><section class="generated-records" aria-labelledby="records-title">
  <header><div><h1 id="records-title">{{ENTITY}} records</h1><p>{{ visible ? snapshot.total : 0 }} records</p></div><button v-if="controller.pendingRepair" type="button" :disabled="controller.busy" @click="settle(() => controller.repairProjection())">Refresh results</button></header>
  <p v-if="failure" role="alert">{{ failure }}</p>
  <form v-if="visible" class="search" @submit.prevent="settle(() => controller.list.search({ search }))"><label>Search <input v-model="search"></label><button :disabled="controller.busy">Search</button><button type="button" :disabled="controller.busy" @click="search = ''; settle(() => controller.list.reset())">Reset</button></form>
  <div v-if="visible" class="workspace"><div class="table"><table><thead><tr>{{HEADERS}}<th>Actions</th></tr></thead><tbody><tr v-for="row in snapshot.rows" :key="row.id">{{CELLS}}<td><button v-if="canWrite" type="button" @click="edit(row)">Edit</button><button v-if="canDelete" type="button" @click="remove(row)">Delete</button></td></tr></tbody></table>
    <nav aria-label="Pagination"><button :disabled="controller.busy || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">Previous</button><span>Page {{ snapshot.page }}</span><button :disabled="controller.busy || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">Next</button></nav></div>
    <form v-if="canWrite" class="editor" @submit.prevent="save"><h2>{{ editing ? 'Edit' : 'Create' }} {{ENTITY}}</h2>{{FORM_FIELDS}}<div><button type="submit" :disabled="controller.busy">Save</button><button type="button" :disabled="controller.busy" @click="reset">Cancel</button></div></form>
  </div>
</section></template>
<style scoped>
.generated-records { display:grid; gap:16px; color:#17202a } header,nav { display:flex; align-items:center; justify-content:space-between; gap:12px } h1,h2,p { margin:0 }.search { display:flex; align-items:end; gap:8px }.workspace { display:grid; grid-template-columns:minmax(0,1fr) minmax(260px,340px); gap:16px }.table { min-width:0; overflow:auto } table { width:100%%; border-collapse:collapse } th,td { padding:8px; border-bottom:1px solid #dfe6e9; text-align:left }.editor { display:grid; align-content:start; gap:10px; padding-left:16px; border-left:1px solid #dfe6e9 } label { display:grid; gap:4px } input,button { font:inherit } [role="alert"] { padding:8px; border-left:3px solid #b42318; background:#fff1f0 } @media(max-width:720px){.workspace{grid-template-columns:1fr}.editor{padding-left:0;border-left:0;border-top:1px solid #dfe6e9;padding-top:16px}}
</style>
`
	return strings.NewReplacer("{{ENTITY}}", model.Entity, "{{MODULE}}", model.Module, "{{EDIT_VALUES}}", editValues.String(), "{{FORM_FIELDS}}", formFields.String(), "{{HEADERS}}", vueHeaders(model), "{{CELLS}}", vueCells(model), "100%%", "100%").Replace(template)
}

func vueHeaders(model Model) string {
	var out strings.Builder
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "<th>%s</th>", column.Field)
	}
	return out.String()
}
func vueCells(model Model) string {
	var out strings.Builder
	for _, column := range model.Columns {
		fmt.Fprintf(&out, "<td>{{ row.%s }}</td>", lowerFirst(column.Field))
	}
	return out.String()
}

func generatedOpenAPI(model Model) string {
	var properties, required, inputProperties, inputRequired strings.Builder
	for _, column := range model.Columns {
		typeValue := column.OpenAPIType
		if column.Nullable {
			typeValue = "[" + column.OpenAPIType + ", 'null']"
		}
		formatValue := ""
		if column.OpenAPIFormat != "" {
			formatValue = ", format: " + column.OpenAPIFormat
		}
		line := fmt.Sprintf("        %s: {type: %s%s}\n", lowerFirst(column.Field), typeValue, formatValue)
		properties.WriteString(line)
		if !standardColumn(column.Name) {
			inputProperties.WriteString(line)
		}
		if !column.Nullable {
			if required.Len() > 0 {
				required.WriteString(", ")
			}
			required.WriteString(lowerFirst(column.Field))
			if !standardColumn(column.Name) {
				if inputRequired.Len() > 0 {
					inputRequired.WriteString(", ")
				}
				inputRequired.WriteString(lowerFirst(column.Field))
			}
		}
	}
	moduleTitle := strings.ToUpper(model.Module[:1]) + model.Module[1:]
	updateRequired := "revision"
	if inputRequired.Len() > 0 {
		updateRequired = inputRequired.String() + ", revision"
	}
	template := `openapi: 3.1.0
jsonSchemaDialect: https://json-schema.org/draft/2020-12/schema
info:
  title: {{ENTITY}} API
  version: 1.0.0
  description: Generated single-table CRUD module.
  license: {name: MIT, identifier: MIT}
servers: [{url: /api}]
x-go-admin-module: {{MODULE}}
x-go-admin-codegen:
  owner: {{MODULE}}
  goPackage: {{MODULE}}transport
  goOutput: go-admin-plus/internal/modules/{{MODULE}}/transport/openapi.gen.go
  typescriptOutput: go-admin-plus-ui/packages/domains/{{MODULE}}/src/generated
security: [{SessionCookie: []}]
paths:
  /{{MODULE}}/{{PLURAL}}:
    get:
      summary: List {{ENTITY}} records
      operationId: list{{MODULE_TITLE}}{{ENTITY}}Records
      parameters:
        - {name: search, in: query, schema: {type: string, maxLength: 100}}
        - {name: page, in: query, schema: {type: integer, minimum: 1, maximum: 1000000, default: 1}}
        - {name: pageSize, in: query, schema: {type: integer, minimum: 1, maximum: 100, default: 20}}
        - {name: sort, in: query, schema: {type: string, maxLength: 64}}
        - {name: direction, in: query, schema: {type: string, enum: [ascending, descending], default: ascending}}
      responses:
        '200': {description: Records., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/json: {schema: {$ref: '#/components/schemas/{{ENTITY}}Page'}}}}
        '400': {$ref: '#/components/responses/ValidationProblem'}
        '401': {$ref: '#/components/responses/AuthenticationProblem'}
        '403': {$ref: '#/components/responses/AuthorizationProblem'}
        '500': {$ref: '#/components/responses/InternalProblem'}
    post:
      summary: Create a {{ENTITY}} record
      operationId: create{{MODULE_TITLE}}{{ENTITY}}Record
      security: [{SessionCookie: [], CsrfToken: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/{{ENTITY}}Input'}}}}
      responses:
        '201': {description: Created., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/json: {schema: {$ref: '#/components/schemas/{{ENTITY}}'}}}}
        '400': {$ref: '#/components/responses/ValidationProblem'}
        '401': {$ref: '#/components/responses/AuthenticationProblem'}
        '403': {$ref: '#/components/responses/AuthorizationProblem'}
        '409': {$ref: '#/components/responses/ConflictProblem'}
        '500': {$ref: '#/components/responses/InternalProblem'}
  /{{MODULE}}/{{PLURAL}}/{id}:
    parameters: [{name: id, in: path, required: true, schema: {type: string, format: uuid}}]
    get:
      summary: Get a {{ENTITY}} record
      operationId: get{{MODULE_TITLE}}{{ENTITY}}Record
      responses:
        '200': {description: Record., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/json: {schema: {$ref: '#/components/schemas/{{ENTITY}}'}}}}
        '401': {$ref: '#/components/responses/AuthenticationProblem'}
        '403': {$ref: '#/components/responses/AuthorizationProblem'}
        '404': {$ref: '#/components/responses/NotFoundProblem'}
        '500': {$ref: '#/components/responses/InternalProblem'}
    patch:
      summary: Update a {{ENTITY}} record
      operationId: update{{MODULE_TITLE}}{{ENTITY}}Record
      security: [{SessionCookie: [], CsrfToken: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/Update{{ENTITY}}Request'}}}}
      responses:
        '200': {description: Updated., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/json: {schema: {$ref: '#/components/schemas/{{ENTITY}}'}}}}
        '400': {$ref: '#/components/responses/ValidationProblem'}
        '401': {$ref: '#/components/responses/AuthenticationProblem'}
        '403': {$ref: '#/components/responses/AuthorizationProblem'}
        '404': {$ref: '#/components/responses/NotFoundProblem'}
        '409': {$ref: '#/components/responses/ConflictProblem'}
        '500': {$ref: '#/components/responses/InternalProblem'}
    delete:
      summary: Delete a {{ENTITY}} record
      operationId: delete{{MODULE_TITLE}}{{ENTITY}}Record
      security: [{SessionCookie: [], CsrfToken: []}]
      parameters: [{name: revision, in: query, required: true, schema: {type: integer, minimum: 1}}]
      responses:
        '204': {description: Deleted., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}}
        '400': {$ref: '#/components/responses/ValidationProblem'}
        '401': {$ref: '#/components/responses/AuthenticationProblem'}
        '403': {$ref: '#/components/responses/AuthorizationProblem'}
        '404': {$ref: '#/components/responses/NotFoundProblem'}
        '409': {$ref: '#/components/responses/ConflictProblem'}
        '500': {$ref: '#/components/responses/InternalProblem'}
components:
  securitySchemes:
    SessionCookie: {$ref: ../components/security-schemes.yaml#/SessionCookie}
    CsrfToken: {$ref: ../components/security-schemes.yaml#/CsrfToken}
  headers:
    CsrfToken: {required: true, schema: {type: string, minLength: 43, maxLength: 43}}
    SetCookie: {required: false, schema: {type: string, maxLength: 4096}}
  schemas:
    {{ENTITY}}:
      type: object
      additionalProperties: false
      required: [{{REQUIRED}}]
      properties:
{{PROPERTIES}}    {{ENTITY}}Input:
      type: object
      additionalProperties: false
      required: [{{INPUT_REQUIRED}}]
      properties:
{{INPUT_PROPERTIES}}    Update{{ENTITY}}Request:
      type: object
      additionalProperties: false
      required: [{{UPDATE_REQUIRED}}]
      properties:
{{INPUT_PROPERTIES}}        revision: {type: integer, minimum: 1}
    {{ENTITY}}Page:
      type: object
      additionalProperties: false
      required: [rows, total]
      properties:
        rows: {type: array, items: {$ref: '#/components/schemas/{{ENTITY}}'}}
        total: {type: integer, minimum: 0}
    Problem: {$ref: ../components/schemas.yaml#/Problem}
  responses:
    AuthenticationProblem: {description: Authentication required., content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
    AuthorizationProblem: {description: Authorization failed., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
    ValidationProblem: {description: Request invalid., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
    NotFoundProblem: {description: Resource not found., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
    ConflictProblem: {description: Resource conflict., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
    InternalProblem: {description: Internal failure., headers: {X-CSRF-Token: {$ref: '#/components/headers/CsrfToken'}, Set-Cookie: {$ref: '#/components/headers/SetCookie'}}, content: {application/problem+json: {schema: {$ref: '#/components/schemas/Problem'}}}}
`
	replacements := map[string]string{"{{MODULE}}": model.Module, "{{MODULE_TITLE}}": moduleTitle, "{{ENTITY}}": model.Entity, "{{PLURAL}}": model.Plural, "{{REQUIRED}}": required.String(), "{{PROPERTIES}}": properties.String(), "{{INPUT_REQUIRED}}": inputRequired.String(), "{{UPDATE_REQUIRED}}": updateRequired, "{{INPUT_PROPERTIES}}": inputProperties.String()}
	for key, value := range replacements {
		template = strings.ReplaceAll(template, key, value)
	}
	return template
}
