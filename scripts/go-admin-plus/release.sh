#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

action=${1:-}
version=${2:-}
require_tool node

case $action in
  preflight)
    test -n "$version" || fail 'release version is required'
    root_sha=$(git -C "$repo_root" rev-parse HEAD)
    node "$repo_root/release/manifest/product-release.mjs" preflight \
      --version "$version" --root-ref "$root_sha"
    node "$repo_root/release/manifest/scan-compatibility.mjs"
    exec node --test "$repo_root/release/manifest/product-release.test.mjs"
    ;;
  verify)
    node "$repo_root/scripts/quality/compatibility-zero.mjs"
    exec node --test \
      "$repo_root/release/manifest/product-release.test.mjs" \
      "$repo_root/release/shared/sidecar/build.test.mjs" \
      "$repo_root/scripts/release/linux/verify-policy.test.mjs" \
      "$repo_root/scripts/release/macos/verify-policy.test.mjs" \
      "$repo_root/scripts/release/windows/verify-policy.test.mjs"
    ;;
  collect)
    test -n "$version" || fail 'release version is required'
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
