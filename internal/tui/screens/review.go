package screens

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen"
)

// Review previews the final autounattend.xml and saves the profile and/or
// the answer file to disk.
type Review struct {
	profile *profile.Profile
	xmlText string
	err     string
	status  string
	bar     widgets.ConfirmBar
}

// NewReview builds the review screen backed by profile.
func NewReview(p *profile.Profile) Review {
	r := Review{
		profile: p,
		bar:     widgets.NewConfirmBar("s: save profile .json", "x: export .xml", "Esc: back"),
	}
	r.build()
	return r
}

func (r *Review) build() {
	if errs := ValidateForReview(r.profile); len(errs) > 0 {
		r.err = errs[0]
		r.xmlText = ""
		return
	}
	xmlStr, err := xmlgen.BuildAnswerFile(r.profile)
	if err != nil {
		r.err = err.Error()
		return
	}
	r.err = ""
	r.xmlText = xmlStr
}

// Init rebuilds the preview.
func (r Review) Init() tea.Cmd {
	return nil
}

// Update handles saving to disk and navigating back.
func (r Review) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return r, Navigate(ScreenApps)
		case "s":
			r.build()
			if r.err != "" {
				r.status = ""
				return r, nil
			}
			path := r.profile.Name + ".json"
			if err := profile.SaveProfile(r.profile, path); err != nil {
				r.status = err.Error()
			} else {
				r.status = "saved " + path
			}
			return r, nil
		case "x":
			r.build()
			if r.err != "" {
				r.status = ""
				return r, nil
			}
			if err := os.WriteFile("autounattend.xml", []byte(r.xmlText), 0o644); err != nil {
				r.status = err.Error()
			} else {
				r.status = "saved autounattend.xml"
			}
			return r, nil
		}
	}
	return r, nil
}

// View renders the validation error or the XML preview, plus any save status.
func (r Review) View() string {
	out := ""
	if r.err != "" {
		out = "Profile is not valid yet:\n" + r.err
	} else {
		out = r.xmlText
	}
	if r.status != "" {
		out += "\n\n" + r.status
	}
	out += "\n\n" + r.bar.View()
	return out
}
