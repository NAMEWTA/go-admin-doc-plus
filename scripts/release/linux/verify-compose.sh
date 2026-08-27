#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 1 ]] || { echo "usage: verify-compose.sh <postgres|sqlite>" >&2; exit 2; }
profile=$1
case "$profile" in
  postgres) api_service=api-postgres; web_service=web-postgres ;;
  sqlite) api_service=api-sqlite; web_service=web-sqlite ;;
  *) echo "unsupported Compose profile" >&2; exit 2 ;;
esac

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
compose="$script_dir/compose.sh"
base_url="http://127.0.0.1:${GO_ADMIN_WEB_PORT:-8080}"

wait_status() {
  local url=$1 expected=$2 status=000
  for _ in $(seq 1 60); do
    status=$(curl -sS -o /dev/null -w '%{http_code}' "$url" || true)
    [[ "$status" == "$expected" ]] && return 0
    sleep 2
  done
  echo "$url returned $status, expected $expected" >&2
  return 1
}

wait_status "$base_url/health/live" 200
wait_status "$base_url/health/ready" 200
capabilities=$(curl -fsS "$base_url/api/v1/runtime/capabilities")
jq -e '.hostProfile == "server" and .desktop == false' <<<"$capabilities" >/dev/null

if [[ "$profile" == postgres ]]; then
  migration_count=$("$compose" postgres exec -T postgres psql -U go_admin_plus -d go_admin_plus -Atc 'SELECT COUNT(*) FROM goose_db_version;')
else
  migration_count=$("$compose" sqlite exec -T api-sqlite sqlite3 /var/lib/go-admin-plus/database.sqlite3 'SELECT COUNT(*) FROM goose_db_version;')
fi
[[ "$migration_count" =~ ^[1-9][0-9]*$ ]] || { echo "migration ledger is empty" >&2; exit 1; }

"$compose" "$profile" restart "$api_service" >/dev/null
wait_status "$base_url/health/ready" 200
if [[ "$profile" == postgres ]]; then
  restarted_count=$("$compose" postgres exec -T postgres psql -U go_admin_plus -d go_admin_plus -Atc 'SELECT COUNT(*) FROM goose_db_version;')
else
  restarted_count=$("$compose" sqlite exec -T api-sqlite sqlite3 /var/lib/go-admin-plus/database.sqlite3 'SELECT COUNT(*) FROM goose_db_version;')
fi
[[ "$restarted_count" == "$migration_count" ]] || { echo "migration state changed across restart" >&2; exit 1; }

for service in "$api_service" "$web_service"; do
  container=$("$compose" "$profile" ps -q "$service")
  [[ -n "$container" ]]
  [[ $(docker inspect -f '{{.HostConfig.ReadonlyRootfs}}' "$container") == true ]]
  [[ $(docker inspect -f '{{.HostConfig.Privileged}}' "$container") == false ]]
  runtime_user=$(docker inspect -f '{{.Config.User}}' "$container")
  [[ -n "$runtime_user" && "$runtime_user" != 0 && "$runtime_user" != root ]]
done

echo "GO_ADMIN_LINUX_COMPOSE_${profile^^}_PASS"
