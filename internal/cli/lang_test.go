package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
)

func badProfile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(p, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLangFlagTranslatesValidationMessages(t *testing.T) {
	defer i18n.Set("")
	path := badProfile(t)

	i18n.Set("")
	_, stderr, err := runCLI(t, "validate", path)
	if err == nil || !strings.Contains(stderr, "Имя профиля обязательно") {
		t.Fatalf("without --lang the CLI keeps its default Russian text; got err=%v stderr=%q", err, stderr)
	}
	for lang, want := range map[string]string{
		"en": "The profile name is required", "zh": "必须填写配置文件名称",
		"es": "El nombre del perfil es obligatorio", "hi": "प्रोफ़ाइल का नाम आवश्यक है", "ru": "Имя профиля обязательно",
	} {
		i18n.Set("")
		_, stderr, err := runCLI(t, "--lang", lang, "validate", path)
		if err == nil || !strings.Contains(stderr, want) {
			t.Errorf("--lang %s: want %q in stderr, got err=%v stderr=%q", lang, want, err, stderr)
		}
	}
}

func TestLangFromEnvironmentAndInvalidValue(t *testing.T) {
	defer i18n.Set("")
	path := badProfile(t)

	i18n.Set("")
	t.Setenv("UNATTEND_GEN_LANG", "es")
	if _, stderr, _ := runCLI(t, "validate", path); !strings.Contains(stderr, "El nombre del perfil es obligatorio") {
		t.Errorf("UNATTEND_GEN_LANG=es was ignored: %q", stderr)
	}

	i18n.Set("")
	_, _, err := runCLI(t, "--lang", "klingon", "validate", path)
	if err == nil || !strings.Contains(err.Error(), "unknown language") {
		t.Errorf("an unknown language must be an error, got %v", err)
	}
}
