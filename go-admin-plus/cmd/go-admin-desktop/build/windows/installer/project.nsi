Unicode true

!include "wails_tools.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion "${INFO_PRODUCTVERSION}.0"
VIAddVersionKey "CompanyName" "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion" "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion" "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright" "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName" "${INFO_PRODUCTNAME}"

ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\go-admin-plus-${INFO_PRODUCTVERSION}-windows-amd64-unsigned-self-use-setup.exe"

!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    !error "Go Admin Plus Phase 1 requires a user-scope installer"
  !endif
!else
  !error "Go Admin Plus Phase 1 requires WAILS_INSTALL_SCOPE=user"
!endif

ShowInstDetails show

Function .onInit
  !insertmacro wails.checkArchitecture
FunctionEnd

Section
  !insertmacro wails.setShellContext

  SetRegView 64
  ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 != ""
    Goto webview2_ready
  ${EndIf}
  ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 != ""
    Goto webview2_ready
  ${EndIf}

  SetDetailsPrint both
  DetailPrint "Installing embedded Microsoft WebView2 Runtime"
  SetDetailsPrint listonly
  InitPluginsDir
  CreateDirectory "$pluginsdir\go-admin-webview2"
  SetOutPath "$pluginsdir\go-admin-webview2"
  File "/oname=MicrosoftEdgeWebView2RuntimeInstallerX64.exe" "tmp\MicrosoftEdgeWebView2RuntimeInstallerX64.exe"
  ExecWait '"$pluginsdir\go-admin-webview2\MicrosoftEdgeWebView2RuntimeInstallerX64.exe" /silent /install' $0
  ${If} $0 != 0
    ${If} $0 != 3010
      SetDetailsPrint both
      DetailPrint "Microsoft WebView2 Runtime installation failed with exit code $0"
      SetErrorLevel 70
      Abort
    ${EndIf}
  ${EndIf}

  webview2_ready:
  SetOutPath "$INSTDIR"
  !insertmacro wails.files

  CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
  CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
  !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
  !insertmacro wails.setShellContext
  Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
  Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
  !insertmacro wails.deleteUninstaller
  RMDir /r "$INSTDIR"
SectionEnd
