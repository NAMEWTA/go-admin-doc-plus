#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
identity=$repo_root/release/windows/identity.json
installer=$repo_root/go-admin-plus/cmd/go-admin-desktop/build/windows/installer/project.nsi
workflow=$repo_root/go-admin-plus/.github/workflows/release-windows.yml

test -f "$identity"
test -f "$installer"
test -f "$workflow"

jq -e '
  .schema_version == 1 and
  .release_status == "approved" and
  .release_class == "unsigned-self-use" and
  .product_version == "0.1.0" and
  .install_scope == "user" and
  .webview2.deployment == "evergreen-standalone" and
  .webview2.architecture == "x64" and
  (.webview2.sha256 | test("^[0-9a-f]{64}$")) and
  .webview2.size_bytes > 200000000
' "$identity" >/dev/null

grep -q 'MicrosoftEdgeWebView2RuntimeInstallerX64.exe' "$installer"
grep -q 'ExecWait' "$installer"
grep -q 'wails.files' "$installer"
! grep -q 'MicrosoftEdgeWebview2Setup.exe' "$installer"

grep -q 'unsigned-self-use' "$workflow"
grep -q "'error'" "$workflow"
! grep -Eq "'webview2',[[:space:]]*'(download|embed|browser)'" "$workflow"
! grep -Eqi 'Set-MpPreference|DisableRealtimeMonitoring|SmartScreenEnabled.*Off|EnableSmartScreen.*false' \
  "$repo_root"/release/windows/* "$workflow"

echo "GO_ADMIN_WINDOWS_RELEASE_CONTRACT_PASS"
