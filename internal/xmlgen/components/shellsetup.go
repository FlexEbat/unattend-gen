package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	shellSetupName = "Microsoft-Windows-Shell-Setup"
	adminUsername  = "Administrator"
)

// ShellSetupSpecialize is the Microsoft-Windows-Shell-Setup component in the
// specialize pass, carrying the computer name. Registry tweaks used to live
// here too, but Shell-Setup has no RunSynchronous list in the real schema:
// that belongs to Microsoft-Windows-Deployment (see NewDeployment below).
type ShellSetupSpecialize struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	ComputerName string `xml:"ComputerName,omitempty"`
}

// NewShellSetupSpecialize builds the specialize-pass component from the
// computer name. Returns nil when computerName is nil: Windows generates a
// random name itself.
func NewShellSetupSpecialize(computerName *string) *ShellSetupSpecialize {
	if computerName == nil {
		return nil
	}
	return &ShellSetupSpecialize{
		Name:          shellSetupName,
		standardAttrs: newStandardAttrs(),
		ComputerName:  *computerName,
	}
}

// disableWindowsUpdateCommand and disableUACCommand are the registry edits
// for the corresponding SystemTweaks flags. They run in the specialize pass.
const (
	disableWindowsUpdateCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU" /v NoAutoUpdate /t REG_DWORD /d 1 /f`
	disableUACCommand           = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v EnableLUA /t REG_DWORD /d 0 /f`
)

// Deployment is the Microsoft-Windows-Deployment component, specialize pass:
// the real home of RunSynchronousCommand outside windowsPE.
type Deployment struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	RunSynchronous *runSynchronous `xml:"RunSynchronous,omitempty"`
}

// NewDeployment builds the specialize-pass component for the
// DisableWindowsUpdate/DisableUAC tweaks. Returns nil when both are off: an
// empty component is not emitted.
func NewDeployment(tweaks profile.SystemTweaks) *Deployment {
	var commands []runSynchronousCommand
	if tweaks.DisableWindowsUpdate {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, disableWindowsUpdateCommand))
	}
	if tweaks.DisableUAC {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, disableUACCommand))
	}
	if len(commands) == 0 {
		return nil
	}
	return &Deployment{
		Name:           "Microsoft-Windows-Deployment",
		standardAttrs:  newStandardAttrs(),
		RunSynchronous: &runSynchronous{RunSynchronousCommand: commands},
	}
}

// password is the shared <Password> element used by local accounts, the
// built-in Administrator account and AutoLogon.
type password struct {
	Value     string `xml:"Value"`
	PlainText bool   `xml:"PlainText"`
}

type localAccount struct {
	Action      string    `xml:"wcm:action,attr"`
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

// autoLogon is Microsoft-Windows-Shell-Setup/AutoLogon. LogonCount is always
// 1: log on automatically once, then behave normally.
type autoLogon struct {
	Password   *password `xml:"Password,omitempty"`
	Enabled    bool      `xml:"Enabled"`
	LogonCount int       `xml:"LogonCount"`
	Username   string    `xml:"Username"`
}

// oobeBlock controls the OOBE express settings (telemetry) prompt. Value 1
// applies the recommended settings automatically (telemetry on); value 3
// turns automatic protection off. See Microsoft-Windows-Shell-Setup | OOBE |
// ProtectYourPC.
type oobeBlock struct {
	ProtectYourPC int `xml:"ProtectYourPC"`
}

// firstLogonCommands is Microsoft-Windows-Shell-Setup/FirstLogonCommands:
// commands run once, after the first user logs on. Used here to provision
// Wi-Fi, since the WLAN service is not guaranteed to be up any earlier.
type firstLogonCommands struct {
	SynchronousCommand []synchronousCommand `xml:"SynchronousCommand"`
}

type synchronousCommand struct {
	Action      string `xml:"wcm:action,attr"`
	CommandLine string `xml:"CommandLine"`
	Order       int    `xml:"Order"`
}

// ShellSetupOOBE is the Microsoft-Windows-Shell-Setup component in the
// oobeSystem pass, carrying local accounts, auto-logon, express settings and
// (via FirstLogonCommands) Wi-Fi provisioning.
type ShellSetupOOBE struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	OOBE               *oobeBlock          `xml:"OOBE,omitempty"`
	UserAccounts       *userAccounts       `xml:"UserAccounts,omitempty"`
	AutoLogon          *autoLogon          `xml:"AutoLogon,omitempty"`
	FirstLogonCommands *firstLogonCommands `xml:"FirstLogonCommands,omitempty"`
}

// NewShellSetupOOBE builds the oobeSystem-pass component from accounts,
// firstLogon, express and wifi. It returns nil when there is nothing to
// configure.
func NewShellSetupOOBE(accounts []profile.UserAccount, firstLogon profile.FirstLogon, express profile.ExpressSettings, wifi *profile.WifiSettings) *ShellSetupOOBE {
	var ua *userAccounts
	if len(accounts) > 0 {
		ua = &userAccounts{LocalAccounts: &localAccounts{}}
		for _, a := range accounts {
			la := localAccount{
				Action: wcmActionAdd,
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
			Enabled:    true,
			LogonCount: 1,
			Username:   adminUsername,
			Password:   &password{Value: pwd, PlainText: true},
		}
	case profile.FirstLogonFirstCreatedAccount:
		if len(accounts) > 0 {
			first := accounts[0]
			al = &autoLogon{Enabled: true, LogonCount: 1, Username: first.Name}
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

	var flc *firstLogonCommands
	if cmd := WifiFirstLogonCommand(wifi); cmd != "" {
		flc = &firstLogonCommands{SynchronousCommand: []synchronousCommand{
			{Action: wcmActionAdd, Order: 1, CommandLine: cmd},
		}}
	}

	if ua == nil && al == nil && oobe == nil && flc == nil {
		return nil
	}
	return &ShellSetupOOBE{
		Name:               shellSetupName,
		standardAttrs:      newStandardAttrs(),
		OOBE:               oobe,
		UserAccounts:       ua,
		AutoLogon:          al,
		FirstLogonCommands: flc,
	}
}
