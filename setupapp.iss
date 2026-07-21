; ==========================================
; ModIn Installer
; Version 1.0
; ==========================================

#define AppName "ModIn"
#define AppVersion "1.0"
#define AppPublisher "ModIn Team"
#define AppExeName "ModIn.exe"
[Messages]


WelcomeLabel2=ModIn - manages Minecraft mods.\n\nRecommended: Install to LocalAppData or another writable folder.\nAvoid installing to Program Files because ModIn stores data inside the installation folder.

[Setup]
AppId=ModInTeam.ModIn
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}

DefaultDirName={localappdata}\ModIn
DefaultGroupName=ModIn

OutputDir=Output
OutputBaseFilename=ModInSetup

SetupIconFile=icon.ico
UninstallDisplayIcon={app}\ModIn.exe

Compression=lzma2
SolidCompression=yes
WizardStyle=modern

ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

PrivilegesRequired=lowest

DisableProgramGroupPage=yes

LicenseFile=LICENSE

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create Desktop Shortcut"; Flags: unchecked

[Dirs]
Name: "{app}\Modpacks"

[Files]
Source: "Build\ModIn.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\ModIn"; Filename: "{app}\ModIn.exe"
Name: "{autodesktop}\ModIn"; Filename: "{app}\ModIn.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\ModIn.exe"; Description: "Launch ModIn"; Flags: nowait postinstall skipifsilent