import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import {
  desktopNativeControlMarkers,
  verifyDesktopProductionAssets,
  verifyDesktopProductionFiles
} from '../../../apps/admin-desktop/scripts/verify-production.mjs'
import { desktopProductionArtifactPaths } from '../../../apps/admin-desktop/scripts/verify-build.mjs'
import { clickButtonScript, fillAndClickScript, quoteAppleScript, windowContainsScript, windowValueScript } from './accessibility.mjs'
import { nativeAccessibilityFailure, nativeFailureDiagnostic, nativePhaseFailure } from './diagnostics.mjs'
import { execute, parseSidecarProcesses, reapNewSidecars, sidecarProcesses } from './processes.mjs'

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..')

test('native runner Go fixture dependencies exist in the current backend', () => {
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  const fixtures = [...runner.matchAll(/'go', \['run', '([^']+)'/g)].map(match => match[1])
  assert.deepEqual([...new Set(fixtures)], ['./test/desktop/fixture'])
  assert.ok(existsSync(join(repositoryRoot, 'go-admin-plus', 'test/desktop/fixture/main.go')))
})

test('native runner accessibility labels match the current Session UI', () => {
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  const login = readFileSync(join(repositoryRoot, 'go-admin-plus-ui/packages/web-domains/iam/src/session/LoginPage.vue'), 'utf8')
  const account = readFileSync(join(repositoryRoot, 'go-admin-plus-ui/packages/web-domains/iam/src/session/AccountPage.vue'), 'utf8')
  const demo = readFileSync(join(repositoryRoot, 'go-admin-plus-ui/packages/web-domains/demo/src/DemoProductsPage.vue'), 'utf8')
  assert.match(login, /'\u767b\u5f55'/)
  assert.match(login, /aria-label="\u8d26\u53f7"/)
  assert.match(login, /aria-label="\u5bc6\u7801"/)
  for (const label of ['SKU', '\u540d\u79f0', '\u63cf\u8ff0', '\u4ef7\u683c\uff08\u5206\uff09']) assert.match(demo, new RegExp(`aria-label="${label}"`))
  assert.match(account, />\u9000\u51fa\u767b\u5f55<\/button>/)
  assert.match(runner, /\], '\u767b\u5f55'\)\)/)
  assert.match(runner, /clickButton\(app\.child\.pid, 'Administrator'\)[\s\S]*windowContains\(app\.child\.pid, '\u9000\u51fa\u767b\u5f55'\)/)
  assert.match(runner, /clickButtonScript\(pid, '\u9000\u51fa\u767b\u5f55'\)/)
  assert.match(runner, /poll\('native logout', \(\) => windowContains\(app\.child\.pid, '\u4f7f\u7528\u7ba1\u7406\u5458\u8d26\u53f7\u767b\u5f55\u63a7\u5236\u53f0'\), 90_000\)/)
  assert.doesNotMatch(runner, /name: '(?:Username|Password)'|, 'Sign in'\)\)|, 'Sign out'\)\)|windowContains\([^\n]+, '(?:Sign in|Sign out)'\)/)
})

test('production Desktop entry contains no native E2E controls', () => {
  const appRoot = join(repositoryRoot, 'go-admin-plus-ui/apps/admin-desktop')
  const app = readFileSync(join(appRoot, 'src/App.vue'), 'utf8')
  const main = readFileSync(join(appRoot, 'src/main.ts'), 'utf8')
  const vite = readFileSync(join(appRoot, 'vite.config.ts'), 'utf8')
  const manifest = JSON.parse(readFileSync(join(appRoot, 'package.json'), 'utf8'))
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  assert.doesNotMatch(app, /VITE_GO_ADMIN_NATIVE_E2E|__desktop\/test-control|native-e2e|window\.confirm\s*=|E2E scope self/)
  assert.match(app, /<ProductWorkspace/)
  assert.match(main, /import App from '@desktop-entry'/)
  assert.match(vite, /mode === 'native-e2e' \? '\.\/src\/native-e2e\/App\.vue' : '\.\/src\/App\.vue'/)
  assert.match(runner, /'--mode', 'native-e2e'/)
  assert.match(runner, /verifyDesktopProductionFiles\(\[sidecarBinary, hostBinary\]\)/)
  assert.doesNotMatch(runner, /fileContains/)
  assert.match(manifest.scripts.build, /vite build[^&]+&& node scripts\/verify-production\.mjs/)
})

test('production Desktop asset verifier rejects native E2E bytes', async t => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-desktop-production-'))
  t.after(() => rmSync(root, { recursive: true, force: true }))
  mkdirSync(join(root, 'assets'))
  writeFileSync(join(root, 'index.html'), '<main>Go Admin Plus</main>')
  writeFileSync(join(root, 'assets/app.css'), '.product-shell{display:grid}')
  await assert.doesNotReject(verifyDesktopProductionAssets(root))
  assert.deepEqual(desktopNativeControlMarkers, [
    '/__desktop/test-control',
    'native-e2e',
    'VITE_GO_ADMIN_NATIVE_E2E',
    'GO_ADMIN_DESKTOP_NATIVE_E2E',
    'GO_ADMIN_DESKTOP_E2E_',
    'desktop_native_e2e',
    'E2E authenticated boundary verified',
    'E2E unauthenticated boundary verified',
    'E2E boundary blocked:',
    'E2E self scope enforced',
    'E2E all scope restored',
    'E2E authorization denied',
    'E2E control failed:',
    'E2E scope self',
    'E2E scope all',
    'E2E permissions off',
    'E2E permissions on',
    'E2E revoke session',
    'E2E-FOREIGN',
    'E2E-001',
    'native E2E credential identity'
  ])
  for (const marker of desktopNativeControlMarkers) {
    writeFileSync(join(root, 'assets/app.css'), `.product-shell{content:${JSON.stringify(marker)}}`)
    await assert.rejects(verifyDesktopProductionAssets(root), new RegExp(`native test control: ${marker.replaceAll(/[.*+?^${}()|[\]\\]/g, '\\$&')}`))
  }
  writeFileSync(join(root, 'assets/app.css'), '.product-shell{display:grid}')
  await assert.doesNotReject(verifyDesktopProductionFiles([join(root, 'index.html'), join(root, 'assets/app.css')]))
  writeFileSync(join(root, 'assets/app.css'), '.product-shell{content:"E2E permissions on"}')
  await assert.rejects(verifyDesktopProductionFiles([join(root, 'assets/app.css')]), /native test control: E2E permissions on/)
})

test('production Desktop artifact paths are exact for each supported host', () => {
  const arm = desktopProductionArtifactPaths('/repository', 'darwin', 'arm64')
  assert.equal(arm.sidecar, '/repository/go-admin-plus-ui/apps/admin-desktop/src-tauri/binaries/go-admin-sidecar-aarch64-apple-darwin')
  assert.equal(arm.host, '/repository/go-admin-plus-ui/apps/admin-desktop/src-tauri/target/release/go-admin-plus-desktop')
  const windows = desktopProductionArtifactPaths('/repository', 'win32', 'x64')
  assert.equal(windows.sidecar, '/repository/go-admin-plus-ui/apps/admin-desktop/src-tauri/binaries/go-admin-sidecar-x86_64-pc-windows-msvc.exe')
  assert.equal(windows.host, '/repository/go-admin-plus-ui/apps/admin-desktop/src-tauri/target/release/go-admin-plus-desktop.exe')
  assert.throws(() => desktopProductionArtifactPaths('/repository', 'linux', 'x64'), /unsupported desktop host target/)
})

test('native runner stops polling after an early host exit', () => {
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  assert.match(runner, /const windowContains = async \(pid, value\) => \{\n  if \(!processIsAlive\(pid\)\) throw new Error\('native host exited before UI observation'\)/)
  assert.match(runner, /nativePhaseFailure\(phase, app\?\.output\(\) \?\? ''\)/)
})

test('native diagnostics preserve the failing phase and ignore state transitions', () => {
  const output = [
    'desktop native identity state: vault empty',
    'desktop native identity state: unauthenticated',
  ].join('\n')
  assert.equal(nativeFailureDiagnostic(output), undefined)
  assert.equal(nativePhaseFailure('session-revocation', output).message, 'desktop native session-revocation failed')

  const failed = `${output}\ndesktop native login failed: desktop login rejected\n`
  assert.equal(nativeFailureDiagnostic(failed), 'desktop native login failed: desktop login rejected')
  assert.equal(
    nativePhaseFailure('login-workspace', failed).message,
    'desktop native login-workspace failed: desktop native login failed: desktop login rejected'
  )
})

test('native accessibility diagnostics expose only a fixed unavailable category', () => {
  assert.equal(nativeAccessibilityFailure('native field unavailable: 密码 (-2700)'), 'desktop native accessibility field unavailable')
  assert.equal(nativeAccessibilityFailure('native field action unavailable: 2 (-1719)'), 'desktop native accessibility field-action-2 unavailable')
  assert.equal(nativeAccessibilityFailure('native submit button unavailable: 登录 (-2700)'), 'desktop native accessibility submit-button unavailable')
  assert.equal(nativeAccessibilityFailure('arbitrary AppleScript failure'), undefined)
})

test('native runner verifies SQLite only while the sidecar is stopped', () => {
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  const stopped = runner.indexOf('await assertNoNewSidecars(sidecarBaseline)', runner.indexOf("phase = 'product-update'"))
  const verified = runner.indexOf("phase = 'persistence-verification'", stopped)
  const restarted = runner.indexOf("phase = 'stronghold-restart'", verified)
  assert.ok(stopped > 0 && stopped < verified && verified < restarted)
})

test('native delete uses the uniquely named product action and a test-only confirmation port', () => {
  const runner = readFileSync(new URL('./run.mjs', import.meta.url), 'utf8')
  const app = readFileSync(join(repositoryRoot, 'go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue'), 'utf8')
  const page = readFileSync(join(repositoryRoot, 'go-admin-plus-ui/packages/web-domains/demo/src/DemoProductsPage.vue'), 'utf8')
  assert.match(runner, /clickButton\(pid, '\u5220\u9664 E2E-001'\)/)
  assert.match(runner, /phase = 'product-create'[\s\S]*clickButton\(app\.child\.pid, '\u65b0\u589e'\)[\s\S]*windowContains\(app\.child\.pid, '\u65b0\u589e\u4ea7\u54c1'\)/)
  assert.match(app, /nativeConfirm = window\.confirm\n  window\.confirm = \(\) => true/)
  assert.match(page, /:aria-label="`\u5220\u9664 \$\{product\.sku\}`"/)
  assert.doesNotMatch(runner, /first button whose name is "Delete"/)
  assert.doesNotMatch(runner, /tell sheet 1 of window 1 to click button "OK"/)
  assert.match(runner, /desktop native \$\{phase\} failed: \$\{error\.message\}/)
})

test('accessibility query walks the native flattened element collection', () => {
  assert.equal(quoteAppleScript('a\\"b'), '"a\\\\\\"b"')
  const script = windowContainsScript(42, '产品搜索')
  assert.match(script, /set elementsToScan to entire contents of window 1/)
  assert.match(script, /repeat with currentElement in elementsToScan/)
  assert.match(script, /if \(name of currentElement as text\) contains expectedValue/)
  assert.doesNotMatch(script, /every UI element of entire contents/)
  assert.match(clickButtonScript(42, '保存'), /role of currentElement is "AXButton" and name of currentElement is "\u4fdd\u5b58"/)
  const form = fillAndClickScript(42, [{ name: '名称', value: 'Native product' }], '保存')
  assert.match(form, /name of currentElement is "\u540d\u79f0"/)
  assert.match(form, /keystroke "Native product"/)
  assert.match(form, /name of currentElement is "\u4fdd\u5b58"/)
  assert.match(windowValueScript(42, 'E2E boundary blocked:'), /observedName starts with "E2E boundary blocked:"/)
})

test('native runner is a default skip with no environment prerequisites', () => {
  const result = spawnSync(process.execPath, [fileURLToPath(new URL('./run.mjs', import.meta.url))], {
    env: { PATH: process.env.PATH ?? '' }, encoding: 'utf8'
  })
  assert.equal(result.status, 0)
  assert.deepEqual(JSON.parse(result.stdout), {
    state: 'skipped',
    reason: 'GO_ADMIN_DESKTOP_NATIVE_E2E is not enabled'
  })
  assert.equal(result.stderr, '')
})

test('controlled exit one represents an empty process query', async () => {
  const output = await execute(process.execPath, ['-e', 'process.exit(1)'], { allowedExitCodes: [0, 1] })
  assert.equal(output, '')
  const processes = await sidecarProcesses(async (command, args, options) => {
    assert.equal(command, '/usr/bin/pgrep')
    assert.deepEqual(args, ['-lf', 'go-admin-sidecar'])
    assert.deepEqual(options.allowedExitCodes, [0, 1])
    return ''
  })
  assert.deepEqual([...processes], [])
})

test('process parser accepts only an exact approved sidecar executable', () => {
  const processes = parseSidecarProcesses([
    '101 /tmp/go-admin-sidecar-aarch64-apple-darwin --desktop',
    '102 /tmp/go-admin-sidecar-x86_64-apple-darwin',
    '103 /tmp/go-admin-sidecar-x86_64-pc-windows-msvc.exe',
    '109 /tmp/go-admin-sidecar',
    '110 /tmp/go-admin-sidecar.exe',
    '104 node test.mjs go-admin-sidecar-aarch64-apple-darwin',
    '105 /bin/sh -c /tmp/go-admin-sidecar-aarch64-apple-darwin',
    '106 /tmp/go-admin-sidecar-aarch64-apple-darwin-copy',
    '107 /tmp/go-admin-sidecar-aarch64-apple-darwin.exe',
    '108 /tmp/go-admin-sidecar-x86_64-pc-windows-msvc',
    'not-a-pid /tmp/go-admin-sidecar-aarch64-apple-darwin'
  ].join('\n'))
  assert.deepEqual([...processes], [101, 102, 103, 109, 110])
})

test('bounded command failures wait for killed process pipes to close', async () => {
  await assert.rejects(
    execute(process.execPath, ['-e', 'process.stdout.write("x".repeat(32768));setInterval(()=>{},1000)']),
    /output exceeded/
  )
  await assert.rejects(
    execute(process.execPath, ['-e', 'setInterval(()=>{},1000)'], { timeout: 10 }),
    /timed out/
  )
})

test('cleanup signals only exact sidecars outside the baseline and verifies reaping', async () => {
  const baseline = new Set([100])
  let current = new Set([100, 200, 300])
  const signals = []
  const leaked = await reapNewSidecars(baseline, {
    query: async () => new Set(current),
    signal(pid, name) {
      signals.push([pid, name])
      if (pid === 200 && name === 'SIGTERM') current.delete(pid)
      if (pid === 300 && name === 'SIGKILL') current.delete(pid)
    },
    pause: async () => {},
    timeout: 0
  })
  assert.deepEqual([...leaked], [200, 300])
  assert.deepEqual(signals, [[200, 'SIGTERM'], [300, 'SIGTERM'], [300, 'SIGKILL']])
  assert.deepEqual([...current], [100])
})
