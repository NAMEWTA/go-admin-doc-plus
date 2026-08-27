#!/bin/sh
set -eu
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
generator_root=$(cd -- "$script_dir/../../.." && pwd -P)
exec "$script_dir/../node/bin/node" "$generator_root/pnpm/bin/pnpm.cjs" \
  --store-dir "$generator_root/pnpm-store" --offline "$@"
