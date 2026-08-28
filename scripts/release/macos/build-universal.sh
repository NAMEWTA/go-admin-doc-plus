#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 3 ]] || { echo "usage: build-universal.sh <version> <build-number> <output-app>" >&2; exit 2; }
version=$1
build_number=$2
output_app=$3
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]
[[ "$build_number" =~ ^[1-9][0-9]*$ ]]
[[ "$(uname -s)" == Darwin ]]
command -v lipo >/dev/null
command -v rustup >/dev/null
command -v pnpm >/dev/null
[[ ! -e "$output_app" ]] || { echo "output app already exists" >&2; exit 1; }

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
tauri_root="$repository/go-admin-plus-ui/apps/admin-desktop/src-tauri"
binaries="$tauri_root/binaries"

for target in aarch64-apple-darwin x86_64-apple-darwin; do
  rustup target list --installed | grep -Fx "$target" >/dev/null
  node "$repository/release/shared/sidecar/build.mjs" --target "$target"
done
lipo -create \
  "$binaries/go-admin-sidecar-aarch64-apple-darwin" \
  "$binaries/go-admin-sidecar-x86_64-apple-darwin" \
  -output "$binaries/go-admin-sidecar-universal-apple-darwin"
test "$(lipo -archs "$binaries/go-admin-sidecar-universal-apple-darwin")" = "x86_64 arm64"

pnpm --dir "$repository/go-admin-plus-ui" --filter @go-admin-plus/admin-desktop build
pnpm --dir "$repository/go-admin-plus-ui" --filter @go-admin-plus/admin-desktop exec tauri build \
  --target universal-apple-darwin \
  --features custom-protocol \
  --bundles app \
  --no-sign \
  --config "{\"version\":\"$version\"}"

built_app="$tauri_root/target/universal-apple-darwin/release/bundle/macos/Go Admin Plus.app"
test -d "$built_app"
mkdir -p -- "$(dirname -- "$output_app")"
ditto "$built_app" "$output_app"
"$script_dir/prepare-app.sh" "$output_app" "$version" "$build_number"
echo "GO_ADMIN_MACOS_UNIVERSAL_BUILD_PASS"
