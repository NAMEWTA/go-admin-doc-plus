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

inventory=$script_dir/legacy-governance-t21.txt
test -f "$inventory" || fail 'missing T-21 legacy governance inventory'
actual=$(mktemp "${TMPDIR:-/tmp}/go-admin-governance.XXXXXX")
expected=$(mktemp "${TMPDIR:-/tmp}/go-admin-governance.XXXXXX")
trap 'rm -f "$actual" "$expected"' EXIT HUP INT TERM

find "$backend_root" "$repo_root/go-admin-ui-plus" \
  \( -name node_modules -o -name dist -o -name coverage -o -name .git -o \
    -path '*/.husky/_' \) -prune -o \
  -type f \( -path '*/.github/*' -o -path '*/.husky/*' -o \
    -name .gitignore -o -name .gitattributes -o -name .editorconfig -o \
    -name .dockerignore -o -name Makefile -o -name 'Dockerfile*' -o \
    -name 'docker-compose*.yml' -o -name 'docker-compose*.yaml' \) -print |
  sed "s|^$repo_root/||" | sort >"$actual"
sed '/^[[:space:]]*#/d; /^[[:space:]]*$/d' "$inventory" | sort >"$expected"

if ! diff -u "$expected" "$actual"; then
  fail 'nested governance differs from the explicit T-21 contraction inventory'
fi

printf '%s\n' 'GOVERNANCE_CHECK_PASS'
