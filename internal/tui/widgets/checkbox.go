package widgets

import "github.com/FlexEbat/unattend-gen/internal/i18n"

// Checkbox is a boolean field rendered as [x]/[ ]. The screen holding it
// decides when it has focus and toggles Checked; Checkbox only renders.
type Checkbox struct {
	Label   string
	Checked bool
}

// View renders the box and label, with a focus marker when focused is true.
func (c Checkbox) View(focused bool) string {
	box := "[ ]"
	if c.Checked {
		box = "[x]"
	}
	if focused {
		return "> " + box + " " + i18n.T(c.Label)
	}
	return "  " + box + " " + i18n.T(c.Label)
}
