package screens

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen"
)

// answerFileName is the file the "x" key exports to (in the current directory).
const answerFileName = "autounattend.xml"

// reviewChrome is the number of terminal rows kept for the status line and
// the key bar below the scrollable XML preview.
const reviewChrome = 6

// Review previews the final autounattend.xml and saves the profile and/or
// the answer file to disk. The preview scrolls once the terminal size is
// known; before that (tests, first frame) the whole text is shown.
type Review struct {
	profile          *profile.Profile
	xmlText          string
	err              string
	status           string
	bar              widgets.ConfirmBar
	vp               viewport.Model
	sized            bool
	confirmOverwrite bool
}

// NewReview builds the review screen backed by profile.
func NewReview(p *profile.Profile) Review {
	r := Review{
		profile: p,
		bar:     widgets.NewConfirmBar("s: save profile .json", "x: export .xml", "↑/↓ PgUp/PgDn: scroll", "Esc: back"),
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
	r.vp.SetContent(xmlStr)
}

// Init rebuilds the preview.
func (r Review) Init() tea.Cmd {
	return nil
}

// Update handles saving to disk, scrolling the preview and navigating back.
func (r Review) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		r.vp.Width = size.Width
		r.vp.Height = max(3, size.Height-reviewChrome)
		r.sized = true
		r.vp.SetContent(r.xmlText)
		return r, nil
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return r, nil
	}
	key := keyMsg.String()
	if r.confirmOverwrite && key != "x" { // any other key cancels the prompt
		r.confirmOverwrite = false
		r.status = ""
	}
	switch key {
	case "esc":
		return r, Navigate(ScreenScripts)
	case "s":
		r.build()
		if r.err != "" {
			r.status = ""
			return r, nil
		}
		path, err := profile.ProfilePath(r.profile.Name)
		if err != nil {
			r.status = err.Error()
			return r, nil
		}
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
		// Like `generate` without --force: never replace an existing answer
		// file without asking.
		if _, err := os.Stat(answerFileName); err == nil && !r.confirmOverwrite {
			r.confirmOverwrite = true
			r.status = answerFileName + " already exists: press x again to overwrite, any other key to cancel"
			return r, nil
		}
		r.confirmOverwrite = false
		if err := profile.WriteFileSecure(answerFileName, []byte(r.xmlText)); err != nil {
			r.status = err.Error()
		} else {
			r.status = "saved " + answerFileName
		}
		return r, nil
	}
	if r.sized && r.err == "" {
		var cmd tea.Cmd
		r.vp, cmd = r.vp.Update(msg)
		return r, cmd
	}
	return r, nil
}

// View renders the validation error or the XML preview, plus any save status.
func (r Review) View() string {
	out := ""
	if r.err != "" {
		out = "Profile is not valid yet:\n" + r.err
	} else if r.sized {
		out = r.vp.View()
		if r.vp.TotalLineCount() > r.vp.Height {
			out += fmt.Sprintf("\n-- %d%% --", int(r.vp.ScrollPercent()*100))
		}
	} else {
		out = r.xmlText
	}
	if r.status != "" {
		out += "\n\n" + r.status
	}
	out += "\n\n" + r.bar.View()
	return out
}
