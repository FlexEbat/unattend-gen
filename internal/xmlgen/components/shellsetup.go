package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	shellSetupName = "Microsoft-Windows-Shell-Setup"
	adminUsername  = "Administrator"
)

// disableWindowsUpdateCommand and disableUACCommand are the registry edits
// for the corresponding SystemTweaks flags. They run in the specialize pass.
const (
	disableWindowsUpdateCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU" /v NoAutoUpdate /t REG_DWORD /d 1 /f`
	disableUACCommand           = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v EnableLUA /t REG_DWORD /d 0 /f`
)

// ShellSetupSpecialize is the Microsoft-Windows-Shell-Setup component in the
// specialize pass, carrying the computer name and registry-based tweaks.
type ShellSetupSpecialize struct {
	XMLName               xml.Name        `xml:"component"`
	Name                  string          `xml:"name,attr"`
	ProcessorArchitecture string          `xml:"processorArchitecture,attr"`
	PublicKeyToken        string          `xml:"publicKeyToken,attr"`
	Language              string          `xml:"language,attr"`
	VersionScope          string          `xml:"versionScope,attr"`
	ComputerName          string          `xml:"ComputerName,omitempty"`
	RunSynchronous        *runSynchronous `xml:"RunSynchronous,omitempty"`
}

// NewShellSetupSpecialize builds the specialize-pass component from the
// computer name and the DisableWindowsUpdate/DisableUAC tweaks. It returns
// nil when there is nothing to configure: computerName is nil and both
// tweaks are off.
func NewShellSetupSpecialize(computerName *string, tweaks profile.SystemTweaks) *ShellSetupSpecialize {
	var commands []runSynchronousCommand
	if tweaks.DisableWindowsUpdate {
		commands = append(commands, runSynchronousCommand{Order: len(commands) + 1, Path: disableWindowsUpdateCommand})
	}
	if tweaks.DisableUAC {
		commands = append(commands, runSynchronousCommand{Order: len(commands) + 1, Path: disableUACCommand})
	}

	if computerName == nil && len(commands) == 0 {
		return nil
	}

	s := &ShellSetupSpecialize{
		Name:                  shellSetupName,
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
	}
	if computerName != nil {
		s.ComputerName = *computerName
	}
	if len(commands) > 0 {
		s.RunSynchronous = &runSynchronous{RunSynchronousCommand: commands}
	}
	return s
}

// password is the shared <Password> element used by local accounts, the
// built-in Administrator account and AutoLogon.
type password struct {
	Value     string `xml:"Value"`
	PlainText bool   `xml:"PlainText"`
}

type localAccount struct {
	Action      string    `xml:"action,attr"`
	Password    *password `xml:"Password,omitempty"`
	DisplayName string    `xml:"DisplayName,omitempty"`
	Group       string    `xml:"Group"`
	Name        string    `xml:"Name"`
}

type localAccounts struct {
	LocalAccount []localAccount `xml:"LocalAccount"`
}

type userAccounts struct {
	LocalAccounts         *localAccounts `xml:"LocalAccounts,omitempty"`
	AdministratorPassword *password      `xml:"AdministratorPassword,omitempty"`
}

type autoLogon struct {
	Password *password `xml:"Password,omitempty"`
	Enabled  bool      `xml:"Enabled"`
	Username string    `xml:"Username"`
}

// oobeBlock controls the OOBE express settings (telemetry) prompt. Value 1
// applies the recommended settings automatically (telemetry on); value 3
// turns automatic protection off. See Microsoft-Windows-Shell-Setup | OOBE |
// ProtectYourPC.
type oobeBlock struct {
	ProtectYourPC int `xml:"ProtectYourPC"`
}

// ShellSetupOOBE is the Microsoft-Windows-Shell-Setup component in the
// oobeSystem pass, carrying local accounts, auto-logon and express settings.
type ShellSetupOOBE struct {
	XMLName               xml.Name      `xml:"component"`
	Name                  string        `xml:"name,attr"`
	ProcessorArchitecture string        `xml:"processorArchitecture,attr"`
	PublicKeyToken        string        `xml:"publicKeyToken,attr"`
	Language              string        `xml:"language,attr"`
	VersionScope          string        `xml:"versionScope,attr"`
	OOBE                  *oobeBlock    `xml:"OOBE,omitempty"`
	UserAccounts          *userAccounts `xml:"UserAccounts,omitempty"`
	AutoLogon             *autoLogon    `xml:"AutoLogon,omitempty"`
}

// NewShellSetupOOBE builds the oobeSystem-pass component from accounts,
// firstLogon and express. It returns nil when there is nothing to configure.
func NewShellSetupOOBE(accounts []profile.UserAccount, firstLogon profile.FirstLogon, express profile.ExpressSettings) *ShellSetupOOBE {
	var ua *userAccounts
	if len(accounts) > 0 {
		ua = &userAccounts{LocalAccounts: &localAccounts{}}
		for _, a := range accounts {
			la := localAccount{
				Action: "add",
				Group:  string(a.Group),
				Name:   a.Name,
			}
			if a.DisplayName != nil {
				la.DisplayName = *a.DisplayName
			}
			if a.Password != nil {
				la.Password = &password{Value: *a.Password, PlainText: true}
			}
			ua.LocalAccounts.LocalAccount = append(ua.LocalAccounts.LocalAccount, la)
		}
	}

	var al *autoLogon
	switch firstLogon.Mode {
	case profile.FirstLogonBuiltinAdmin:
		pwd := ""
		if firstLogon.BuiltinAdministratorPassword != nil {
			pwd = *firstLogon.BuiltinAdministratorPassword
		}
		if ua == nil {
			ua = &userAccounts{}
		}
		ua.AdministratorPassword = &password{Value: pwd, PlainText: true}
		al = &autoLogon{
			Enabled:  true,
			Username: adminUsername,
			Password: &password{Value: pwd, PlainText: true},
		}
	case profile.FirstLogonFirstCreatedAccount:
		if len(accounts) > 0 {
			first := accounts[0]
			al = &autoLogon{Enabled: true, Username: first.Name}
			if first.Password != nil {
				al.Password = &password{Value: *first.Password, PlainText: true}
			}
		}
	case profile.FirstLogonNone:
		// no AutoLogon element
	}

	var oobe *oobeBlock
	switch express.Mode {
	case profile.ExpressAllEnabled:
		oobe = &oobeBlock{ProtectYourPC: 1}
	case profile.ExpressAllDisabled:
		oobe = &oobeBlock{ProtectYourPC: 3}
	case profile.ExpressInteractive:
		// no OOBE element: Windows Setup asks interactively
	}

	if ua == nil && al == nil && oobe == nil {
		return nil
	}
	return &ShellSetupOOBE{
		Name:                  shellSetupName,
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		OOBE:                  oobe,
		UserAccounts:          ua,
		AutoLogon:             al,
	}
}
