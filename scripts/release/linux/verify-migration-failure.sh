#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 1 ]] || { echo "usage: verify-migration-failure.sh <postgres|sqlite>" >&2; exit 2; }
profile=$1
case "$profile" in
  postgres) api_service=api-postgres ;;
  sqlite) api_service=api-sqlite ;;
  *) echo "unsupported Compose profile" >&2; exit 2 ;;
esac
[[ ${GO_ADMIN_COMPOSE_PROJECT:-} == go-admin-plus-failure-* ]] || {
  echo "failure fixture requires an isolated go-admin-plus-failure-* project" >&2
  exit 2
}

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
compose_dir="$repository/deploy/compose"
environment_file=${GO_ADMIN_COMPOSE_ENV:-"$compose_dir/.env"}
[[ -f "$environment_file" ]] || environment_file="$compose_dir/.env.example"
arguments=(docker compose -p "$GO_ADMIN_COMPOSE_PROJECT" --env-file "$environment_file" -f "$compose_dir/compose.yml" -f "$compose_dir/compose.migration-failure.yml" --profile "$profile")

cleanup() { "${arguments[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
set +e
"${arguments[@]}" up -d --wait --wait-timeout 60 "$api_service" > migration-failure.log 2>&1
result=$?
set -e
[[ $result -ne 0 ]] || { echo "invalid migration input did not block startup" >&2; exit 1; }
container=$("${arguments[@]}" ps -aq "$api_service")
if [[ -n "$container" ]]; then
  [[ $(docker inspect -f '{{.State.Running}}' "$container") == false ]]
fi
echo "GO_ADMIN_LINUX_MIGRATION_FAILURE_${profile^^}_PASS"
