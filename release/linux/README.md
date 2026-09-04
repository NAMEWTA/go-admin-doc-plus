# Linux Server Release

The Linux release contains `go-admin-plus-server` service archives for `linux/amd64` and
`linux/arm64`. Each archive includes both SQLite and PostgreSQL profile examples, systemd units,
and `SERVER-INSTALL.md` with the full deployment procedure.

The service is a Go binary. It owns forward migrations for SQLite while holding its file lock. For
PostgreSQL, run the one-shot `migrate` command before starting the API unit; the API and worker must
use the same data root and DSN configuration. Keep credentials in permission-restricted environment
or secret files, never in Git.

The release workflow builds both architectures with `CGO_ENABLED=0`, verifies the checksums, and
uploads the archives to the GitHub Release. The existing Compose definitions in `deploy/compose/`
remain available for operators who prefer container deployment.
