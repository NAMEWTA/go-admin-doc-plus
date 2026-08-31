//go:build !desktop_native_e2e

package product

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	desktophost "github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/host/desktop"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/config"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

func TestBuildDesktopExposesFirstSetupAfterMigration(t *testing.T) {
	ctx := context.Background()
	db, err := database.NewProcess().Open(ctx, database.Config{
		Profile: config.ProfileDesktopSQLite, SQLitePath: filepath.Join(t.TempDir(), "desktop.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repositoryRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatal(err)
	}
	built, err := BuildDesktop(ctx, db, desktophost.ProductOptions{
		FilesRoot: filepath.Join(t.TempDir(), "files"), RepositoryRoot: repositoryRoot,
		GeneratorOutputRoot: filepath.Join(t.TempDir(), "generated"), WorkerOwner: "desktop-setup-route-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if built.PrivateRoute == nil || built.PrivateRoute.Pattern != "POST "+desktopSetupPath {
		t.Fatal("desktop setup private route is unavailable")
	}
	request := httptest.NewRequest(http.MethodPost, desktopSetupPath, strings.NewReader(`{"action":"first-setup-state"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	built.PrivateRoute.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"required"`) {
		t.Fatalf("initial setup state = %d %s", response.Code, response.Body.String())
	}
}
