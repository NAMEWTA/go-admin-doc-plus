#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 6 ]] || { echo "usage: emit-artifacts.sh <version> <source-sha> <app> <dmg> <notary-json> <output>" >&2; exit 2; }
version=$1
source_sha=$2
app=$3
dmg=$4
notary=$5
output=$6
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]
[[ "$source_sha" =~ ^[0-9a-f]{40}$ ]]
[[ -d "$app" && -f "$dmg" && -f "$notary" && ! -e "$output" ]]
command -v jq >/dev/null
command -v syft >/dev/null
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
mkdir -p -- "$output/sbom"
ditto "$app" "$output/Go Admin Plus.app"
cp "$dmg" "$output/"
cp "$notary" "$output/notary.json"
cp "$repository/release/macos/identity.json" "$repository/release/macos/INSTALL.md" "$output/"
syft "dir:$app" -o "spdx-json=$output/sbom/go-admin-plus-macos-universal.spdx.json"
team=$(codesign -dvv "$app" 2>&1 | sed -n 's/^TeamIdentifier=//p')
jq -n --arg version "$version" --arg sourceSha "$source_sha" --arg teamIdentifier "$team" \
  '{schemaVersion:1,product:"go-admin-plus",version:$version,sourceSha:$sourceSha,platform:"macos-universal",architectures:["x86_64","arm64"],releaseClass:"signed-production",signed:true,notarized:true,teamIdentifier:$teamIdentifier,remotePublished:false,sbomFormat:"SPDX JSON"}' \
  > "$output/provenance.json"
(cd -- "$output" && find . -type f ! -name SHA256SUMS -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 > SHA256SUMS)
(cd -- "$output" && shasum -a 256 -c SHA256SUMS)
echo "GO_ADMIN_MACOS_SUPPLY_CHAIN_PASS"
