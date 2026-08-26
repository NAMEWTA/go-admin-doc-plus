package organization

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"go-admin/internal/modules/iam/authorization"
	organizationmigration "go-admin/internal/modules/organization/migrations"
	"go-admin/internal/platform/config"
	"go-admin/internal/platform/database"
	"go-admin/internal/platform/migrations"
)

const rootDepartmentID = "department-root-001"

type authorizerStub struct {
	scope       authorization.Scope
	err         error
	permissions []string
}

func (stub *authorizerStub) RequireInTx(_ context.Context, _ database.Tx, _, permission string) (authorization.Decision, error) {
	stub.permissions = append(stub.permissions, permission)
	return authorization.Decision{Scope: stub.scope}, stub.err
}

func TestOrganizationTreePositionAndProjectionLifecycle(t *testing.T) {
	db := organizationDatabase(t)
	authorizer := &authorizerStub{scope: authorization.ScopeAll}
	service, err := NewService(db, authorizer, WithClock(func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	actor := "account-admin-001"

	engineering, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "engineering", Name: "Engineering", ParentID: rootDepartmentID, SortOrder: 10})
	if err != nil {
		t.Fatal(err)
	}
	operations, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "operations", Name: "Operations", ParentID: rootDepartmentID, SortOrder: 20})
	if err != nil {
		t.Fatal(err)
	}
	platform, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "platform", Name: "Platform", ParentID: engineering.ID, SortOrder: 5})
	if err != nil {
		t.Fatal(err)
	}
	departments, err := service.ListDepartments(ctx, actor)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, len(departments))
	for index, department := range departments {
		keys[index] = department.Key
	}
	if !reflect.DeepEqual(keys, []string{"root", "engineering", "platform", "operations"}) {
		t.Fatalf("department preorder = %v", keys)
	}

	position, err := service.CreatePosition(ctx, actor, PositionInput{Key: "platform-engineer", Name: "Platform Engineer", DepartmentID: platform.ID, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	page, err := service.ListPositions(ctx, actor, "engineer", 1, 20)
	if err != nil || page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].ID != position.ID {
		t.Fatalf("position page = %#v, %v", page, err)
	}

	platform, err = service.UpdateDepartment(ctx, actor, platform.ID, DepartmentInput{Key: platform.Key, Name: platform.Name, ParentID: operations.ID, SortOrder: 5})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateDepartment(ctx, actor, operations.ID, DepartmentInput{Key: operations.Key, Name: operations.Name, ParentID: platform.ID, SortOrder: 20}); !errors.Is(err, ErrConflict) {
		t.Fatalf("cycle update = %v", err)
	}
	if err := service.DeleteDepartment(ctx, actor, platform.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("position-referenced department deletion = %v", err)
	}
	if err := service.DeleteDepartment(ctx, actor, rootDepartmentID); !errors.Is(err, ErrConflict) {
		t.Fatalf("protected root deletion = %v", err)
	}

	projection, err := NewProjectionAdapter(db)
	if err != nil {
		t.Fatal(err)
	}
	lineage, err := projection.DepartmentLineage(ctx, platform.ID)
	if err != nil || !reflect.DeepEqual(lineage.AncestorIDs, []string{operations.ID, rootDepartmentID}) {
		t.Fatalf("lineage = %#v, %v", lineage, err)
	}
	departmentID, err := projection.PositionDepartment(ctx, position.ID)
	if err != nil || departmentID != platform.ID {
		t.Fatalf("position projection = %q, %v", departmentID, err)
	}

	if err := service.DeletePosition(ctx, actor, position.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteDepartment(ctx, actor, platform.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteDepartment(ctx, actor, operations.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteDepartment(ctx, actor, engineering.ID); err != nil {
		t.Fatal(err)
	}
}

func TestOrganizationRejectsDuplicateInvalidAndUnauthorizedCommandsWithoutStateChange(t *testing.T) {
	db := organizationDatabase(t)
	authorizer := &authorizerStub{scope: authorization.ScopeAll}
	service, err := NewService(db, authorizer)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	actor := "account-admin-001"
	created, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "finance", Name: "Finance", ParentID: rootDepartmentID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "finance", Name: "Duplicate", ParentID: rootDepartmentID}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate key = %v", err)
	}
	if _, err := service.CreateDepartment(ctx, actor, DepartmentInput{Key: "bad key", Name: "Invalid", ParentID: rootDepartmentID}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid key = %v", err)
	}
	if _, err := service.ListPositions(ctx, actor, "", maximumPositionPage+1, 20); !errors.Is(err, ErrValidation) {
		t.Fatalf("unbounded page = %v", err)
	}

	authorizer.scope = authorization.ScopeSelf
	if _, err := service.CreatePosition(ctx, actor, PositionInput{Key: "accountant", Name: "Accountant", DepartmentID: created.ID, Enabled: true}); !errors.Is(err, ErrDenied) {
		t.Fatalf("self-scoped mutation = %v", err)
	}
	authorizer.err = authorization.ErrDenied
	if _, err := service.ListDepartments(ctx, actor); !errors.Is(err, ErrDenied) {
		t.Fatalf("permission denial = %v", err)
	}

	authorizer.err = nil
	authorizer.scope = authorization.ScopeAll
	page, err := service.ListPositions(ctx, actor, "", 1, 20)
	if err != nil || page.Total != 0 {
		t.Fatalf("rejected commands changed positions: %#v, %v", page, err)
	}
	departments, err := service.ListDepartments(ctx, actor)
	if err != nil || len(departments) != 2 {
		t.Fatalf("rejected commands changed departments: %#v, %v", departments, err)
	}
	if len(authorizer.permissions) == 0 {
		t.Fatal("permission seam was not exercised")
	}
}

func TestOrganizationContextAndProjectionFailuresStayStable(t *testing.T) {
	db := organizationDatabase(t)
	authorizer := &authorizerStub{scope: authorization.ScopeAll}
	service, err := NewService(db, authorizer)
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.ListDepartments(cancelled, "account-admin-001"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled list = %v", err)
	}
	projection, err := NewProjectionAdapter(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projection.DepartmentLineage(context.Background(), "missing-department"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing department projection = %v", err)
	}
	if _, err := projection.PositionDepartment(context.Background(), "missing-position"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing position projection = %v", err)
	}
}

func organizationDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewProcess().Open(context.Background(), database.Config{Profile: config.ProfileServerSQLite, SQLitePath: filepath.Join(t.TempDir(), "organization.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := migrations.NewRunner(organizationmigration.Provider{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := runner.Up(context.Background(), db)
	if err != nil || first.Applied != 1 {
		t.Fatalf("first migration = %#v, %v", first, err)
	}
	second, err := runner.Up(context.Background(), db)
	if err != nil || second.Applied != 0 {
		t.Fatalf("idempotent migration = %#v, %v", second, err)
	}
	return db
}
