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
// specialize pass, carrying the computer name.
type ShellSetupSpecialize struct {
	XMLName               xml.Name `xml:"component"`
	Name                  string   `xml:"name,attr"`
	ProcessorArchitecture string   `xml:"processorArchitecture,attr"`
	PublicKeyToken        string   `xml:"publicKeyToken,attr"`
	Language              string   `xml:"language,attr"`
	VersionScope          string   `xml:"versionScope,attr"`
	ComputerName          string   `xml:"ComputerName"`
}

// NewShellSetupSpecialize builds the specialize-pass component. It returns
// nil when computerName is nil: Windows generates a random name itself.
func NewShellSetupSpecialize(computerName *string) *ShellSetupSpecialize {
	if computerName == nil {
		return nil
	}
	return &ShellSetupSpecialize{
		Name:                  shellSetupName,
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		ComputerName:          *computerName,
	}
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

// ShellSetupOOBE is the Microsoft-Windows-Shell-Setup component in the
// oobeSystem pass, carrying local accounts and auto-logon.
type ShellSetupOOBE struct {
	XMLName               xml.Name      `xml:"component"`
	Name                  string        `xml:"name,attr"`
	ProcessorArchitecture string        `xml:"processorArchitecture,attr"`
	PublicKeyToken        string        `xml:"publicKeyToken,attr"`
	Language              string        `xml:"language,attr"`
	VersionScope          string        `xml:"versionScope,attr"`
	UserAccounts          *userAccounts `xml:"UserAccounts,omitempty"`
	AutoLogon             *autoLogon    `xml:"AutoLogon,omitempty"`
}

// NewShellSetupOOBE builds the oobeSystem-pass component from accounts and
// firstLogon. It returns nil when there is nothing to configure: no accounts
// and FirstLogonMode == FirstLogonNone.
func NewShellSetupOOBE(accounts []profile.UserAccount, firstLogon profile.FirstLogon) *ShellSetupOOBE {
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

	if ua == nil && al == nil {
		return nil
	}
	return &ShellSetupOOBE{
		Name:                  shellSetupName,
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		UserAccounts:          ua,
		AutoLogon:             al,
	}
}
