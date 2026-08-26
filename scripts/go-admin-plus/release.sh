#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

action=${1:-}
version=${2:-}
test -n "$version" || fail 'release version is required'
require_tool node

case $action in
  preflight)
    root_sha=$(git -C "$repo_root" rev-parse HEAD)
    node "$repo_root/release/manifest/product-release.mjs" preflight \
      --version "$version" --root-ref "$root_sha"
    node "$repo_root/release/manifest/scan-compatibility.mjs"
    exec node --test "$repo_root/release/manifest/product-release.test.mjs"
    ;;
  dispatch)
    require_tool gh
    root_sha=$(git -C "$repo_root" rev-parse HEAD)
    remote_sha=$(git -C "$repo_root" ls-remote origin refs/heads/main | awk '{print $1}')
    test "$remote_sha" = "$root_sha" || fail 'origin/main must equal the local commit before dispatch'
    gh workflow run release-linux.yml --repo NAMEWTA/go-admin-plus --ref main \
      -f orchestration_ref="$root_sha" -f frontend_ref="$root_sha" -f release_version="$version"
    gh workflow run release-macos.yml --repo NAMEWTA/go-admin-plus --ref main \
      -f orchestration_ref="$root_sha" -f frontend_ref="$root_sha" -f release_version="$version" \
      -f release_mode=unsigned-self-use
    exec gh workflow run release-windows.yml --repo NAMEWTA/go-admin-plus --ref main \
      -f orchestration_ref="$root_sha" -f frontend_ref="$root_sha" -f release_version="$version" \
      -f release_mode=unsigned-self-use
    ;;
  collect)
    test "$#" -eq 8 || fail 'collect requires six run/artifact IDs after the version'
    exec node "$repo_root/release/manifest/product-release.mjs" collect \
      --version "$version" \
      --linux-run-id "$3" --linux-artifact-id "$4" \
      --macos-run-id "$5" --macos-artifact-id "$6" \
      --windows-run-id "$7" --windows-artifact-id "$8" \
      --output "$repo_root/product-manifest.json"
    ;;
  *)
    fail "unsupported release action: $action"
    ;;
esac
