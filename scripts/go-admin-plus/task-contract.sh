#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
taskfile=$repo_root/Taskfile.yml

fail() {
  printf '%s\n' "task contract: $*" >&2
  exit 1
}

for task_name in dev build test lint generate migrate package release; do
  grep -Eq "^  ${task_name}:$" "$taskfile" || fail "missing public task: $task_name"
done

grep -Eq '^  governance:check:$' "$taskfile" || fail 'missing governance:check task'
grep -Eq '^  task:contract:$' "$taskfile" || fail 'missing task:contract task'

for task_name in dev build test lint generate migrate package release; do
  awk -v task_name="$task_name" '
    $0 == "  " task_name ":" { in_task = 1; next }
    in_task && /^  [^ ].*:$/ { exit }
    in_task && /scripts\/(go-admin-plus|go-admin-plus-ui)\/[^[:space:]]+\.sh/ { found = 1 }
    END { exit found ? 0 : 1 }
  ' "$taskfile" || fail "$task_name does not delegate to a managed script"
done

managed_scripts=$(grep -Eo 'scripts/(go-admin-plus|go-admin-plus-ui)/[a-z0-9-]+\.sh' "$taskfile" | sort -u)
test -n "$managed_scripts" || fail 'Taskfile has no managed script references'
for relative_script in $managed_scripts; do
  test -f "$repo_root/$relative_script" || fail "missing managed script: $relative_script"
  test -x "$repo_root/$relative_script" || fail "managed script is not executable: $relative_script"
done

hook=$repo_root/.husky/pre-commit
test -f "$hook" || fail 'missing root pre-commit hook'
grep -Eq '(^|[[:space:]])task[[:space:]]+lint:staged([[:space:]]|$)' "$hook" ||
  fail 'pre-commit hook does not delegate to task lint:staged'

if ( . "$repo_root/scripts/go-admin-plus/common.sh"; require_tool go-admin-contract-missing-tool ) \
  >/dev/null 2>&1; then
  fail 'managed scripts do not propagate a missing-tool failure'
fi

hook_status=0
PATH=/nonexistent /bin/sh "$hook" >/dev/null 2>&1 || hook_status=$?
test "$hook_status" -ne 0 || fail 'pre-commit hook hides a missing Task CLI failure'

if task_command=$(command -v task 2>/dev/null); then
  for task_name in dev build test lint generate migrate package; do
    "$task_command" --dry "$task_name" >/dev/null ||
      fail "$task_name cannot resolve from the repository root with default inputs"
  done
  "$task_command" --dry release VERSION=0.0.0-contract >/dev/null ||
    fail 'release cannot resolve from the repository root with an explicit version'

  unknown_status=0
  "$task_command" __contract_unknown_task__ >/dev/null 2>&1 || unknown_status=$?
  test "$unknown_status" -ne 0 || fail 'unknown root tasks do not fail'

  child_status=0
  "$task_command" task:failure-probe >/dev/null 2>&1 || child_status=$?
  test "$child_status" -ne 0 || fail 'managed child failures do not propagate through Task'
fi

printf '%s\n' 'TASK_CONTRACT_PASS'
