import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import { execute, parseSidecarProcesses, sidecarProcesses } from './processes.mjs'

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
    '104 node test.mjs go-admin-sidecar-aarch64-apple-darwin',
    '105 /bin/sh -c /tmp/go-admin-sidecar-aarch64-apple-darwin',
    '106 /tmp/go-admin-sidecar-aarch64-apple-darwin-copy',
    '107 /tmp/go-admin-sidecar-aarch64-apple-darwin.exe',
    '108 /tmp/go-admin-sidecar-x86_64-pc-windows-msvc',
    'not-a-pid /tmp/go-admin-sidecar-aarch64-apple-darwin'
  ].join('\n'))
  assert.deepEqual([...processes], [101, 102, 103])
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
