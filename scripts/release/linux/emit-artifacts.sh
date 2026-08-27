#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 3 ]] || { echo "usage: emit-artifacts.sh <version> <source-sha> <output-directory>" >&2; exit 2; }
version=$1
source_sha=$2
output=$3
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]
[[ "$source_sha" =~ ^[0-9a-f]{40}$ ]]
command -v syft >/dev/null
command -v jq >/dev/null
[[ ! -e "$output" ]] || { echo "artifact output already exists" >&2; exit 1; }

server_image=${GO_ADMIN_SERVER_IMAGE:-go-admin-plus-server}
web_image=${GO_ADMIN_WEB_IMAGE:-go-admin-plus-web}
mkdir -p -- "$output/images" "$output/sbom" "$output/compose"
images_metadata=$(mktemp)
printf '[]\n' > "$images_metadata"
trap 'rm -f -- "$images_metadata" "$images_metadata.next"' EXIT

for architecture in amd64 arm64; do
  for component in server web; do
    if [[ "$component" == server ]]; then image=$server_image; else image=$web_image; fi
    reference="$image:$source_sha-$architecture"
    docker image inspect "$reference" >/dev/null
    image_id=$(docker image inspect -f '{{.Id}}' "$reference")
    [[ "$image_id" =~ ^sha256:[0-9a-f]{64}$ ]]
    docker save "$reference" | gzip -9 > "$output/images/go-admin-plus-$component-linux-$architecture.tar.gz"
    syft "$reference" -o "spdx-json=$output/sbom/go-admin-plus-$component-linux-$architecture.spdx.json"
    jq --arg component "$component" --arg platform "linux/$architecture" --arg reference "$reference" --arg imageId "$image_id" \
      '. + [{component:$component,platform:$platform,reference:$reference,imageId:$imageId}]' \
      "$images_metadata" > "$images_metadata.next"
    mv -f -- "$images_metadata.next" "$images_metadata"
  done
done

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
cp "$repository/deploy/compose/compose.yml" "$repository/deploy/compose/.env.example" "$output/compose/"
cp -R "$repository/deploy/compose/config" "$output/compose/"
cp "$repository/release/linux/README.md" "$repository/release/linux/identity.json" "$output/"
cp "$repository/scripts/release/linux/compose.sh" "$repository/scripts/release/linux/prepare-secrets.sh" "$repository/scripts/release/linux/verify-compose.sh" "$output/"

jq -n \
  --arg version "$version" \
  --arg source "$source_sha" \
  --arg server "$server_image" \
  --arg web "$web_image" \
  --slurpfile images "$images_metadata" \
  '{schemaVersion:1,product:"go-admin-plus",version:$version,sourceSha:$source,platforms:["linux/amd64","linux/arm64"],imageNames:{server:$server,web:$web},images:$images[0],profiles:["server-postgres","server-sqlite"],remotePublished:false,sbomFormat:"SPDX JSON"}' \
  > "$output/provenance.json"
(cd -- "$output" && find . -type f ! -name SHA256SUMS -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > SHA256SUMS)
(cd -- "$output" && sha256sum -c SHA256SUMS)
echo "GO_ADMIN_LINUX_SUPPLY_CHAIN_PASS"
