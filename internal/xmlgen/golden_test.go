package xmlgen

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/presets"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/golden")

// goldenProfiles are the profiles whose complete answer files are pinned
// byte for byte. Any change to generated XML shows up as a diff of these
// files in review; refresh them with: go test ./internal/xmlgen -update
func goldenProfiles(t *testing.T) map[string]*profile.Profile {
	t.Helper()
	out := map[string]*profile.Profile{}
	for _, name := range presets.Names {
		data, err := presets.Load(name)
		if err != nil {
			t.Fatal(err)
		}
		res := profile.ValidateProfile(data)
		if len(res.Errors) > 0 {
			t.Fatalf("preset %s is invalid: %v", name, res.Errors)
		}
		out["preset-"+name] = res.Profile
	}
	typical := profile.Default("typical")
	pw := "correct horse"
	typical.Accounts = []profile.UserAccount{{Name: "user", Password: &pw, Group: profile.GroupAdministrators}}
	typical.FirstLogon.Mode = profile.FirstLogonFirstCreatedAccount
	typical.Wifi = &profile.WifiSettings{SSID: "home", Authentication: profile.WifiWPA2Personal, Password: &pw}
	typical.SystemTweaks.DisableSmartScreen = true
	typical.SystemTweaks.EnableLongPaths = true
	typical.SystemScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "Write-Host hello"}}
	out["typical"] = typical
	return out
}

func TestGoldenAnswerFiles(t *testing.T) {
	for name, p := range goldenProfiles(t) {
		t.Run(name, func(t *testing.T) {
			got, err := BuildAnswerFile(p)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", "golden", name+".xml")
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing golden file (run with -update): %v", err)
			}
			if string(want) != got {
				t.Fatalf("answer file for %s differs from %s; if the change is intended, run: go test ./internal/xmlgen -update", name, path)
			}
		})
	}
}

// FuzzValidateAndBuild feeds arbitrary bytes through the whole pipeline: a
// profile that validates must always produce well-formed XML in which every
// command fits Windows Setup's length limit, and nothing may panic.
func FuzzValidateAndBuild(f *testing.F) {
	for _, name := range presets.Names {
		data, err := presets.Load(name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	if data, err := json.Marshal(maxProfile()); err == nil {
		f.Add(data)
	}
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		res := profile.ValidateProfile(data)
		if len(res.Errors) > 0 {
			if res.Profile != nil {
				t.Fatal("Profile must be nil when there are errors")
			}
			return
		}
		out, err := BuildAnswerFile(res.Profile)
		if err != nil {
			t.Fatalf("a valid profile failed to build: %v", err)
		}
		_, lists := parseOffloadDoc(t, out)
		for _, l := range lists {
			for i, c := range l.cmds {
				if len(c) > 259 {
					t.Fatalf("%s command %d is %d characters", l.pass, i+1, len(c))
				}
			}
		}
	})
}
