package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

const profilesDir = "profiles"

const newProfileValue = "__new__"

// Welcome is the first screen: pick an existing profile from ./profiles or
// start a new one.
type Welcome struct {
	profile     *profile.Profile
	choices     widgets.LabeledSelect
	nameInput   widgets.LabeledInput
	creatingNew bool
	bar         widgets.ConfirmBar
	err         string
}

// NewWelcome builds the welcome screen backed by profile.
func NewWelcome(p *profile.Profile) Welcome {
	options := []widgets.SelectOption{{Value: newProfileValue, Label: "New profile"}}
	if paths, err := profile.ListProfiles(profilesDir); err == nil {
		for _, path := range paths {
			options = append(options, widgets.SelectOption{Value: path, Label: path})
		}
	}
	return Welcome{
		profile:   p,
		choices:   widgets.NewLabeledSelect("Choose a profile", options),
		nameInput: widgets.NewLabeledInput("New profile name", "demo"),
		bar:       widgets.NewConfirmBar("↑/↓: select", "Enter: confirm", "Ctrl+R: review"),
	}
}

// Init focuses the selector.
func (w Welcome) Init() tea.Cmd {
	return nil
}

// Update handles selection and, for a new profile, the name field.
func (w Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+r":
			return w, Navigate(ScreenReview)
		case "enter":
			if w.creatingNew {
				name := w.nameInput.Value()
				if name == "" {
					w.err = "введите имя профиля"
					return w, nil
				}
				w.profile.Name = name
				return w, Navigate(ScreenLanguage)
			}
			switch w.choices.Value() {
			case newProfileValue, "":
				w.creatingNew = true
				return w, w.nameInput.Focus()
			default:
				loaded, err := profile.LoadProfile(w.choices.Value())
				if err != nil {
					w.err = err.Error()
					return w, nil
				}
				*w.profile = *loaded
				return w, Navigate(ScreenLanguage)
			}
		}
	}

	var cmd tea.Cmd
	if w.creatingNew {
		w.nameInput, cmd = w.nameInput.Update(msg)
	} else {
		w.choices, cmd = w.choices.Update(msg)
	}
	return w, cmd
}

// View renders the current sub-mode.
func (w Welcome) View() string {
	out := "unattend-gen\n\n"
	if w.creatingNew {
		out += w.nameInput.View()
	} else {
		out += w.choices.View()
	}
	if w.err != "" {
		out += "\n" + w.err
	}
	out += "\n\n" + w.bar.View()
	return out
}
