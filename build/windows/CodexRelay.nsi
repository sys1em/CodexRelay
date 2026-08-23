Unicode true

!ifndef APP_NAME
  !define APP_NAME "CodexRelay"
!endif
!ifndef APP_VERSION
  !define APP_VERSION "0.0.0"
!endif
!ifndef APP_FILE_VERSION
  !define APP_FILE_VERSION "0.0.0.0"
!endif
!ifndef APP_ARCH
  !define APP_ARCH "amd64"
!endif
!ifndef APP_EXE
  !error "APP_EXE is required"
!endif
!ifndef APP_OUTPUT
  !define APP_OUTPUT "CodexRelay-${APP_VERSION}-${APP_ARCH}-setup.exe"
!endif
!ifndef APP_ICON
  !define APP_ICON "app.ico"
!endif
!ifndef APP_WEBVIEW2
  !error "APP_WEBVIEW2 is required"
!endif

Name "${APP_NAME}"
OutFile "${APP_OUTPUT}"
Icon "${APP_ICON}"
InstallDir "$LOCALAPPDATA\Programs\${APP_NAME}"
RequestExecutionLevel user
SetCompressor /SOLID lzma
ShowInstDetails show
ShowUninstDetails show

VIAddVersionKey "ProductName" "${APP_NAME}"
VIAddVersionKey "ProductVersion" "${APP_VERSION}"
VIAddVersionKey "FileDescription" "Codex API relay desktop application"
VIAddVersionKey "LegalCopyright" "Copyright (c) xxloocee"
VIProductVersion "${APP_FILE_VERSION}"
VIFileVersion "${APP_FILE_VERSION}"

Section "Install"
  SetOutPath "$INSTDIR"
  File /oname=CodexRelay.exe "${APP_EXE}"
  File /oname=MicrosoftEdgeWebview2Setup.exe "${APP_WEBVIEW2}"
  ExecWait '"$INSTDIR\MicrosoftEdgeWebview2Setup.exe" /silent /install'
  Delete "$INSTDIR\MicrosoftEdgeWebview2Setup.exe"
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  CreateDirectory "$SMPROGRAMS\${APP_NAME}"
  CreateShortCut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\CodexRelay.exe"
  CreateShortCut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\CodexRelay.exe"

  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayVersion" "${APP_VERSION}"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "Publisher" "xxloocee"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$DESKTOP\${APP_NAME}.lnk"
  Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
  RMDir "$SMPROGRAMS\${APP_NAME}"
  Delete "$INSTDIR\CodexRelay.exe"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
  DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
SectionEnd
