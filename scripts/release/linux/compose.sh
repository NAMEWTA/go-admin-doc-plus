#!/usr/bin/env bash
set -euo pipefail

[[ $# -ge 2 ]] || { echo "usage: compose.sh <postgres|sqlite> <compose arguments...>" >&2; exit 2; }
profile=$1
shift
case "$profile" in
  postgres|sqlite) ;;
  *) echo "unsupported Compose profile" >&2; exit 2 ;;
esac

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
compose_dir=${GO_ADMIN_COMPOSE_DIR:-"$repository/deploy/compose"}
environment_file=${GO_ADMIN_COMPOSE_ENV:-"$compose_dir/.env"}
[[ -f "$environment_file" ]] || environment_file="$compose_dir/.env.example"

arguments=(docker compose --env-file "$environment_file" -f "$compose_dir/compose.yml")
if [[ ${GO_ADMIN_COMPOSE_BUILD:-0} == 1 ]]; then
  arguments+=(-f "$compose_dir/compose.build.yml")
fi
if [[ -n ${GO_ADMIN_COMPOSE_PROJECT:-} ]]; then
  [[ "$GO_ADMIN_COMPOSE_PROJECT" =~ ^[a-z0-9][a-z0-9_-]{2,62}$ ]] || {
    echo "invalid Compose project name" >&2
    exit 2
  }
  arguments+=(-p "$GO_ADMIN_COMPOSE_PROJECT")
fi
arguments+=(--profile "$profile")
exec "${arguments[@]}" "$@"
