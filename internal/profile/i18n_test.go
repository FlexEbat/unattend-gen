package profile

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
)

var cyrillic = regexp.MustCompile(`[А-Яа-я]`)

// Validation messages are written in Russian and translated by template (see
// package i18n). A new or edited message without a catalog entry would show
// up in Russian in every other language.
func TestEveryMessageTemplateIsTranslated(t *testing.T) {
	fset := token.NewFileSet()
	var n int
	for _, f := range []string{"validate.go", "warnings.go"} {
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(af, func(node ast.Node) bool {
			bl, ok := node.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(bl.Value)
			if err != nil || !cyrillic.MatchString(s) {
				return true
			}
			n++
			if !i18n.HasMsg(s) {
				t.Errorf("%s:%d: message %q has no translation (add it to internal/i18n/messages.go)", f, fset.Position(bl.Pos()).Line, s)
			}
			return true
		})
	}
	if n < 60 {
		t.Fatalf("found only %d Russian templates; the scan is broken", n)
	}
}

// The text screens match on must not depend on the selected language.
func TestCanonicalMessagesAreLanguageIndependent(t *testing.T) {
	defer i18n.Set("")
	data := []byte(`{"schema_version":1,"name":"x","accounts":[{"name":"bad/name"}]}`)
	var ref []string
	for i, l := range append([]i18n.Lang{""}, i18n.EN, i18n.RU, i18n.ZH, i18n.ES, i18n.HI) {
		i18n.Set(l)
		res := ValidateProfile(data)
		if len(res.Errors) == 0 {
			t.Fatal("expected a validation error")
		}
		if i == 0 {
			ref = res.Errors
			continue
		}
		if len(res.Errors) != len(ref) || res.Errors[0] != ref[0] {
			t.Errorf("language %q changed the canonical errors: %v vs %v", l, res.Errors, ref)
		}
		if l != i18n.RU && i18n.Display(res.Errors[0]) == res.Errors[0] {
			t.Errorf("language %q: error was not translated: %q", l, res.Errors[0])
		}
	}
}
