#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
workspace_dir=$(CDPATH= cd -- "$script_dir/../.." && pwd)
compose_dir=${GO_ADMIN_COMPOSE_DIR:-"$workspace_dir/deploy/compose"}
runtime_dir="$compose_dir/runtime"
secret_dir="$runtime_dir/secrets"
postgres_secret="$secret_dir/postgres_password"
jwt_secret="$secret_dir/jwt_secret"
settings_file="$runtime_dir/settings.yml"

mkdir -p "$secret_dir"
umask 077
chmod 0700 "$runtime_dir" "$secret_dir"

generate_secret() {
  target=$1
  if [ ! -s "$target" ]; then
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n' > "$target"
  fi
}

generate_secret "$postgres_secret"
generate_secret "$jwt_secret"

postgres_password=$(tr -d '\r\n' < "$postgres_secret")
jwt_value=$(tr -d '\r\n' < "$jwt_secret")
case "$postgres_password" in
  *[!A-Za-z0-9_-]*|'') echo "postgres password must use only A-Z, a-z, 0-9, _ or -" >&2; exit 1 ;;
esac
case "$jwt_value" in
  *[!A-Za-z0-9_-]*|'') echo "JWT secret must use only A-Z, a-z, 0-9, _ or -" >&2; exit 1 ;;
esac

sed \
  -e "s/__POSTGRES_PASSWORD__/$postgres_password/g" \
  -e "s/__JWT_SECRET__/$jwt_value/g" \
  "$compose_dir/settings.template.yml" > "$settings_file"
chmod 0600 "$postgres_secret" "$jwt_secret"
# Local Compose secrets are bind mounts and retain host mode. The protected
# parent directory prevents host traversal while the non-root container can
# read the mounted settings file.
chmod 0444 "$settings_file"

echo "Prepared runtime configuration under $runtime_dir"
