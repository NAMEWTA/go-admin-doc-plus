const nativeFailurePattern = /^(desktop native (?:startup|login|logout|request|identity) failed: [A-Za-z ]{1,96})$/gm
const accessibilityFailurePattern = /native (submit button|process|button|field) unavailable/

export const nativeFailureDiagnostic = output =>
  [...output.matchAll(nativeFailurePattern)].at(-1)?.[1]

export const nativeAccessibilityFailure = output => {
  const category = output.match(accessibilityFailurePattern)?.[1]
  return category ? `desktop native accessibility ${category.replace(' ', '-')} unavailable` : undefined
}

export const nativePhaseFailure = (phase, output) => {
  const diagnostic = nativeFailureDiagnostic(output)
  return new Error(`desktop native ${phase} failed${diagnostic ? `: ${diagnostic}` : ''}`)
}
