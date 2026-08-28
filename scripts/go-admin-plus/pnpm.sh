#!/bin/sh

required_pnpm_version=11.1.3

require_pnpm() {
  if command -v pnpm >/dev/null 2>&1; then
    installed_pnpm_version=$(pnpm --version 2>/dev/null) || fail 'pnpm version cannot be determined'
    test "$installed_pnpm_version" = "$required_pnpm_version" ||
      fail "pnpm $required_pnpm_version is required; found $installed_pnpm_version"
    return 0
  fi

  command -v corepack >/dev/null 2>&1 || fail 'required tool is not installed: pnpm or Corepack'
  pnpm_shim_root=$artifacts_root/tool-shims/corepack
  pnpm_shim=$pnpm_shim_root/pnpm
  pnpm_shim_tmp=$pnpm_shim_root/.pnpm.$$.tmp
  pnpm_cmd_tmp=$pnpm_shim_root/.pnpm.cmd.$$.tmp
  mkdir -p "$pnpm_shim_root"
  {
    printf '%s\n' '#!/bin/sh'
    printf '%s\n' "exec corepack pnpm@$required_pnpm_version \"\$@\""
  } >"$pnpm_shim_tmp"
  chmod 0755 "$pnpm_shim_tmp"
  mv -f "$pnpm_shim_tmp" "$pnpm_shim"
  {
    printf '%s\r\n' '@echo off'
    printf 'corepack pnpm@%s %%*\r\n' "$required_pnpm_version"
  } >"$pnpm_cmd_tmp"
  mv -f "$pnpm_cmd_tmp" "$pnpm_shim_root/pnpm.cmd"
  PATH=$pnpm_shim_root${PATH:+:$PATH}
  export PATH

  installed_pnpm_version=$(pnpm --version 2>/dev/null) || fail 'Corepack did not provide the pnpm shim'
  test "$installed_pnpm_version" = "$required_pnpm_version" ||
    fail "Corepack did not provide pnpm $required_pnpm_version; found $installed_pnpm_version"
}

run_pnpm() {
  require_pnpm
  pnpm "$@"
}

exec_pnpm() {
  require_pnpm
  exec pnpm "$@"
}
