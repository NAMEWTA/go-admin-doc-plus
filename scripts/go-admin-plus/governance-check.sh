#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

required_files='
.gitignore
.gitattributes
.editorconfig
Taskfile.yml
.github/ISSUE_TEMPLATE/bug-report.yml
.github/ISSUE_TEMPLATE/feature-request.yml
.github/PULL_REQUEST_TEMPLATE.md
.husky/pre-commit
deploy/README.md
release/README.md
database/README.md
'

for relative_path in $required_files; do
  test -f "$repo_root/$relative_path" || fail "missing root-owned governance asset: $relative_path"
done

for relative_path in scripts/go-admin-plus scripts/go-admin-plus-ui; do
  test -d "$repo_root/$relative_path" || fail "missing root-owned script directory: $relative_path"
done

actual=$(mktemp "${TMPDIR:-/tmp}/go-admin-governance.XXXXXX")
trap 'rm -f "$actual"' EXIT HUP INT TERM

find "$backend_root" "$repo_root/go-admin-plus-ui" \
  \( -name node_modules -o -name dist -o -name target -o -name coverage -o -name .git -o \
    -path '*/.husky/_' \) -prune -o \
  -type f \( -path '*/.github/*' -o -path '*/.husky/*' -o \
    -name .gitignore -o -name .gitattributes -o -name .editorconfig -o \
    -name .dockerignore -o -name Makefile -o -name 'Dockerfile*' -o \
    -name 'docker-compose*.yml' -o -name 'docker-compose*.yaml' \) -print >"$actual"

if test -s "$actual"; then
  sed "s|^$repo_root/||" "$actual" >&2
  fail 'nested governance assets must be owned by the repository root'
fi

printf '%s\n' 'GOVERNANCE_CHECK_PASS'
