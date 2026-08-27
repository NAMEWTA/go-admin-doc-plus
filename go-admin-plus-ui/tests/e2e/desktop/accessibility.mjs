export const quoteAppleScript = value => `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`

export const windowContainsScript = (pid, value) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  set expectedValue to ${quoteAppleScript(value)}
  set allElements to entire contents of window 1
  repeat with currentElement in allElements
    try
      if ((name of currentElement) as text) contains expectedValue then return "true"
    end try
    try
      if ((value of currentElement) as text) contains expectedValue then return "true"
    end try
    try
      if ((description of currentElement) as text) contains expectedValue then return "true"
    end try
  end repeat
end tell
return "false"
end tell`
