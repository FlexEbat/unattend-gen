package xmlgen

import (
	"encoding/xml"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

type testShellComponent struct {
	Name                  string             `xml:"name,attr"`
	ComputerName          string             `xml:"ComputerName"`
	LocalAccounts         []testLocalAccount `xml:"UserAccounts>LocalAccounts>LocalAccount"`
	AdministratorPassword *testPassword      `xml:"UserAccounts>AdministratorPassword"`
	AutoLogon             *testAutoLogon     `xml:"AutoLogon"`
}

type testLocalAccount struct {
	Name        string `xml:"Name"`
	DisplayName string `xml:"DisplayName"`
	Group       string `xml:"Group"`
}

type testPassword struct {
	Value string `xml:"Value"`
}

type testAutoLogon struct {
	Username string        `xml:"Username"`
	Enabled  bool          `xml:"Enabled"`
	Password *testPassword `xml:"Password"`
}

type testAccountsPass struct {
	Pass       string               `xml:"pass,attr"`
	Components []testShellComponent `xml:"component"`
}

type testAccountsDoc struct {
	XMLName  xml.Name           `xml:"unattend"`
	Settings []testAccountsPass `xml:"settings"`
}

func buildAndParseAccounts(t *testing.T, p *profile.Profile) testAccountsDoc {
	t.Helper()
	xmlStr, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatalf("BuildAnswerFile: %v", err)
	}
	var doc testAccountsDoc
	if err := xml.Unmarshal([]byte(xmlStr), &doc); err != nil {
		t.Fatalf("Unmarshal result: %v\nxml:\n%s", err, xmlStr)
	}
	return doc
}

func findShellComponent(doc testAccountsDoc, pass string) *testShellComponent {
	for _, s := range doc.Settings {
		if s.Pass != pass {
			continue
		}
		for i := range s.Components {
			if s.Components[i].Name == "Microsoft-Windows-Shell-Setup" {
				return &s.Components[i]
			}
		}
	}
	return nil
}

func TestBuildAnswerFileComputerName(t *testing.T) {
	p := baseProfile()
	name := "MYPC01"
	p.ComputerName = &name

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "specialize")
	if shell == nil {
		t.Fatal("expected Shell-Setup component in specialize pass")
	}
	if shell.ComputerName != name {
		t.Fatalf("ComputerName = %q, want %q", shell.ComputerName, name)
	}
}

func TestBuildAnswerFileComputerNameNilOmitsElement(t *testing.T) {
	p := baseProfile()
	p.ComputerName = nil

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponent(doc, "specialize"); shell != nil {
		t.Fatalf("expected no Shell-Setup component in specialize pass, got %+v", shell)
	}
}

func TestBuildAnswerFileLocalAccounts(t *testing.T) {
	p := baseProfile()
	displayName := "Alice A"
	p.Accounts = []profile.UserAccount{
		{Name: "alice", DisplayName: &displayName, Group: profile.GroupAdministrators},
		{Name: "bob", Group: profile.GroupUsers},
	}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil {
		t.Fatal("expected Shell-Setup component in oobeSystem pass")
	}
	if len(shell.LocalAccounts) != 2 {
		t.Fatalf("got %d LocalAccount elements, want 2", len(shell.LocalAccounts))
	}
	if shell.LocalAccounts[0].Name != "alice" || shell.LocalAccounts[0].DisplayName != displayName || shell.LocalAccounts[0].Group != "Administrators" {
		t.Fatalf("unexpected first account: %+v", shell.LocalAccounts[0])
	}
	if shell.LocalAccounts[1].Name != "bob" || shell.LocalAccounts[1].DisplayName != "" || shell.LocalAccounts[1].Group != "Users" {
		t.Fatalf("unexpected second account: %+v", shell.LocalAccounts[1])
	}
}

func TestBuildAnswerFileFirstLogonBuiltinAdmin(t *testing.T) {
	p := baseProfile()
	pwd := "Sekret1!"
	p.FirstLogon = profile.FirstLogon{Mode: profile.FirstLogonBuiltinAdmin, BuiltinAdministratorPassword: &pwd}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.AutoLogon == nil {
		t.Fatal("expected AutoLogon in oobeSystem Shell-Setup component")
	}
	if shell.AutoLogon.Username != "Administrator" {
		t.Fatalf("AutoLogon.Username = %q, want Administrator", shell.AutoLogon.Username)
	}
	if shell.AutoLogon.Password == nil || shell.AutoLogon.Password.Value != pwd {
		t.Fatalf("AutoLogon.Password = %+v, want value %q", shell.AutoLogon.Password, pwd)
	}
	if shell.AdministratorPassword == nil || shell.AdministratorPassword.Value != pwd {
		t.Fatalf("AdministratorPassword = %+v, want value %q", shell.AdministratorPassword, pwd)
	}
}

func TestBuildAnswerFileFirstLogonFirstCreatedAccount(t *testing.T) {
	p := baseProfile()
	p.Accounts = []profile.UserAccount{{Name: "alice", Group: profile.GroupAdministrators}}
	p.FirstLogon = profile.FirstLogon{Mode: profile.FirstLogonFirstCreatedAccount}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.AutoLogon == nil {
		t.Fatal("expected AutoLogon in oobeSystem Shell-Setup component")
	}
	if shell.AutoLogon.Username != "alice" {
		t.Fatalf("AutoLogon.Username = %q, want alice", shell.AutoLogon.Username)
	}
}

func TestBuildAnswerFileFirstLogonNoneOmitsAutoLogon(t *testing.T) {
	p := baseProfile()
	p.FirstLogon = profile.FirstLogon{Mode: profile.FirstLogonNone}

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponent(doc, "oobeSystem"); shell != nil {
		t.Fatalf("expected no oobeSystem Shell-Setup component, got %+v", shell)
	}
}
