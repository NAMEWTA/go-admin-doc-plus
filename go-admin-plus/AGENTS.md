# Backend Contribution Contract

- Keep process entry points in `cmd/` and product composition in `internal/app/product`.
- Put business behavior, transport, persistence, permissions, and migrations in the owning `internal/modules/<module>` package.
- Use application or consumer-owned ports for cross-module collaboration; do not import another module's private implementation.
- Keep infrastructure in `internal/platform` and profile-specific host behavior in `internal/host` or the Desktop host.
- Every schema change must be a forward-only, globally unique migration for both SQLite and PostgreSQL.
- Server code must preserve both database profiles. Desktop code must remain SQLite-only and loopback-only.
- Run `go test ./...`, `go test -tags sqlite ./...`, `go vet ./...`, and `go mod tidy` before handoff.
