package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileExpressAllDisabled(t *testing.T) {
	p := baseProfile()
	p.ExpressSettings = profile.ExpressSettings{Mode: profile.ExpressAllDisabled}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.OOBE == nil {
		t.Fatal("expected an OOBE block in oobeSystem Shell-Setup component")
	}
	if shell.OOBE.ProtectYourPC != 3 {
		t.Fatalf("ProtectYourPC = %d, want 3 for ExpressAllDisabled", shell.OOBE.ProtectYourPC)
	}
}

func TestBuildAnswerFileExpressAllEnabled(t *testing.T) {
	p := baseProfile()
	p.ExpressSettings = profile.ExpressSettings{Mode: profile.ExpressAllEnabled}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.OOBE == nil {
		t.Fatal("expected an OOBE block in oobeSystem Shell-Setup component")
	}
	if shell.OOBE.ProtectYourPC != 1 {
		t.Fatalf("ProtectYourPC = %d, want 1 for ExpressAllEnabled", shell.OOBE.ProtectYourPC)
	}
}

func TestBuildAnswerFileExpressInteractiveOmitsProtectYourPC(t *testing.T) {
	p := baseProfile()
	// Accounts alone now sets HideOnlineAccountScreens (see the Wi-Fi/OOBE
	// slice), so this test checks ProtectYourPC specifically, not the whole
	// OOBE block.
	p.Accounts = []profile.UserAccount{{Name: "alice", Group: profile.GroupAdministrators}}
	p.ExpressSettings = profile.ExpressSettings{Mode: profile.ExpressInteractive}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil {
		t.Fatal("expected oobeSystem Shell-Setup component (for the account)")
	}
	if shell.OOBE == nil {
		t.Fatal("expected an OOBE block: accounts imply HideOnlineAccountScreens")
	}
	if shell.OOBE.ProtectYourPC != 0 {
		t.Fatalf("expected no ProtectYourPC for ExpressInteractive, got %d", shell.OOBE.ProtectYourPC)
	}
}

func TestBuildAnswerFileSystemTweaksEachAddsOneCommand(t *testing.T) {
	name := "MYPC01"

	cases := []struct {
		name    string
		tweaks  profile.SystemTweaks
		pass    string
		wantSub string
	}{
		{
			name:    "disable windows update",
			tweaks:  profile.SystemTweaks{DisableWindowsUpdate: true},
			pass:    "specialize",
			wantSub: "NoAutoUpdate",
		},
		{
			name:    "disable UAC",
			tweaks:  profile.SystemTweaks{DisableUAC: true},
			pass:    "specialize",
			wantSub: "EnableLUA",
		},
		{
			name:    "bypass Windows 11 requirements",
			tweaks:  profile.SystemTweaks{BypassWin11Requirements: true},
			pass:    "windowsPE",
			wantSub: "BypassTPMCheck",
		},
		{
			name:    "disable Smart App Control",
			tweaks:  profile.SystemTweaks{DisableSmartAppControl: true},
			pass:    "specialize",
			wantSub: "VerifiedAndReputablePolicyState",
		},
		{
			name:    "disable SmartScreen",
			tweaks:  profile.SystemTweaks{DisableSmartScreen: true},
			pass:    "specialize",
			wantSub: "SmartScreenEnabled",
		},
		{
			name:    "disable Fast Startup",
			tweaks:  profile.SystemTweaks{DisableFastStartup: true},
			pass:    "specialize",
			wantSub: "HiberbootEnabled",
		},
		{
			name:    "disable System Restore",
			tweaks:  profile.SystemTweaks{DisableSystemRestore: true},
			pass:    "specialize",
			wantSub: "DisableSR",
		},
		{
			name:    "enable long paths",
			tweaks:  profile.SystemTweaks{EnableLongPaths: true},
			pass:    "specialize",
			wantSub: "LongPathsEnabled",
		},
		{
			name:    "enable Remote Desktop",
			tweaks:  profile.SystemTweaks{EnableRemoteDesktop: true},
			pass:    "specialize",
			wantSub: "fDenyTSConnections",
		},
		{
			name:    "allow PowerShell scripts",
			tweaks:  profile.SystemTweaks{AllowPowerShellScripts: true},
			pass:    "specialize",
			wantSub: "ExecutionPolicy",
		},
		{
			name:    "disable last access timestamp",
			tweaks:  profile.SystemTweaks{DisableLastAccessTimestamp: true},
			pass:    "specialize",
			wantSub: "disablelastaccess",
		},
		{
			name:    "prevent device encryption",
			tweaks:  profile.SystemTweaks{PreventDeviceEncryption: true},
			pass:    "specialize",
			wantSub: "PreventDeviceEncryption",
		},
		{
			name:    "disable auto sign-on of last user",
			tweaks:  profile.SystemTweaks{DisableAutoSignOnLastUser: true},
			pass:    "specialize",
			wantSub: "DisableAutomaticRestartSignOn",
		},
		{
			name:    "disable WPBT execution",
			tweaks:  profile.SystemTweaks{DisableWPBT: true},
			pass:    "specialize",
			wantSub: "DisableWpbtExecution",
		},
		{
			name:    "audit process creation",
			tweaks:  profile.SystemTweaks{AuditProcessCreation: true},
			pass:    "specialize",
			wantSub: "ProcessCreationIncludeCmdLine_Enabled",
		},
		{
			name:    "hide Edge first run",
			tweaks:  profile.SystemTweaks{HideEdgeFirstRun: true},
			pass:    "specialize",
			wantSub: "HideFirstRunExperience",
		},
		{
			name:    "disable Edge startup boost",
			tweaks:  profile.SystemTweaks{DisableEdgeStartupBoost: true},
			pass:    "specialize",
			wantSub: "StartupBoostEnabled",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := baseProfile()
			p.ComputerName = &name
			p.SystemTweaks = tc.tweaks

			doc := buildAndParseAccounts(t, p)
			shell := findShellComponentByName(doc, tc.pass, componentNameFor(tc.tweaks))
			if shell == nil {
				t.Fatalf("expected a component with a command in pass %s", tc.pass)
			}
			if len(shell.RunSynchronousCommand) != 1 {
				t.Fatalf("got %d RunSynchronousCommand elements, want exactly 1", len(shell.RunSynchronousCommand))
			}
			if !strings.Contains(shell.RunSynchronousCommand[0].Path, tc.wantSub) {
				t.Fatalf("command Path = %q, want it to contain %q", shell.RunSynchronousCommand[0].Path, tc.wantSub)
			}
		})
	}
}

func TestBuildAnswerFileAllTweaksOffAddsNoRunSynchronous(t *testing.T) {
	p := baseProfile()
	name := "MYPC01"
	p.ComputerName = &name
	p.SystemTweaks = profile.SystemTweaks{}

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); shell != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", shell)
	}
	if shell := findShellComponent(doc, "windowsPE"); shell != nil {
		t.Fatalf("expected no Shell-Setup component in windowsPE, got %+v", shell)
	}
}

// componentNameFor returns which component carries the command for a given
// (single-flag) tweaks value: Microsoft-Windows-Deployment for the
// specialize-pass tweaks (Shell-Setup has no RunSynchronous list in the real
// schema), Microsoft-Windows-Setup for the windowsPE bypass.
func componentNameFor(t profile.SystemTweaks) string {
	if t.BypassWin11Requirements {
		return "Microsoft-Windows-Setup"
	}
	return "Microsoft-Windows-Deployment"
}

func findShellComponentByName(doc testAccountsDoc, pass, name string) *testShellComponent {
	for _, s := range doc.Settings {
		if s.Pass != pass {
			continue
		}
		for i := range s.Components {
			if s.Components[i].Name == name {
				return &s.Components[i]
			}
		}
	}
	return nil
}
