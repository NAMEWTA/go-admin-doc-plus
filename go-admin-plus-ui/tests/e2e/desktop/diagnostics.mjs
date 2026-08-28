const nativeFailurePattern = /^(desktop native (?:startup|login|logout|request|identity) failed: [A-Za-z ]{1,96})$/gm

export const nativeFailureDiagnostic = output =>
  [...output.matchAll(nativeFailurePattern)].at(-1)?.[1]

export const nativePhaseFailure = (phase, output) => {
  const diagnostic = nativeFailureDiagnostic(output)
  return new Error(`desktop native ${phase} failed${diagnostic ? `: ${diagnostic}` : ''}`)
}
