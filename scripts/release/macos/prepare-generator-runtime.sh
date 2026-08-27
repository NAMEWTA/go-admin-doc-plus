#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 1 ]] || { echo "usage: prepare-generator-runtime.sh <app>" >&2; exit 2; }
app=$1
[[ -d "$app/Contents/Resources" ]]
for command in curl git jq shasum tar; do command -v "$command" >/dev/null; done

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
identity="$repository/release/macos/identity.json"
generator="$app/Contents/Resources/generator"
[[ ! -e "$generator" ]] || { echo "generator runtime already exists" >&2; exit 1; }
staging=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/go-admin-macos-runtime.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT
mkdir -p -- "$generator/repository" "$generator/toolchains" "$generator/pnpm-store" "$generator/go-mod"

fetch() {
  local url=$1 expected=$2 output=$3
  curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$output" "$url"
  printf '%s  %s\n' "$expected" "$output" | shasum -a 256 -c - >/dev/null
}
archive_value() { jq -er ".generatorRuntime.archives.$1.$2" "$identity"; }

go_version=$(jq -er '.generatorRuntime.goVersion' "$identity")
node_version=$(jq -er '.generatorRuntime.nodeVersion' "$identity")
pnpm_version=$(jq -er '.generatorRuntime.pnpmVersion' "$identity")
for layout in darwin-amd64 darwin-arm64; do
  if [[ "$layout" == darwin-amd64 ]]; then go_key=goDarwinAmd64; node_key=nodeDarwinX64; node_arch=x64; else go_key=goDarwinArm64; node_key=nodeDarwinArm64; node_arch=arm64; fi
  toolchain="$generator/toolchains/$layout"
  mkdir -p -- "$toolchain/bin"
  go_archive="$staging/$(archive_value "$go_key" name)"
  node_archive="$staging/$(archive_value "$node_key" name)"
  fetch "https://go.dev/dl/$(basename -- "$go_archive")" "$(archive_value "$go_key" sha256)" "$go_archive"
  fetch "https://nodejs.org/dist/v$node_version/$(basename -- "$node_archive")" "$(archive_value "$node_key" sha256)" "$node_archive"
  tar -xzf "$go_archive" -C "$toolchain"
  tar -xzf "$node_archive" -C "$toolchain"
  mv "$toolchain/node-v$node_version-darwin-$node_arch" "$toolchain/node"
done

pnpm_archive="$staging/$(archive_value pnpm name)"
fetch "https://registry.npmjs.org/pnpm/-/pnpm-$pnpm_version.tgz" "$(archive_value pnpm sha256)" "$pnpm_archive"
mkdir -p -- "$staging/pnpm"
tar -xzf "$pnpm_archive" -C "$staging/pnpm"
mv "$staging/pnpm/package" "$generator/pnpm"
for layout in darwin-amd64 darwin-arm64; do
  toolchain="$generator/toolchains/$layout"
  cp "$script_dir/pnpm-offline.sh" "$toolchain/bin/pnpm"
  chmod 0755 "$toolchain/bin/pnpm"
done

git -C "$repository" archive --format=tar HEAD -- \
  .gitignore contracts/openapi scripts/contracts go-admin-plus \
  go-admin-plus-ui/.npmrc go-admin-plus-ui/package.json go-admin-plus-ui/pnpm-lock.yaml \
  go-admin-plus-ui/pnpm-workspace.yaml go-admin-plus-ui/apps go-admin-plus-ui/packages \
  go-admin-plus-ui/tests/shell | tar -xf - -C "$generator/repository"
git -C "$generator/repository" init --quiet
git -C "$generator/repository" add .gitignore contracts scripts go-admin-plus go-admin-plus-ui
if git -C "$generator/repository" remote get-url origin >/dev/null 2>&1; then
  git -C "$generator/repository" remote remove origin
fi

case "$(uname -m)" in
  arm64) host_layout=darwin-arm64 ;;
  x86_64) host_layout=darwin-amd64 ;;
  *) echo "unsupported macOS builder architecture" >&2; exit 1 ;;
esac
host_node="$generator/toolchains/$host_layout/node/bin/node"
host_pnpm="$generator/pnpm/bin/pnpm.cjs"
workspace="$generator/repository/go-admin-plus-ui"
workspace_config="$workspace/pnpm-workspace.yaml"
cp "$workspace_config" "$staging/pnpm-workspace.yaml"
"$host_node" "$host_pnpm" --dir "$workspace" config set --location=project --json \
  supportedArchitectures '{"os":["darwin"],"cpu":["x64","arm64"]}'
GOMODCACHE="$generator/go-mod" GOENV=off GOTOOLCHAIN=local \
  "$generator/toolchains/$host_layout/go/bin/go" -C "$generator/repository/go-admin-plus" mod download
"$host_node" "$host_pnpm" --store-dir "$generator/pnpm-store" --dir "$workspace" fetch --frozen-lockfile
"$host_node" "$host_pnpm" --store-dir "$generator/pnpm-store" --offline --dir "$workspace" install --frozen-lockfile --ignore-scripts
cp "$staging/pnpm-workspace.yaml" "$workspace_config"

git -C "$generator/repository" diff --exit-code -- . ':!go-admin-plus-ui/.npmrc'
test -z "$(git -C "$generator/repository" remote)"
chmod -R go-w "$generator"
echo "GO_ADMIN_MACOS_GENERATOR_RUNTIME_PASS"
