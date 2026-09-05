package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var scriptFormatOptions = []widgets.SelectOption{
	{Value: string(profile.ScriptCmd), Label: ".cmd"},
	{Value: string(profile.ScriptPs1), Label: ".ps1"},
	{Value: string(profile.ScriptReg), Label: ".reg"},
	{Value: string(profile.ScriptVbs), Label: ".vbs"},
}

// defaultUserScriptFormatOptions excludes .vbs: DefaultUser scripts don't
// support it (see profile.CustomScript).
var defaultUserScriptFormatOptions = scriptFormatOptions[:3]

// Scripts is the custom-scripts screen. The CLI/JSON profile supports up to
// 3-4 scripts per category; this screen offers one script per category
// (System, DefaultUser, FirstLogon, UserOnce) — the common case — plus the
// restart-explorer toggle. Profiles with more scripts per category keep
// them when edited here: only the first slot in each category is shown and
// editable, the rest pass through untouched.
type Scripts struct {
	profile *profile.Profile

	systemFormat  widgets.LabeledSelect
	systemContent widgets.LabeledTextArea

	defaultUserFormat  widgets.LabeledSelect
	defaultUserContent widgets.LabeledTextArea

	firstLogonFormat  widgets.LabeledSelect
	firstLogonContent widgets.LabeledTextArea

	userOnceFormat  widgets.LabeledSelect
	userOnceContent widgets.LabeledTextArea

	restartExplorer widgets.Checkbox

	rest struct {
		system, defaultUser, firstLogon, userOnce []profile.CustomScript
	}

	focus int
	bar   widgets.ConfirmBar
}

const scriptsFieldCount = 9 // 4 categories x (format + content) + restart checkbox

// NewScripts builds the scripts screen backed by profile.
func NewScripts(p *profile.Profile) Scripts {
	s := Scripts{
		profile:           p,
		systemFormat:      widgets.NewLabeledSelect("System script format", scriptFormatOptions),
		systemContent:     widgets.NewLabeledTextArea("System script (runs before accounts are created)", ""),
		defaultUserFormat: widgets.NewLabeledSelect("DefaultUser script format", defaultUserScriptFormatOptions),
		defaultUserContent: widgets.NewLabeledTextArea(
			"DefaultUser script (edits the template every new account is created from)", ""),
		firstLogonFormat:  widgets.NewLabeledSelect("FirstLogon script format", scriptFormatOptions),
		firstLogonContent: widgets.NewLabeledTextArea("FirstLogon script (runs once, first logon)", ""),
		userOnceFormat:    widgets.NewLabeledSelect("UserOnce script format", scriptFormatOptions),
		userOnceContent:   widgets.NewLabeledTextArea("UserOnce script (runs once per account, including future ones)", ""),
		restartExplorer:   widgets.Checkbox{Label: "Restart File Explorer after scripts run"},
		bar:               widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	s.systemFormat.SetValue(string(profile.ScriptCmd))
	s.defaultUserFormat.SetValue(string(profile.ScriptCmd))
	s.firstLogonFormat.SetValue(string(profile.ScriptCmd))
	s.userOnceFormat.SetValue(string(profile.ScriptCmd))

	if len(p.SystemScripts) > 0 {
		s.systemFormat.SetValue(string(p.SystemScripts[0].Format))
		s.systemContent.SetValue(p.SystemScripts[0].Content)
		s.rest.system = p.SystemScripts[1:]
	}
	if len(p.DefaultUserScripts) > 0 {
		s.defaultUserFormat.SetValue(string(p.DefaultUserScripts[0].Format))
		s.defaultUserContent.SetValue(p.DefaultUserScripts[0].Content)
		s.rest.defaultUser = p.DefaultUserScripts[1:]
	}
	if len(p.FirstLogonScripts) > 0 {
		s.firstLogonFormat.SetValue(string(p.FirstLogonScripts[0].Format))
		s.firstLogonContent.SetValue(p.FirstLogonScripts[0].Content)
		s.rest.firstLogon = p.FirstLogonScripts[1:]
	}
	if len(p.UserOnceScripts) > 0 {
		s.userOnceFormat.SetValue(string(p.UserOnceScripts[0].Format))
		s.userOnceContent.SetValue(p.UserOnceScripts[0].Content)
		s.rest.userOnce = p.UserOnceScripts[1:]
	}
	s.restartExplorer.Checked = p.RestartExplorerAfterScripts
	s.systemContent.Focus()
	return s
}

// Init returns the cursor-blink command; the focus state itself was already
// set in NewScripts.
func (s Scripts) Init() tea.Cmd {
	return nil
}

func slotOrEmpty(format widgets.LabeledSelect, content widgets.LabeledTextArea, rest []profile.CustomScript) []profile.CustomScript {
	v := content.Value()
	if v == "" {
		return rest
	}
	first := profile.CustomScript{Format: profile.ScriptFormat(format.Value()), Content: v}
	return append([]profile.CustomScript{first}, rest...)
}

func (s *Scripts) sync() {
	s.profile.SystemScripts = slotOrEmpty(s.systemFormat, s.systemContent, s.rest.system)
	s.profile.DefaultUserScripts = slotOrEmpty(s.defaultUserFormat, s.defaultUserContent, s.rest.defaultUser)
	s.profile.FirstLogonScripts = slotOrEmpty(s.firstLogonFormat, s.firstLogonContent, s.rest.firstLogon)
	s.profile.UserOnceScripts = slotOrEmpty(s.userOnceFormat, s.userOnceContent, s.rest.userOnce)
	s.profile.RestartExplorerAfterScripts = s.restartExplorer.Checked
}

func (s *Scripts) setFocus(i int) tea.Cmd {
	s.systemContent.Blur()
	s.defaultUserContent.Blur()
	s.firstLogonContent.Blur()
	s.userOnceContent.Blur()
	s.focus = i
	switch i {
	case 1:
		return s.systemContent.Focus()
	case 3:
		return s.defaultUserContent.Focus()
	case 5:
		return s.firstLogonContent.Focus()
	case 7:
		return s.userOnceContent.Focus()
	}
	return nil
}

// Update handles focus cycling, text/select input, checkbox toggling and
// screen navigation.
func (s Scripts) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			cmd := s.setFocus((s.focus + 1) % scriptsFieldCount)
			s.sync()
			return s, cmd
		case "shift+tab":
			cmd := s.setFocus((s.focus - 1 + scriptsFieldCount) % scriptsFieldCount)
			s.sync()
			return s, cmd
		case " ":
			if s.focus == 8 {
				s.restartExplorer.Checked = !s.restartExplorer.Checked
				s.sync()
				return s, nil
			}
		case "ctrl+n":
			s.sync()
			return s, Navigate(ScreenReview)
		case "esc":
			s.sync()
			return s, Navigate(ScreenDesktop)
		case "ctrl+r":
			s.sync()
			return s, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch s.focus {
	case 0:
		s.systemFormat, cmd = s.systemFormat.Update(msg)
	case 1:
		s.systemContent, cmd = s.systemContent.Update(msg)
	case 2:
		s.defaultUserFormat, cmd = s.defaultUserFormat.Update(msg)
	case 3:
		s.defaultUserContent, cmd = s.defaultUserContent.Update(msg)
	case 4:
		s.firstLogonFormat, cmd = s.firstLogonFormat.Update(msg)
	case 5:
		s.firstLogonContent, cmd = s.firstLogonContent.Update(msg)
	case 6:
		s.userOnceFormat, cmd = s.userOnceFormat.Update(msg)
	case 7:
		s.userOnceContent, cmd = s.userOnceContent.Update(msg)
	}
	s.sync()
	return s, cmd
}

// View renders every category's format selector and content area, plus the
// restart-explorer checkbox.
func (s Scripts) View() string {
	out := s.systemFormat.View() + "\n" + s.systemContent.View() + "\n\n"
	out += s.defaultUserFormat.View() + "\n" + s.defaultUserContent.View() + "\n\n"
	out += s.firstLogonFormat.View() + "\n" + s.firstLogonContent.View() + "\n\n"
	out += s.userOnceFormat.View() + "\n" + s.userOnceContent.View() + "\n\n"
	out += s.restartExplorer.View(s.focus == 8)
	out += "\n\n" + s.bar.View()
	return out
}
