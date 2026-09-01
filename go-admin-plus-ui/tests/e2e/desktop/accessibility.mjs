export const quoteAppleScript = value => `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`

export const clickButtonScript = (pid, name) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then error "native process unavailable"
tell (first process whose unix id is ${pid})
  set didClick to false
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if role of currentElement is "AXButton" and name of currentElement is ${quoteAppleScript(name)} then
        click currentElement
        set didClick to true
        exit repeat
      end if
    end try
  end repeat
end tell
if not didClick then error ${quoteAppleScript(`native button unavailable: ${name}`)}
return "clicked"
end tell`

export const pressButtonScript = (pid, name) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then error "native process unavailable"
tell (first process whose unix id is ${pid})
  set frontmost to true
  set didPress to false
  set navigationStart to missing value
  set targetElement to missing value
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if role of currentElement is "AXButton" and name of currentElement is ${quoteAppleScript(name)} and enabled of currentElement is true then
        set targetElement to contents of currentElement
      else if role of currentElement is "AXButton" and name of currentElement is "\u7528\u6237\u7ba1\u7406" and enabled of currentElement is true then
        set navigationStart to contents of currentElement
      end if
    end try
  end repeat
  if targetElement is missing value then error ${quoteAppleScript(`native button unavailable: ${name}`)}
  if navigationStart is missing value then error "native navigation start unavailable"
  try
    set focused of navigationStart to true
    repeat with traversalIndex from 1 to 11
      key code 48 using option down
    end repeat
    delay 0.2
  on error
    error "native navigation traversal unavailable"
  end try
  set refreshedElement to missing value
  set refreshedElementsToScan to entire contents of window 1
  repeat with refreshedCandidate in refreshedElementsToScan
    try
      if role of refreshedCandidate is "AXButton" and name of refreshedCandidate is ${quoteAppleScript(name)} and enabled of refreshedCandidate is true and focused of refreshedCandidate is true then
        set refreshedElement to contents of refreshedCandidate
        exit repeat
      end if
    end try
  end repeat
  if refreshedElement is missing value then error "native navigation traversal unavailable"
  try
    set buttonPosition to position of refreshedElement
    set buttonSize to size of refreshedElement
    set clickPoint to {(item 1 of buttonPosition) + ((item 1 of buttonSize) div 2), (item 2 of buttonPosition) + ((item 2 of buttonSize) div 2)}
    click at clickPoint
    set didPress to true
  on error
    error "native navigation action unavailable"
  end try
end tell
if not didPress then error "native navigation action unavailable"
return "pressed"
end tell`

export const buttonCurrentScript = (pid, name) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if role of currentElement is "AXButton" and name of currentElement is ${quoteAppleScript(name)} then
        if (value of attribute "AXARIACurrent" of currentElement as text) is "page" then return "true"
      end if
    end try
  end repeat
end tell
return "false"
end tell`

export const windowBusyScript = pid => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if value of attribute "AXBusy" of currentElement is true then return "true"
    end try
  end repeat
end tell
return "false"
end tell`

export const fillAndSubmitScript = (pid, fields, button) => {
  const actions = fields.map(({ name, role = 'AXTextField', value }, index) => `  set targetField${index} to missing value
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if role of currentElement is ${quoteAppleScript(role)} and name of currentElement is ${quoteAppleScript(name)} then
        set targetField${index} to contents of currentElement
        exit repeat
      end if
    end try
  end repeat
  if targetField${index} is missing value then error ${quoteAppleScript(`native field unavailable: ${name}`)}
  try
    set focused of targetField${index} to true
    delay 0.2
    keystroke "a" using command down
    keystroke ${quoteAppleScript(value)}
    delay 0.2
  on error
    error ${quoteAppleScript(`native field action unavailable: ${index}`)}
  end try`).join('\n')
  return `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then error "native process unavailable"
tell (first process whose unix id is ${pid})
  set frontmost to true
${actions}
  set submitControl to missing value
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if role of currentElement is "AXButton" and name of currentElement is ${quoteAppleScript(button)} then
        set submitControl to contents of currentElement
        exit repeat
      end if
    end try
  end repeat
  if submitControl is missing value then error ${quoteAppleScript(`native submit button unavailable: ${button}`)}
  set focused of submitControl to true
  delay 0.2
  key code 36
end tell
return "submitted"
end tell`
}

export const windowFrameScript = pid => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then error "native process unavailable"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then error "native window unavailable"
  set windowPosition to position of window 1
  set windowSize to size of window 1
  return ((item 1 of windowPosition) as text) & "," & ((item 2 of windowPosition) as text) & "," & ((item 1 of windowSize) as text) & "," & ((item 2 of windowSize) as text)
end tell
end tell`

export const windowContainsScript = (pid, value) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  set expectedValue to ${quoteAppleScript(value)}
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      if (name of currentElement as text) contains expectedValue then return "true"
    end try
  end repeat
end tell
return "false"
end tell`

export const windowValueScript = (pid, prefix) => `tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return ""
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return ""
  set elementsToScan to entire contents of window 1
  repeat with currentElement in elementsToScan
    try
      set observedName to name of currentElement as text
      if observedName starts with ${quoteAppleScript(prefix)} then return observedName
    end try
  end repeat
end tell
return ""
end tell`
