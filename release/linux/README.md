# Linux AMD64 Compose release

This bundle runs Admin Web as the only host-published service. API, migration,
PostgreSQL and Redis stay on Compose networks. The default bind is
`127.0.0.1:8080`; put a TLS reverse proxy in front of it for remote access.

## Build and start from source

```sh
cp deploy/compose/.env.example deploy/compose/.env
./release/linux/prepare-config.sh
docker compose --env-file deploy/compose/.env \
  -f deploy/compose/compose.yml \
  -f deploy/compose/compose.build.yml \
  up -d --build --wait
```

The first command that changes data is the one-shot `migrate` service. A failed
migration prevents `api` from starting. Named volumes retain PostgreSQL, Redis
and server files across container replacement.

For a release, set `GO_ADMIN_VERSION` and the two image names in `.env` to the
published immutable version tags, pull them, and omit `compose.build.yml`.
Record the resolved image digests from the release manifest before rollout.

Runtime secrets live below `deploy/compose/runtime/`, which Git ignores.
`prepare-config.sh` creates them once and never overwrites non-empty values.
It protects the runtime directories with mode `0700`; the rendered settings
file is `0444` because local Compose secret mounts retain the host file mode
and the API/migration containers intentionally run as non-root.
Back up all three named volumes and the runtime secret directory before an
upgrade. Recovery is restoring those assets and starting the previous image
digests with the same Compose bundle.

## Verify and stop

`verify-compose.sh` requires Docker Compose, curl and jq. It checks same-origin
login, database persistence after API restart, health semantics, proxy 404
behavior and runtime least privilege.

```sh
./release/linux/verify-compose.sh
docker compose --env-file deploy/compose/.env \
  -f deploy/compose/compose.yml \
  -f deploy/compose/compose.build.yml down
```

Do not add `-v` to `down` unless permanent deletion of all named-volume data is
explicitly intended and a verified backup exists.
