# Linux Server release

The Linux release contains two OCI images built from one source commit:

- `go-admin-plus-server` for `linux/amd64` and `linux/arm64`;
- `go-admin-plus-web` for `linux/amd64` and `linux/arm64`.

Compose exposes only the Web listener and offers two explicit profiles. Use
`postgres` for a PostgreSQL deployment and `sqlite` for a single-instance SQLite
deployment. Both profiles run the same product binary, apply forward-only
migrations before readiness, keep application files on named volumes, and run
with a read-only root filesystem and dropped capabilities.

```sh
cp deploy/compose/.env.example deploy/compose/.env
scripts/release/linux/prepare-secrets.sh
GO_ADMIN_COMPOSE_BUILD=1 scripts/release/linux/compose.sh postgres up -d --build --wait
scripts/release/linux/verify-compose.sh postgres
```

For SQLite, replace `postgres` with `sqlite`. Never run both profiles in the
same Compose project because they intentionally publish the same Web port.

Production deployment consumes immutable image digests recorded in
`release-output/provenance.json`. `scripts/release/linux/emit-artifacts.sh`
produces compressed OCI image archives, SPDX JSON SBOMs, provenance metadata,
and `SHA256SUMS`. It does not authenticate to or publish into a remote registry.

Runtime database credentials live only under the ignored
`deploy/compose/runtime/secrets/` directory. The typed JSON profile files contain
no credentials. Back up the selected database volume and the corresponding
product-data volume before an upgrade. Removing volumes is a deliberate data
destruction operation and is not part of normal shutdown.
