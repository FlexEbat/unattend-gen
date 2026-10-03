package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var taskbarSearchOptions = []widgets.SelectOption{
	{Value: "", Label: "Windows default (search box)"},
	{Value: string(profile.TaskbarSearchModeHide), Label: "Hide"},
	{Value: string(profile.TaskbarSearchModeIcon), Label: "Icon only"},
	{Value: string(profile.TaskbarSearchModeBox), Label: "Search box"},
	{Value: string(profile.TaskbarSearchModeLabel), Label: "Icon with label"},
}

var startPinsModeOptions = []widgets.SelectOption{
	{Value: string(profile.StartPinsModeDefault), Label: "Windows default"},
	{Value: string(profile.StartPinsModeEmpty), Label: "Empty (no pins)"},
	{Value: string(profile.StartPinsModeCustom), Label: "Custom (raw JSON)"},
}

var startTilesModeOptions = []widgets.SelectOption{
	{Value: string(profile.StartTilesModeDefault), Label: "Windows default"},
	{Value: string(profile.StartTilesModeEmpty), Label: "Empty (no tiles)"},
	{Value: string(profile.StartTilesModeCustom), Label: "Custom (raw LayoutModification.xml)"},
}

var taskbarIconsModeOptions = []widgets.SelectOption{
	{Value: string(profile.TaskbarIconsModeDefault), Label: "Windows default"},
	{Value: string(profile.TaskbarIconsModeEmpty), Label: "Empty (no pinned icons)"},
	{Value: string(profile.TaskbarIconsModeCustom), Label: "Custom (raw taskbar layout XML)"},
}

// Taskbar is the Start menu/taskbar screen: 5 simple checkboxes, a
// taskbar-search-mode select, and Start pins (Windows 11), tiles
// (Windows 10) and taskbar icons, each with its own default/empty/custom
// mode select. The custom text area only appears when its mode is Custom.
type Taskbar struct {
	profile *profile.Profile

	disableWidgets     widgets.Checkbox
	leftTaskbar        widgets.Checkbox
	hideTaskView       widgets.Checkbox
	disableBingResults widgets.Checkbox
	showAllTrayIcons   widgets.Checkbox
	taskbarSearch      widgets.LabeledSelect
	startPinsMode      widgets.LabeledSelect
	startPinsJSON      widgets.LabeledTextArea
	startTilesMode     widgets.LabeledSelect
	startTilesXML      widgets.LabeledTextArea
	taskbarIconsMode   widgets.LabeledSelect
	taskbarIconsXML    widgets.LabeledTextArea

	focus int
	bar   widgets.ConfirmBar
}

// NewTaskbar builds the taskbar screen backed by profile.
func NewTaskbar(p *profile.Profile) Taskbar {
	t := Taskbar{
		profile:            p,
		disableWidgets:     widgets.Checkbox{Label: "Disable Widgets", Checked: p.SystemTweaks.DisableWidgets},
		leftTaskbar:        widgets.Checkbox{Label: "Left-align taskbar (Windows 11)", Checked: p.SystemTweaks.LeftTaskbar},
		hideTaskView:       widgets.Checkbox{Label: "Hide Task View button", Checked: p.SystemTweaks.HideTaskViewButton},
		disableBingResults: widgets.Checkbox{Label: "Disable Bing results in search", Checked: p.SystemTweaks.DisableBingResults},
		showAllTrayIcons:   widgets.Checkbox{Label: "Always show all tray icons", Checked: p.SystemTweaks.ShowAllTrayIcons},
		taskbarSearch:      widgets.NewLabeledSelect("Taskbar search box", taskbarSearchOptions),
		startPinsMode:      widgets.NewLabeledSelect("Start pins (Windows 11)", startPinsModeOptions),
		startPinsJSON:      widgets.NewLabeledTextArea("Start pins JSON (pinnedList)", ""),
		startTilesMode:     widgets.NewLabeledSelect("Start tiles (Windows 10)", startTilesModeOptions),
		startTilesXML:      widgets.NewLabeledTextArea("Start tiles LayoutModification.xml", ""),
		taskbarIconsMode:   widgets.NewLabeledSelect("Taskbar icons (pinned apps)", taskbarIconsModeOptions),
		taskbarIconsXML:    widgets.NewLabeledTextArea("Taskbar layout XML (LayoutModification)", ""),
		bar:                widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	t.taskbarSearch.SetValue(string(p.TaskbarSearch))
	t.startPinsMode.SetValue(string(p.StartPins.Mode))
	if p.StartPins.JSON != nil {
		t.startPinsJSON.SetValue(*p.StartPins.JSON)
	}
	t.startTilesMode.SetValue(string(p.StartTiles.Mode))
	if p.StartTiles.XML != nil {
		t.startTilesXML.SetValue(*p.StartTiles.XML)
	}
	t.taskbarIconsMode.SetValue(string(p.TaskbarIcons.Mode))
	if p.TaskbarIcons.XML != nil {
		t.taskbarIconsXML.SetValue(*p.TaskbarIcons.XML)
	}
	return t
}

// Init is a no-op.
func (t Taskbar) Init() tea.Cmd {
	return nil
}

func (t Taskbar) startPinsCustom() bool {
	return t.startPinsMode.Value() == string(profile.StartPinsModeCustom)
}

func (t Taskbar) startTilesCustom() bool {
	return t.startTilesMode.Value() == string(profile.StartTilesModeCustom)
}

func (t Taskbar) taskbarIconsCustom() bool {
	return t.taskbarIconsMode.Value() == string(profile.TaskbarIconsModeCustom)
}

// Field layout: 0-4 checkboxes, 5 taskbarSearch, 6 startPinsMode,
// [7 startPinsJSON if custom], next startTilesMode, [next startTilesXML if custom],
// next taskbarIconsMode, [next taskbarIconsXML if custom].
func (t Taskbar) startPinsModeIdx() int { return 6 }
func (t Taskbar) startPinsJSONIdx() int { return 7 }
func (t Taskbar) startTilesModeIdx() int {
	if t.startPinsCustom() {
		return 8
	}
	return 7
}
func (t Taskbar) startTilesXMLIdx() int { return t.startTilesModeIdx() + 1 }
func (t Taskbar) taskbarIconsModeIdx() int {
	if t.startTilesCustom() {
		return t.startTilesXMLIdx() + 1
	}
	return t.startTilesModeIdx() + 1
}
func (t Taskbar) taskbarIconsXMLIdx() int { return t.taskbarIconsModeIdx() + 1 }

func (t Taskbar) fieldCount() int {
	n := t.taskbarIconsModeIdx() + 1
	if t.taskbarIconsCustom() {
		n++
	}
	return n
}

// setFocus moves focus to field i. The three custom text areas only accept
// input while focused (textarea ignores keys otherwise), so they are blurred
// and the active one is focused explicitly.
func (t *Taskbar) setFocus(i int) tea.Cmd {
	t.startPinsJSON.Blur()
	t.startTilesXML.Blur()
	t.taskbarIconsXML.Blur()
	t.focus = i
	switch {
	case t.startPinsCustom() && i == t.startPinsJSONIdx():
		return t.startPinsJSON.Focus()
	case t.startTilesCustom() && i == t.startTilesXMLIdx():
		return t.startTilesXML.Focus()
	case t.taskbarIconsCustom() && i == t.taskbarIconsXMLIdx():
		return t.taskbarIconsXML.Focus()
	}
	return nil
}

func (t *Taskbar) sync() {
	t.profile.SystemTweaks.DisableWidgets = t.disableWidgets.Checked
	t.profile.SystemTweaks.LeftTaskbar = t.leftTaskbar.Checked
	t.profile.SystemTweaks.HideTaskViewButton = t.hideTaskView.Checked
	t.profile.SystemTweaks.DisableBingResults = t.disableBingResults.Checked
	t.profile.SystemTweaks.ShowAllTrayIcons = t.showAllTrayIcons.Checked
	t.profile.TaskbarSearch = profile.TaskbarSearchMode(t.taskbarSearch.Value())

	t.profile.StartPins.Mode = profile.StartPinsMode(t.startPinsMode.Value())
	if t.startPinsCustom() {
		v := t.startPinsJSON.Value()
		t.profile.StartPins.JSON = &v
	} else {
		t.profile.StartPins.JSON = nil
	}

	t.profile.StartTiles.Mode = profile.StartTilesMode(t.startTilesMode.Value())
	if t.startTilesCustom() {
		v := t.startTilesXML.Value()
		t.profile.StartTiles.XML = &v
	} else {
		t.profile.StartTiles.XML = nil
	}

	t.profile.TaskbarIcons.Mode = profile.TaskbarIconsMode(t.taskbarIconsMode.Value())
	if t.taskbarIconsCustom() {
		v := t.taskbarIconsXML.Value()
		t.profile.TaskbarIcons.XML = &v
	} else {
		t.profile.TaskbarIcons.XML = nil
	}
}

// Update handles focus cycling, checkbox/select/text input and navigation.
func (t Taskbar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			cmd := t.setFocus((t.focus + 1) % t.fieldCount())
			t.sync()
			return t, cmd
		case "shift+tab":
			cmd := t.setFocus((t.focus - 1 + t.fieldCount()) % t.fieldCount())
			t.sync()
			return t, cmd
		case " ":
			switch t.focus {
			case 0:
				t.disableWidgets.Checked = !t.disableWidgets.Checked
			case 1:
				t.leftTaskbar.Checked = !t.leftTaskbar.Checked
			case 2:
				t.hideTaskView.Checked = !t.hideTaskView.Checked
			case 3:
				t.disableBingResults.Checked = !t.disableBingResults.Checked
			case 4:
				t.showAllTrayIcons.Checked = !t.showAllTrayIcons.Checked
			}
			t.sync()
			return t, nil
		case "ctrl+n":
			t.sync()
			return t, Navigate(ScreenAdvanced)
		case "esc":
			t.sync()
			return t, Navigate(ScreenVisualEffects)
		case "ctrl+r":
			t.sync()
			return t, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch {
	case t.focus == 5:
		t.taskbarSearch, cmd = t.taskbarSearch.Update(msg)
	case t.focus == t.startPinsModeIdx():
		t.startPinsMode, cmd = t.startPinsMode.Update(msg)
	case t.startPinsCustom() && t.focus == t.startPinsJSONIdx():
		t.startPinsJSON, cmd = t.startPinsJSON.Update(msg)
	case t.focus == t.startTilesModeIdx():
		t.startTilesMode, cmd = t.startTilesMode.Update(msg)
	case t.startTilesCustom() && t.focus == t.startTilesXMLIdx():
		t.startTilesXML, cmd = t.startTilesXML.Update(msg)
	case t.focus == t.taskbarIconsModeIdx():
		t.taskbarIconsMode, cmd = t.taskbarIconsMode.Update(msg)
	case t.taskbarIconsCustom() && t.focus == t.taskbarIconsXMLIdx():
		t.taskbarIconsXML, cmd = t.taskbarIconsXML.Update(msg)
	}
	t.sync()
	if t.focus >= t.fieldCount() {
		t.focus = t.fieldCount() - 1
	}
	return t, cmd
}

// View renders the checkboxes, taskbar search select, and Start pins/tiles/
// taskbar icons mode selects with their custom text areas when applicable.
func (t Taskbar) View() string {
	out := t.disableWidgets.View(t.focus == 0) + "\n"
	out += t.leftTaskbar.View(t.focus == 1) + "\n"
	out += t.hideTaskView.View(t.focus == 2) + "\n"
	out += t.disableBingResults.View(t.focus == 3) + "\n"
	out += t.showAllTrayIcons.View(t.focus == 4) + "\n"
	out += "\n" + t.taskbarSearch.View()
	out += "\n\n" + t.startPinsMode.View()
	if t.startPinsCustom() {
		out += "\n\n" + t.startPinsJSON.View()
	}
	out += "\n\n" + t.startTilesMode.View()
	if t.startTilesCustom() {
		out += "\n\n" + t.startTilesXML.View()
	}
	out += "\n\n" + t.taskbarIconsMode.View()
	if t.taskbarIconsCustom() {
		out += "\n\n" + t.taskbarIconsXML.View()
	}
	out += "\n\n" + t.bar.View()
	return out
}
