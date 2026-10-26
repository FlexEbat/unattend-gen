package xmlgen

import (
	"encoding/base64"
	"encoding/xml"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen/components"
)

// maxProfile enables as many command-producing features as the schema
// allows, so every mechanism that can emit a long command is exercised.
func maxProfile() *profile.Profile {
	p := profile.Default("max")
	v := reflect.ValueOf(&p.SystemTweaks).Elem()
	for i := 0; i < v.NumField(); i++ {
		v.Field(i).SetBool(true)
	}
	pw := "pw&<secret>"
	p.RemoveApps = append(p.RemoveApps, profile.RemovableApps...)
	p.RemoveFeatures = append(p.RemoveFeatures, profile.RemovableFeatures...)
	p.RemoveOptionalFeatures = append(p.RemoveOptionalFeatures, profile.RemovableOptionalFeatures...)
	p.Accounts = []profile.UserAccount{{Name: "user", Password: &pw, Group: profile.GroupAdministrators}}
	p.FirstLogon.Mode = profile.FirstLogonFirstCreatedAccount
	p.Wifi = &profile.WifiSettings{SSID: "net", Authentication: profile.WifiWPA2Personal, Password: &pw}
	p.SystemScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "Write-Host \"system\""}}
	p.DefaultUserScripts = []profile.CustomScript{{Format: profile.ScriptReg, Content: "Windows Registry Editor Version 5.00"}}
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptCmd, Content: "echo first"}}
	p.UserOnceScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "Write-Host once"}}
	p.InstallVMGuestTools = append(p.InstallVMGuestTools, profile.VMGuestTools...)
	p.DesktopIcons = map[profile.DesktopIcon]bool{}
	for _, d := range profile.DesktopIcons {
		p.DesktopIcons[d] = true
	}
	p.StartFolders = append(p.StartFolders, profile.StartFolders...)
	p.VisualEffects.Mode = profile.VisualEffectsModeBestPerformance
	p.TaskbarSearch = profile.TaskbarSearchModeHide
	p.TaskbarIcons.Mode = profile.TaskbarIconsModeEmpty
	p.StartPins.Mode = profile.StartPinsModeEmpty
	p.StartTiles.Mode = profile.StartTilesModeEmpty
	p.StickyKeys.Mode = profile.StickyKeysModeDisabled
	p.UseNarrator = true
	p.RestartExplorerAfterScripts = true
	name := "return 'PC'"
	p.ComputerName = nil
	p.ComputerNameScript = &name
	policy := `<AppLockerPolicy Version="1"/>`
	p.AppLockerPolicyXML = &policy
	return p
}

// docCommands is one pass's list of commands, in document order.
type docCommands struct {
	pass string
	cmds []string
	ord  []int
}

type offloadDoc struct {
	XMLName  xml.Name `xml:"unattend"`
	Settings []struct {
		Pass       string `xml:"pass,attr"`
		Components []struct {
			Name           string `xml:"name,attr"`
			RunSynchronous struct {
				Commands []struct {
					Order int    `xml:"Order"`
					Path  string `xml:"Path"`
				} `xml:"RunSynchronousCommand"`
			} `xml:"RunSynchronous"`
			FirstLogonCommands struct {
				Commands []struct {
					Order       int    `xml:"Order"`
					CommandLine string `xml:"CommandLine"`
				} `xml:"SynchronousCommand"`
			} `xml:"FirstLogonCommands"`
		} `xml:"component"`
	} `xml:"settings"`
	Extensions *struct {
		XMLName       xml.Name `xml:"https://schneegans.de/windows/unattend-generator/ Extensions"`
		ExtractScript string   `xml:"ExtractScript"`
		Files         []struct {
			Path string `xml:"path,attr"`
			Text string `xml:",chardata"`
		} `xml:"File"`
	} `xml:"Extensions"`
}

func parseOffloadDoc(t *testing.T, xmlStr string) (offloadDoc, []docCommands) {
	t.Helper()
	var d offloadDoc
	if err := xml.Unmarshal([]byte(xmlStr), &d); err != nil {
		t.Fatalf("Unmarshal: %v\n%s", err, xmlStr)
	}
	var lists []docCommands
	for _, s := range d.Settings {
		for _, c := range s.Components {
			if n := len(c.RunSynchronous.Commands); n > 0 {
				l := docCommands{pass: s.Pass + "/" + c.Name}
				for _, x := range c.RunSynchronous.Commands {
					l.cmds = append(l.cmds, x.Path)
					l.ord = append(l.ord, x.Order)
				}
				lists = append(lists, l)
			}
			if n := len(c.FirstLogonCommands.Commands); n > 0 {
				l := docCommands{pass: s.Pass + "/" + c.Name + "/FirstLogon"}
				for _, x := range c.FirstLogonCommands.Commands {
					l.cmds = append(l.cmds, x.CommandLine)
					l.ord = append(l.ord, x.Order)
				}
				lists = append(lists, l)
			}
		}
	}
	return d, lists
}

func TestEveryCommandFitsWindowsSetupLimit(t *testing.T) {
	out, err := BuildAnswerFile(maxProfile())
	if err != nil {
		t.Fatal(err)
	}
	_, lists := parseOffloadDoc(t, out)
	total := 0
	for _, l := range lists {
		for i, c := range l.cmds {
			total++
			if len(c) > components.MaxCommandLength {
				t.Errorf("%s command %d is %d characters (max %d): %.120s...", l.pass, i+1, len(c), components.MaxCommandLength, c)
			}
		}
	}
	if total < 50 {
		t.Fatalf("only %d commands found - the maximal profile no longer exercises every mechanism", total)
	}
}

func TestCommandOrdersStayContiguous(t *testing.T) {
	out, err := BuildAnswerFile(maxProfile())
	if err != nil {
		t.Fatal(err)
	}
	_, lists := parseOffloadDoc(t, out)
	for _, l := range lists {
		for i, o := range l.ord {
			if o != i+1 {
				t.Errorf("%s: command %d has Order %d", l.pass, i+1, o)
			}
		}
	}
}

func TestOffloadingCommandsIsLossless(t *testing.T) {
	p := maxProfile()
	inlineXML, err := buildInline(p)
	if err != nil {
		t.Fatal(err)
	}
	offXML, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatal(err)
	}
	_, inline := parseOffloadDoc(t, inlineXML)
	d, off := parseOffloadDoc(t, offXML)
	if d.Extensions == nil {
		t.Fatal("max profile produced no <Extensions>")
	}
	files := map[string]string{}
	for _, f := range d.Extensions.Files {
		files[f.Path] = strings.TrimSpace(f.Text)
	}

	fileCmd := regexp.MustCompile(`^powershell\.exe -NoProfile -ExecutionPolicy Bypass -File "([^"]+)"$`)
	restore := func(c string) string {
		if m := fileCmd.FindStringSubmatch(c); m != nil {
			body, ok := files[m[1]]
			if !ok {
				t.Fatalf("command refers to %s, which is not in <Extensions>", m[1])
			}
			return `powershell -NoProfile -Command "` + strings.ReplaceAll(body, `"`, `\"`) + `"`
		}
		return c
	}
	// Expected: the inline commands, with long "cmd.exe /c a && b" chains
	// split into one command per part.
	var want []string
	for _, l := range inline {
		for _, c := range l.cmds {
			if len(c) > components.MaxCommandLength && strings.HasPrefix(c, "cmd.exe /c ") && strings.Contains(c, " && ") {
				for _, part := range strings.Split(strings.TrimPrefix(c, "cmd.exe /c "), " && ") {
					want = append(want, "cmd.exe /c "+part)
				}
				continue
			}
			want = append(want, c)
		}
	}
	var got []string
	for _, l := range off {
		for i, c := range l.cmds {
			if strings.HasPrefix(l.pass, "specialize/Microsoft-Windows-Deployment") && i == 0 && strings.Contains(c, "ExtractScript") {
				continue // the extraction command itself is new
			}
			got = append(got, restore(c))
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("offloaded commands differ from inline ones: got %d commands, want %d", len(got), len(want))
	}
}

func TestExtractionCommandRunsFirstInSpecialize(t *testing.T) {
	out, err := BuildAnswerFile(maxProfile())
	if err != nil {
		t.Fatal(err)
	}
	d, lists := parseOffloadDoc(t, out)
	if d.Extensions == nil || strings.TrimSpace(d.Extensions.ExtractScript) != strings.TrimSpace(components.ExtractScript) {
		t.Fatal("ExtractScript missing or altered")
	}
	for _, l := range lists {
		if l.pass == "specialize/Microsoft-Windows-Deployment" {
			if !strings.Contains(l.cmds[0], `C:\Windows\Panther\unattend.xml`) || !strings.Contains(l.cmds[0], "ExtractScript") {
				t.Fatalf("first specialize command is not the extraction command: %s", l.cmds[0])
			}
			return
		}
	}
	t.Fatal("no specialize Deployment component")
}

func TestExtractionIsAddedWhenOnlyFirstLogonCommandsAreLong(t *testing.T) {
	p := profile.Default("wifi-only")
	pw := "password123"
	p.Wifi = &profile.WifiSettings{SSID: "net", Authentication: profile.WifiWPA2Personal, Password: &pw}
	p.KeepSensitiveFiles = true // no cleanup command: the Wi-Fi command is the only long one
	out, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatal(err)
	}
	d, lists := parseOffloadDoc(t, out)
	if d.Extensions == nil || len(d.Extensions.Files) != 1 {
		t.Fatalf("want exactly one extension file, got %+v", d.Extensions)
	}
	found := false
	for _, l := range lists {
		if l.pass == "specialize/Microsoft-Windows-Deployment" && len(l.cmds) == 1 && strings.Contains(l.cmds[0], "ExtractScript") {
			found = true
		}
	}
	if !found {
		t.Fatal("specialize Deployment with the extraction command was not created")
	}
}

func TestNoExtensionsWhenNothingIsLong(t *testing.T) {
	p := profile.Default("small")
	p.SystemTweaks.DisableUAC = true
	p.KeepSensitiveFiles = true // otherwise the (long) cleanup command is always emitted
	out, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Extensions") {
		t.Fatalf("unexpected <Extensions> in a profile with only short commands:\n%s", out)
	}
}

func TestSensitiveFilesCleanupAlsoRemovesOffloadedCommandFiles(t *testing.T) {
	out, err := BuildAnswerFile(maxProfile())
	if err != nil {
		t.Fatal(err)
	}
	d, _ := parseOffloadDoc(t, out)
	payload := regexp.MustCompile(`FromBase64String\('([A-Za-z0-9+/=]+)'\)`)
	for _, f := range d.Extensions.Files {
		if !strings.Contains(f.Text, "unattend-delete-sensitive-files.ps1") {
			continue
		}
		m := payload.FindStringSubmatch(f.Text)
		if m == nil {
			t.Fatalf("no Base64 payload in %s", f.Path)
		}
		raw, err := base64.StdEncoding.DecodeString(m[1])
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), "unattend-cmd-*") {
			t.Fatalf("cleanup script does not remove offloaded command files:\n%s", raw)
		}
		return
	}
	t.Fatal("cleanup command not found among extension files")
}

func TestWifiPasswordNeverAppearsInPlaintext(t *testing.T) {
	out, err := BuildAnswerFile(maxProfile())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<secret>") {
		t.Fatal("Wi-Fi password leaked into the answer file in plaintext outside the account password element")
	}
}
