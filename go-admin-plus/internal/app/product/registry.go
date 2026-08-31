// Package product is the only composition point for the complete Go Admin Plus product.
package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/audit"
	auditmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/audit/migrations/0011-audit"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo"
	productsmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/demo/migrations/0010-products"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files"
	filesmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files/migrations/0010-files"
	capacitymigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/files/migrations/0020-capacity"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/generator"
	configmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/generator/migrations/0010-config"
	sessionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0010-session-schema"
	administrationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0020-administration-schema"
	bootstraprecoverymigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0030-bootstrap-recovery"
	sessionprotectionmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0040-session-protection"
	datascopemigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/migrations/0050-data-scope"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization"
	organizationmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/organization/migrations"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/scheduler"
	schedulermigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/scheduler/migrations"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/settings"
	settingsmigration "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/settings/migrations/0010-settings"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations"
	reliableruntime "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/migrations/reliable-runtime"
)

var ErrCapabilityRegistration = errors.New("product capability registration failed")

type ModuleID string

const (
	ModuleIAM          ModuleID = "iam"
	ModuleAudit        ModuleID = "audit"
	ModuleOrganization ModuleID = "organization"
	ModuleSettings     ModuleID = "settings"
	ModuleGenerator    ModuleID = "generator"
	ModuleScheduler    ModuleID = "scheduler"
	ModuleDemo         ModuleID = "demo"
	ModuleFiles        ModuleID = "files"
)

type ModuleDefinition struct {
	ID               ModuleID
	MigrationModules []string
	RegistersMenu    bool
}

type ProfileDefinition struct {
	Profile config.Profile
	Dialect database.Dialect
}

type CapabilityRegistrar interface {
	Register(context.Context, capabilities.ModuleCapabilities) error
}

type capabilityRegistration struct {
	module   ModuleID
	register func(context.Context, CapabilityRegistrar) error
}

var moduleDefinitions = []ModuleDefinition{
	{ID: ModuleIAM, MigrationModules: []string{
		"iam-session",
		"iam-administration",
		"iam-bootstrap-recovery",
		"iam-session-protection",
		"iam-data-scope",
	}},
	{ID: ModuleAudit, MigrationModules: []string{"audit"}, RegistersMenu: true},
	{ID: ModuleOrganization, MigrationModules: []string{"organization"}, RegistersMenu: true},
	{ID: ModuleSettings, MigrationModules: []string{"settings"}, RegistersMenu: true},
	{ID: ModuleGenerator, MigrationModules: []string{"generator-config"}, RegistersMenu: true},
	{ID: ModuleScheduler, MigrationModules: []string{"scheduler"}, RegistersMenu: true},
	{ID: ModuleDemo, MigrationModules: []string{"demo-products"}, RegistersMenu: true},
	{ID: ModuleFiles, MigrationModules: []string{"files", "files-capacity"}, RegistersMenu: true},
}

var capabilityRegistrations = []capabilityRegistration{
	{module: ModuleAudit, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return audit.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleOrganization, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return organization.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleSettings, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return settings.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleGenerator, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return generator.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleScheduler, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return scheduler.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleDemo, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return demo.RegisterCapabilities(ctx, registrar)
	}},
	{module: ModuleFiles, register: func(ctx context.Context, registrar CapabilityRegistrar) error {
		return files.RegisterCapabilities(ctx, registrar)
	}},
}

func Modules() []ModuleDefinition {
	modules := make([]ModuleDefinition, len(moduleDefinitions))
	for index, module := range moduleDefinitions {
		modules[index] = module
		modules[index].MigrationModules = append([]string(nil), module.MigrationModules...)
	}
	return modules
}

func Profiles() []ProfileDefinition {
	return []ProfileDefinition{
		{Profile: config.ProfileServerPostgres, Dialect: database.DialectPostgres},
		{Profile: config.ProfileServerSQLite, Dialect: database.DialectSQLite},
		{Profile: config.ProfileDesktopSQLite, Dialect: database.DialectSQLite},
	}
}

func NewMigrationRunner() (*migrations.Runner, error) {
	return migrations.NewRunner(
		sessionmigration.Provider{},
		administrationmigration.Provider{},
		bootstraprecoverymigration.Provider{},
		sessionprotectionmigration.Provider{},
		datascopemigration.Provider{},
		organizationmigration.Provider{},
		auditmigration.Provider{},
		productsmigration.Provider{},
		configmigration.Provider{},
		settingsmigration.Provider{},
		schedulermigration.Provider{},
		filesmigration.Provider{},
		capacitymigration.Provider{},
		reliableruntime.Provider{},
	)
}

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	if registrar == nil {
		return ErrCapabilityRegistration
	}
	for _, registration := range capabilityRegistrations {
		if err := registration.register(ctx, registrar); err != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return errors.Join(ErrCapabilityRegistration, contextErr)
			}
			return fmt.Errorf("%w: %s", ErrCapabilityRegistration, registration.module)
		}
	}
	return nil
}
