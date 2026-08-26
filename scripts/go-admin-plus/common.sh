#!/bin/sh

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
backend_root=$repo_root/go-admin-plus
artifacts_root=${GO_ADMIN_ARTIFACTS_DIR:-$repo_root/.artifacts}

fail() {
  printf '%s\n' "go-admin-plus task: $*" >&2
  exit 1
}

require_tool() {
  command -v "$1" >/dev/null 2>&1 || fail "required tool is not installed: $1"
}

validate_profile() {
  case $1 in
    server-sqlite|server-postgres) ;;
    *) fail "unsupported profile: $1 (expected server-sqlite or server-postgres)" ;;
  esac
}

profile_config() {
  validate_profile "$1"
  case $1 in
    server-sqlite)
      printf '%s\n' "$backend_root/config/settings.sqlite.yml"
      ;;
    server-postgres)
      test -n "${GO_ADMIN_CONFIG_FILE:-}" ||
        fail 'server-postgres requires GO_ADMIN_CONFIG_FILE to name an explicit config file'
      test -f "$GO_ADMIN_CONFIG_FILE" || fail 'GO_ADMIN_CONFIG_FILE does not name a readable file'
      printf '%s\n' "$GO_ADMIN_CONFIG_FILE"
      ;;
  esac
}
