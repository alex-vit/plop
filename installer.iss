#ifndef AppVersion
  #define AppVersion "dev"
#endif

[Setup]
AppId={{F2E9BB76-0AF1-4F1B-A9F5-2DA8A9E80E70}
AppName=Plop
AppVersion={#AppVersion}
AppVerName=Plop {#AppVersion}
AppPublisher=alex-vit
AppPublisherURL=https://github.com/alex-vit/plop
DefaultDirName={localappdata}\Plop
DefaultGroupName=Plop
PrivilegesRequired=lowest
OutputDir=out
OutputBaseFilename=plop-setup
SetupIconFile=icon\icon.ico
UninstallDisplayIcon={app}\plop.exe
Compression=lzma2
SolidCompression=yes
AppMutex=PlopMutex
CloseApplications=yes
WizardStyle=modern

[Files]
Source: "out\plop.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Plop"; Filename: "{app}\plop.exe"
Name: "{group}\Uninstall Plop"; Filename: "{uninstallexe}"

[Tasks]
Name: "autostart"; Description: "Start with Windows"; GroupDescription: "Additional options:"

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "Plop"; ValueData: """{app}\plop.exe"""; Flags: uninsdeletevalue; Tasks: autostart

[UninstallDelete]
Type: filesandordirs; Name: "{app}"

[Run]
Filename: "{app}\plop.exe"; Description: "Launch Plop"; Flags: nowait postinstall skipifsilent
