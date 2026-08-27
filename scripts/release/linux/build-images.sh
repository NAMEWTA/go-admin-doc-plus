#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 2 ]] || { echo "usage: build-images.sh <version> <source-sha>" >&2; exit 2; }
version=$1
source_sha=$2
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] || { echo "invalid release version" >&2; exit 2; }
[[ "$source_sha" =~ ^[0-9a-f]{40}$ ]] || { echo "source SHA must be lowercase and exact" >&2; exit 2; }

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
server_image=${GO_ADMIN_SERVER_IMAGE:-go-admin-plus-server}
web_image=${GO_ADMIN_WEB_IMAGE:-go-admin-plus-web}

for platform in linux/amd64 linux/arm64; do
  architecture=${platform#linux/}
  docker buildx build --load --platform "$platform" \
    --file "$repository/release/linux/Containerfile.server" \
    --build-arg "VERSION=$version" --build-arg "VCS_REF=$source_sha" \
    --tag "$server_image:$source_sha-$architecture" "$repository"
  docker buildx build --load --platform "$platform" \
    --file "$repository/release/linux/Containerfile.web" \
    --build-arg "VERSION=$version" --build-arg "VCS_REF=$source_sha" \
    --tag "$web_image:$source_sha-$architecture" "$repository"
done

echo "GO_ADMIN_LINUX_MULTIARCH_BUILD_PASS"
