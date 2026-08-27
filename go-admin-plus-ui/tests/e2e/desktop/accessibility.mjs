export const quoteAppleScript = value => `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`

export const windowContainsScript = (pid, value) => `on collectionContains(itemsToScan, expectedValue)
repeat with currentItem in itemsToScan
  set rawItem to contents of currentItem
  if class of rawItem is list then
    if my collectionContains(rawItem, expectedValue) then return true
  else if rawItem is not missing value then
    try
      if (rawItem as text) contains expectedValue then return true
    end try
  end if
end repeat
return false
end collectionContains

tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  set expectedValue to ${quoteAppleScript(value)}
  set elementNames to name of every UI element of entire contents of window 1
  if my collectionContains(elementNames, expectedValue) then return "true"
end tell
return "false"
end tell`
