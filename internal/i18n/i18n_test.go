package i18n

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

var verbRe = regexp.MustCompile(`%[sdqv%]|%\[\d+\][sdqv]`)

func verbs(s string) string { return strings.Join(verbRe.FindAllString(s, -1), " ") }

func TestParse(t *testing.T) {
	for in, want := range map[string]Lang{
		"en": EN, "EN": EN, "en_US.UTF-8": EN, "ru": RU, "ru_RU.UTF-8": RU,
		"zh": ZH, "zh-CN": ZH, "zh_CN.UTF-8": ZH, "es_ES@euro": ES, "hi-IN": HI,
	} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "de", "C", "fr_FR"} {
		if _, ok := Parse(in); ok {
			t.Errorf("Parse(%q) should fail", in)
		}
	}
}

func TestTranslateAndFallback(t *testing.T) {
	defer Set("")
	Set("")
	if got := T("Disabled"); got != "Disabled" {
		t.Errorf("unset language must give English, got %q", got)
	}
	Set(EN)
	if got := T("Disabled"); got != "Disabled" {
		t.Errorf("English must give the key, got %q", got)
	}
	want := map[Lang]string{RU: "Отключено", ZH: "已禁用", ES: "Desactivado", HI: "अक्षम"}
	for l, w := range want {
		Set(l)
		if got := T("Disabled"); got != w {
			t.Errorf("T(Disabled) in %s = %q, want %q", l, got, w)
		}
	}
	Set(RU)
	ResetMissing()
	if got := T("no such text"); got != "no such text" {
		t.Errorf("unknown text must fall back to itself, got %q", got)
	}
	if m := Missing(); len(m) != 1 || m[0] != "ru: no such text" {
		t.Errorf("Missing() = %v", m)
	}
	if got := Tf("saved %s", "a.json"); got != "сохранено: a.json" {
		t.Errorf("Tf = %q", got)
	}
	Set(ZH)
	if T(".ps1") != ".ps1" {
		t.Error("technical tokens must stay unchanged")
	}
}

func TestMessagesKeepCanonicalTextAndDisplayTranslation(t *testing.T) {
	defer Set("")
	const tpl = "%s: обязательное поле"
	Set("")
	if got := Display(M(tpl, "name")); got != "name: обязательное поле" {
		t.Errorf("unset language must keep Russian, got %q", got)
	}
	for l, want := range map[Lang]string{
		EN: "name: required field", ZH: "name：必填字段", ES: "name: campo obligatorio", HI: "name: आवश्यक फ़ील्ड",
	} {
		Set(l)
		canon := M(tpl, "name")
		if canon != "name: обязательное поле" {
			t.Errorf("%s: canonical text changed: %q (screens route errors on it)", l, canon)
		}
		if got := Display(canon); got != want {
			t.Errorf("%s: Display = %q, want %q", l, got, want)
		}
	}
	Set(RU)
	if got := Display(M(tpl, "name")); got != "name: обязательное поле" {
		t.Errorf("Russian must be shown as is, got %q", got)
	}
}

// Every catalog entry must be complete and must use the same format verbs as
// its source text, otherwise Tf/M would print %!s(MISSING) at run time.
func TestCatalogsAreCompleteAndConsistent(t *testing.T) {
	langsUI := []string{"ru", "zh", "es", "hi"}
	for key, row := range catalog {
		for i, tr := range row {
			if strings.TrimSpace(tr) == "" || !utf8.ValidString(tr) {
				t.Errorf("UI %q: empty or invalid %s translation", key, langsUI[i])
			}
			if verbs(tr) != verbs(key) {
				t.Errorf("UI %q: %s translation %q has different verbs", key, langsUI[i], tr)
			}
		}
		if _, dup := keep[key]; dup {
			t.Errorf("UI %q is both translated and marked identical", key)
		}
	}
	langsMsg := []string{"en", "zh", "es", "hi"}
	for key, row := range msgCatalog {
		for i, tr := range row {
			if strings.TrimSpace(tr) == "" || !utf8.ValidString(tr) {
				t.Errorf("message %q: empty or invalid %s translation", key, langsMsg[i])
			}
			if verbs(tr) != verbs(key) {
				t.Errorf("message %q: %s translation %q has different verbs", key, langsMsg[i], tr)
			}
		}
	}
	if len(catalog) < 200 || len(msgCatalog) < 60 {
		t.Errorf("catalogs look truncated: %d UI, %d messages", len(catalog), len(msgCatalog))
	}
}
