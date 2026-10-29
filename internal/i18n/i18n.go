// Package i18n translates the interface text of unattend-gen.
//
// UI strings are looked up by their English text (T, Tf); a missing
// translation falls back to the English text, so a gap never breaks the UI.
// Validation messages are different: package profile produces them in Russian
// and screens match on that exact text to route an error to the right field,
// so the canonical text must not change. M builds such a message from a
// Russian template and remembers its translation, and Display turns the
// canonical text into the selected language at render time.
//
// The language is chosen once per run (see Set) and is never persisted.
package i18n

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

// Lang is a supported interface language code.
type Lang string

// Supported languages. The empty Lang means "not chosen": UI text is English
// and validation messages keep their canonical Russian form (the behaviour of
// the CLI before language support was added).
const (
	EN Lang = "en"
	RU Lang = "ru"
	ZH Lang = "zh"
	ES Lang = "es"
	HI Lang = "hi"
)

// Info describes one selectable language.
type Info struct {
	Code Lang
	Name string // name in the language itself
}

// Languages lists the supported languages in the order the picker shows them.
var Languages = []Info{
	{EN, "English"},
	{RU, "Русский"},
	{ZH, "简体中文"},
	{ES, "Español"},
	{HI, "हिन्दी"},
}

var (
	mu  sync.RWMutex
	cur Lang

	// displayed maps a canonical (Russian) validation message to its text in
	// the current language, filled by M.
	displayed sync.Map
	// missing collects translations that were requested but not found.
	missing sync.Map
)

// Set selects the interface language.
func Set(l Lang) {
	mu.Lock()
	cur = l
	mu.Unlock()
	displayed.Range(func(k, _ any) bool { displayed.Delete(k); return true })
}

// Current returns the selected language ("" if none was chosen).
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return cur
}

// Parse accepts a language code or a locale name such as "ru", "zh-CN",
// "es_ES.UTF-8" or "en_US" and returns the matching supported language.
func Parse(s string) (Lang, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if i := strings.IndexAny(s, "._@"); i >= 0 {
		s = s[:i]
	}
	s = strings.SplitN(s, "-", 2)[0]
	for _, info := range Languages {
		if string(info.Code) == s {
			return info.Code, true
		}
	}
	return "", false
}

// uiCol is the column of l in a UI catalog row (ru, zh, es, hi); -1 for EN.
func uiCol(l Lang) int {
	switch l {
	case RU:
		return 0
	case ZH:
		return 1
	case ES:
		return 2
	case HI:
		return 3
	}
	return -1
}

// msgCol is the column of l in a message catalog row (en, zh, es, hi); -1 for RU.
func msgCol(l Lang) int {
	switch l {
	case EN:
		return 0
	case ZH:
		return 1
	case ES:
		return 2
	case HI:
		return 3
	}
	return -1
}

// T returns the translation of the English UI text key, or key itself.
func T(key string) string {
	l := Current()
	if l == "" || l == EN {
		return key
	}
	if _, ok := keep[key]; ok {
		return key
	}
	if c := uiCol(l); c >= 0 {
		if row, ok := catalog[key]; ok && row[c] != "" {
			return row[c]
		}
	}
	missing.Store(string(l)+"\x00"+key, struct{}{})
	return key
}

// Tf is T followed by fmt.Sprintf. The translation must contain the same
// verbs as the English text (a test enforces this).
func Tf(format string, args ...any) string {
	return fmt.Sprintf(T(format), args...)
}

// M builds a validation message from a Russian template, like fmt.Sprintf,
// and returns the canonical Russian text. If a language other than Russian
// is selected, the translated text is remembered for Display.
func M(ruFormat string, args ...any) string {
	canon := fmt.Sprintf(ruFormat, args...)
	l := Current()
	if l == "" || l == RU {
		return canon
	}
	var tr string
	if row, ok := msgCatalog[ruFormat]; ok && msgCol(l) >= 0 {
		tr = row[msgCol(l)]
	}
	if tr == "" {
		missing.Store(string(l)+"\x00"+ruFormat, struct{}{})
		return canon
	}
	displayed.Store(canon, fmt.Sprintf(tr, args...))
	return canon
}

// Display returns the selected-language text of a canonical validation
// message produced by M, or s unchanged.
func Display(s string) string {
	if v, ok := displayed.Load(s); ok {
		return v.(string)
	}
	return s
}

// Missing returns "lang: key" for every translation requested but absent
// since the last ResetMissing. Used by tests.
func Missing() []string {
	var out []string
	missing.Range(func(k, _ any) bool {
		out = append(out, strings.Replace(k.(string), "\x00", ": ", 1))
		return true
	})
	sort.Strings(out)
	return out
}

// ResetMissing clears the missing-translation record.
func ResetMissing() {
	missing.Range(func(k, _ any) bool { missing.Delete(k); return true })
}

// MD is Display(M(...)): a canonical Russian message built from a template and
// shown in the selected language. Used for CLI output.
func MD(ruFormat string, args ...any) string {
	return Display(M(ruFormat, args...))
}

// FromEnv guesses the user's language from LC_ALL, LC_MESSAGES and LANG. It
// only pre-selects an entry in the language picker; the user is still asked.
func FromEnv() Lang {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(name); v != "" {
			if l, ok := Parse(v); ok {
				return l
			}
		}
	}
	return EN
}

// Has reports whether a UI text is translated (or deliberately identical in
// every language). Used by the test that guards against untranslated text.
func Has(key string) bool {
	if _, ok := keep[key]; ok {
		return true
	}
	_, ok := catalog[key]
	return ok
}

// HasMsg reports whether a Russian validation-message template is translated.
func HasMsg(ruFormat string) bool {
	_, ok := msgCatalog[ruFormat]
	return ok
}
