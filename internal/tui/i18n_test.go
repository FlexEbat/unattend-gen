package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
)

var identifierLike = regexp.MustCompile(`^[0-9A-Za-z]*$`)

// uiCalls are the constructors whose string arguments are shown to the user.
var uiCalls = map[string]bool{
	"NewConfirmBar": true, "NewLabeledSelect": true, "NewLabeledInput": true,
	"NewLabeledTextArea": true, "NewPasswordInput": true, "T": true, "Tf": true,
}

func callName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	}
	return ""
}

func stringLit(e ast.Expr) (string, bool) {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	return s, err == nil
}

// TestEveryUITextIsTranslated fails when interface text is added to the TUI
// without an entry in the i18n catalog (or an explicit "same in every
// language" mark). It reads the sources, so it also covers text that only
// appears after toggling a setting.
func TestEveryUITextIsTranslated(t *testing.T) {
	var found, missing int
	check := func(pos token.Position, s string) {
		if !strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") {
			return
		}
		found++
		if !i18n.Has(s) {
			missing++
			t.Errorf("%s:%d: %q has no translation (add it to internal/i18n/ui.go)", filepath.Base(pos.Filename), pos.Line, s)
		}
	}
	fset := token.NewFileSet()
	for _, dir := range []string{"screens", "widgets", "."} {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") || strings.HasSuffix(f, "trace.go") {
				continue
			}
			af, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(af, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					if uiCalls[callName(x.Fun)] {
						for _, a := range x.Args {
							if s, ok := stringLit(a); ok {
								check(fset.Position(a.Pos()), s)
							}
						}
					}
				case *ast.KeyValueExpr:
					if id, ok := x.Key.(*ast.Ident); ok && (id.Name == "Label" || id.Name == "Title") {
						if s, ok := stringLit(x.Value); ok {
							check(fset.Position(x.Value.Pos()), s)
						}
					}
				case *ast.ValueSpec:
					if strings.HasSuffix(x.Names[0].Name, "Labels") {
						ast.Inspect(x, func(m ast.Node) bool {
							if bl, ok := m.(*ast.BasicLit); ok {
								if s, ok := stringLit(bl); ok {
									check(fset.Position(bl.Pos()), s)
								}
							}
							return true
						})
					}
				}
				return true
			})
		}
	}
	if found < 200 {
		t.Fatalf("scan found only %d UI strings; the test is looking in the wrong place", found)
	}
	_ = missing
}

// Screens must still build and render in every language, with no panic from
// a bad format string and no untranslated text on the first screen.
func TestScreensRenderInEveryLanguage(t *testing.T) {
	defer i18n.Set("")
	for _, info := range i18n.Languages {
		i18n.Set(info.Code)
		i18n.ResetMissing()
		for _, id := range allIDs() {
			m := modelAt(id)
			if out := m.View(); strings.TrimSpace(out) == "" {
				t.Errorf("%s: screen %d rendered nothing", info.Code, id)
			}
		}
		var miss []string
		for _, m := range i18n.Missing() {
			// Data such as Windows app identifiers ("Calculator", "3DViewer")
			// and numbers is shown as is; interface text is covered strictly
			// by TestEveryUITextIsTranslated.
			if text := strings.TrimPrefix(m, string(info.Code)+": "); !identifierLike.MatchString(text) {
				miss = append(miss, m)
			}
		}
		if len(miss) > 0 {
			t.Errorf("%s: untranslated text was requested: %v", info.Code, miss)
		}
	}
}

func TestChosenLanguageIsUsedByTheScreens(t *testing.T) {
	defer i18n.Set("")
	i18n.Set(i18n.RU)
	view := modelAt(screens.ScreenWelcome).View()
	if !strings.Contains(view, "Новый профиль") {
		t.Fatalf("Russian welcome screen must contain %q:\n%s", "Новый профиль", view)
	}
	i18n.Set(i18n.ZH)
	if view := modelAt(screens.ScreenWelcome).View(); !strings.Contains(view, "新建配置文件") {
		t.Fatalf("Chinese welcome screen must contain %q:\n%s", "新建配置文件", view)
	}
}

func TestLanguagePicker(t *testing.T) {
	press := func(p languagePicker, key string) languagePicker {
		var k tea.KeyMsg
		switch key {
		case "enter":
			k = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			k = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			k = tea.KeyMsg{Type: tea.KeyUp}
		default:
			k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		next, _ := p.Update(k)
		return next.(languagePicker)
	}
	// Every language is listed by its own name.
	view := newLanguagePicker(i18n.EN).View()
	for _, info := range i18n.Languages {
		if !strings.Contains(view, info.Name) {
			t.Errorf("picker does not list %q", info.Name)
		}
	}
	// Pre-selection, arrows with wrap-around, Enter.
	p := newLanguagePicker(i18n.ES)
	if p = press(p, "enter"); p.chosen != i18n.ES {
		t.Errorf("pre-selected Spanish, Enter chose %q", p.chosen)
	}
	p = press(press(newLanguagePicker(i18n.EN), "up"), "enter")
	if p.chosen != i18n.HI {
		t.Errorf("up from English must wrap to the last language, got %q", p.chosen)
	}
	p = press(press(newLanguagePicker(i18n.EN), "down"), "enter")
	if p.chosen != i18n.RU {
		t.Errorf("down from English must give Russian, got %q", p.chosen)
	}
	// Number keys choose directly.
	if p = press(newLanguagePicker(i18n.EN), "3"); p.chosen != i18n.ZH {
		t.Errorf("key 3 must choose Chinese, got %q", p.chosen)
	}
	// Esc, q and out-of-range digits choose nothing.
	for _, k := range []string{"q", "9", "0"} {
		if p = press(newLanguagePicker(i18n.EN), k); p.chosen != "" {
			t.Errorf("key %q must not choose a language, got %q", k, p.chosen)
		}
	}
	if p = press(newLanguagePicker(i18n.EN), "esc"); p.chosen != "" {
		t.Errorf("esc must not choose, got %q", p.chosen)
	}
}

// Errors are routed to a field by their Russian text but must be shown in the
// selected language.
func TestValidationErrorsAreShownTranslated(t *testing.T) {
	defer i18n.Set("")
	for lang, want := range map[i18n.Lang]string{
		i18n.ZH: "计算机名称超过 15 个字符",
		i18n.ES: "El nombre del equipo tiene más de 15 caracteres",
		i18n.EN: "The computer name is longer than 15 characters",
		i18n.HI: "कंप्यूटर का नाम 15 अक्षरों से अधिक लंबा है",
	} {
		i18n.Set(lang)
		m := modelAt(screens.ScreenAccounts)
		recv, ok := m.screens[m.current].(screens.ErrorReceiver)
		if !ok {
			t.Fatal("the accounts screen must receive validation errors")
		}
		m.screens[m.current] = recv.SetErrors([]string{i18n.M("Имя компьютера длиннее 15 символов")})
		view := m.View()
		if !strings.Contains(view, want) {
			t.Errorf("%s: the error must be shown as %q (routed to the computer-name field):\n%s", lang, want, view)
		}
		if strings.Contains(view, "Имя компьютера длиннее") {
			t.Errorf("%s: the Russian canonical text leaked into the view", lang)
		}
	}
}
