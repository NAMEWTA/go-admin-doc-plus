#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
compose_dir=${GO_ADMIN_COMPOSE_DIR:-"$repository/deploy/compose"}
secret_dir="$compose_dir/runtime/secrets"
password_file="$secret_dir/postgres_password"
dsn_file="$secret_dir/database_dsn"

mkdir -p -- "$secret_dir"
chmod 0700 "$compose_dir/runtime" "$secret_dir"
umask 077

if [[ ! -s "$password_file" ]]; then
  temporary=$(mktemp "$secret_dir/.postgres-password.XXXXXX")
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n' > "$temporary"
  mv -f -- "$temporary" "$password_file"
fi

password=$(tr -d '\r\n' < "$password_file")
[[ "$password" =~ ^[0-9a-f]{64}$ ]] || {
  echo "postgres credential file has an invalid format" >&2
  exit 1
}

temporary=$(mktemp "$secret_dir/.database-dsn.XXXXXX")
printf 'postgres://go_admin_plus:%s@postgres:5432/go_admin_plus?sslmode=disable\n' "$password" > "$temporary"
mv -f -- "$temporary" "$dsn_file"
chmod 0600 "$password_file" "$dsn_file"
unset password
echo "Compose secret references are ready"
