package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var visualEffectsModeOptions = []widgets.SelectOption{
	{Value: "", Label: "Windows default (let Windows decide)"},
	{Value: string(profile.VisualEffectsModeBestAppearance), Label: "Best appearance (all effects on)"},
	{Value: string(profile.VisualEffectsModeBestPerformance), Label: "Best performance (all effects off)"},
	{Value: string(profile.VisualEffectsModeCustom), Label: "Custom"},
}

var visualEffectLabels = map[profile.VisualEffect]string{
	profile.EffectControlAnimations:       "Animate controls inside windows",
	profile.EffectAnimateMinMax:           "Animate windows when minimizing/maximizing",
	profile.EffectTaskbarAnimations:       "Animations in the taskbar",
	profile.EffectDWMAeroPeekEnabled:      "Peek at desktop when hovering Show desktop",
	profile.EffectMenuAnimation:           "Animate menus",
	profile.EffectTooltipAnimation:        "Fade or slide tooltips",
	profile.EffectSelectionFade:           "Fade out menu items after clicking",
	profile.EffectDWMSaveThumbnailEnabled: "Save taskbar thumbnail previews",
	profile.EffectCursorShadow:            "Show shadow under mouse pointer",
	profile.EffectListviewShadow:          "Show shadows under icon labels on desktop",
	profile.EffectThumbnailsOrIcon:        "Show thumbnails instead of icons",
	profile.EffectListviewAlphaSelect:     "Use translucent selection rectangle",
	profile.EffectDragFullWindows:         "Show window contents while dragging",
	profile.EffectComboBoxAnimation:       "Slide open combo boxes",
	profile.EffectFontSmoothing:           "Smooth edges of screen fonts",
	profile.EffectListBoxSmoothScrolling:  "Smooth-scroll list boxes",
	profile.EffectDropShadow:              "Show shadows under windows",
}

// VisualEffects is the Windows "Performance Options" screen: a mode select
// (default/best appearance/best performance/custom), with all 17
// individual effect checkboxes appearing only when Custom is selected.
// Choosing Custom always writes an explicit value for every effect (not a
// partial map) — same "picking the mode implies setting the whole set"
// convention as the Desktop screen's icon-visibility customization.
type VisualEffects struct {
	profile *profile.Profile
	mode    widgets.LabeledSelect
	checks  []widgets.Checkbox
	focus   int
	bar     widgets.ConfirmBar
}

// NewVisualEffects builds the visual effects screen backed by profile.
func NewVisualEffects(p *profile.Profile) VisualEffects {
	v := VisualEffects{
		profile: p,
		mode:    widgets.NewLabeledSelect("Visual effects", visualEffectsModeOptions),
		checks:  make([]widgets.Checkbox, len(profile.VisualEffects)),
		bar:     widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	v.mode.SetValue(string(p.VisualEffects.Mode))
	for i, e := range profile.VisualEffects {
		on, ok := p.VisualEffects.Custom[e]
		v.checks[i] = widgets.Checkbox{Label: visualEffectLabels[e], Checked: ok && on}
	}
	return v
}

// Init is a no-op.
func (v VisualEffects) Init() tea.Cmd {
	return nil
}

func (v VisualEffects) isCustom() bool {
	return v.mode.Value() == string(profile.VisualEffectsModeCustom)
}

func (v VisualEffects) fieldCount() int {
	if v.isCustom() {
		return 1 + len(v.checks)
	}
	return 1
}

func (v *VisualEffects) sync() {
	v.profile.VisualEffects.Mode = profile.VisualEffectsMode(v.mode.Value())
	if v.isCustom() {
		custom := make(map[profile.VisualEffect]bool, len(profile.VisualEffects))
		for i, e := range profile.VisualEffects {
			custom[e] = v.checks[i].Checked
		}
		v.profile.VisualEffects.Custom = custom
	} else {
		v.profile.VisualEffects.Custom = nil
	}
}

// Update handles focus cycling, select/checkbox input and screen navigation.
func (v VisualEffects) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			v.focus = (v.focus + 1) % v.fieldCount()
			return v, nil
		case "shift+tab":
			v.focus = (v.focus - 1 + v.fieldCount()) % v.fieldCount()
			return v, nil
		case " ":
			if v.isCustom() && v.focus >= 1 {
				i := v.focus - 1
				v.checks[i].Checked = !v.checks[i].Checked
				v.sync()
			}
			return v, nil
		case "ctrl+n":
			v.sync()
			return v, Navigate(ScreenTaskbar)
		case "esc":
			v.sync()
			return v, Navigate(ScreenDesktop)
		case "ctrl+r":
			v.sync()
			return v, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	if v.focus == 0 {
		v.mode, cmd = v.mode.Update(msg)
	}
	v.sync()
	return v, cmd
}

// View renders the mode select and, when Custom, all 17 effect checkboxes.
func (v VisualEffects) View() string {
	out := v.mode.View()
	if v.isCustom() {
		out += "\n\n"
		for i, c := range v.checks {
			out += c.View(v.focus == 1+i) + "\n"
		}
	}
	out += "\n" + v.bar.View()
	return out
}
