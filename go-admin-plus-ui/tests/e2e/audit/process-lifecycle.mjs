const hasExited = (child) => child.exitCode !== null || child.signalCode !== null

export const waitForExit = (child, timeoutMilliseconds) => new Promise((resolve) => {
  if (hasExited(child)) {
    resolve({ exited: true, code: child.exitCode, signal: child.signalCode })
    return
  }
  const onExit = (code, signal) => {
    clearTimeout(timer)
    resolve({ exited: true, code, signal })
  }
  const timer = setTimeout(() => {
    child.removeListener('exit', onExit)
    resolve({ exited: false, code: null, signal: null })
  }, timeoutMilliseconds)
  child.once('exit', onExit)
})

export const terminateChild = async (child, timeoutMilliseconds = 5_000) => {
  if (hasExited(child)) return { code: child.exitCode, signal: child.signalCode, forced: false }
  child.kill('SIGTERM')
  let result = await waitForExit(child, timeoutMilliseconds)
  if (result.exited) return { ...result, forced: false }
  child.kill('SIGKILL')
  result = await waitForExit(child, timeoutMilliseconds)
  if (!result.exited) throw new Error('child process did not exit after SIGKILL')
  return { ...result, forced: true }
}
