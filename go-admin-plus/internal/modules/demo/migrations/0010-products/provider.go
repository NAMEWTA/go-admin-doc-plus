package productsmigration

import (
	"embed"
	"errors"
	"io/fs"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

//go:embed postgres/*.sql sqlite/*.sql
var files embed.FS

type Provider struct{}

func (Provider) Module() string { return "demo-products" }

func (Provider) Migrations(dialect database.Dialect) (fs.FS, error) {
	switch dialect {
	case database.DialectPostgres:
		return fs.Sub(files, "postgres")
	case database.DialectSQLite:
		return fs.Sub(files, "sqlite")
	default:
		return nil, errors.New("demo product migration dialect is unsupported")
	}
}
