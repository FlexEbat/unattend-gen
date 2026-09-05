package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var desktopIconLabels = map[profile.DesktopIcon]string{
	profile.DesktopIconThisPC:       "This PC",
	profile.DesktopIconUserFiles:    "User's Files",
	profile.DesktopIconNetwork:      "Network",
	profile.DesktopIconRecycleBin:   "Recycle Bin",
	profile.DesktopIconControlPanel: "Control Panel",
	profile.DesktopIconDesktop:      "Desktop",
	profile.DesktopIconDocuments:    "Documents",
	profile.DesktopIconDownloads:    "Downloads",
	profile.DesktopIconMusic:        "Music",
	profile.DesktopIconPictures:     "Pictures",
	profile.DesktopIconVideos:       "Videos",
	profile.DesktopIconGallery:      "Gallery",
	profile.DesktopIconHome:         "Home",
}

var startFolderLabels = map[profile.StartFolder]string{
	profile.StartFolderSettings:       "Settings",
	profile.StartFolderFileExplorer:   "File Explorer",
	profile.StartFolderDocuments:      "Documents",
	profile.StartFolderDownloads:      "Downloads",
	profile.StartFolderMusic:          "Music",
	profile.StartFolderPictures:       "Pictures",
	profile.StartFolderVideos:         "Videos",
	profile.StartFolderNetwork:        "Network",
	profile.StartFolderPersonalFolder: "Personal Folder",
}

// Desktop is the "desktop icons and Start folders" screen. Desktop icon
// visibility is all-or-nothing: a master checkbox opts into customizing
// them at all (unchecked = Profile.DesktopIcons stays nil, Windows' own
// defaults apply); once opted in, every icon gets an explicit shown/hidden
// checkbox. Start folder pins are a plain checklist - an empty selection
// already means "leave Windows' default set alone" (see
// StartFoldersUserOnceCommand), so no master toggle is needed there.
type Desktop struct {
	profile *profile.Profile

	customizeIcons widgets.Checkbox
	iconChecks     []widgets.Checkbox // len(profile.DesktopIcons), only focusable when customizeIcons is checked
	folderChecks   []widgets.Checkbox // len(profile.StartFolders)

	focus int
	bar   widgets.ConfirmBar
}

// NewDesktop builds the desktop screen backed by profile.
func NewDesktop(p *profile.Profile) Desktop {
	d := Desktop{
		profile:        p,
		customizeIcons: widgets.Checkbox{Label: "Customize desktop icon visibility"},
		iconChecks:     make([]widgets.Checkbox, len(profile.DesktopIcons)),
		folderChecks:   make([]widgets.Checkbox, len(profile.StartFolders)),
		bar:            widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	if p.DesktopIcons != nil {
		d.customizeIcons.Checked = true
	}
	for i, ic := range profile.DesktopIcons {
		shown, ok := p.DesktopIcons[ic]
		d.iconChecks[i] = widgets.Checkbox{Label: desktopIconLabels[ic], Checked: ok && shown}
	}
	selectedFolders := make(map[profile.StartFolder]bool, len(p.StartFolders))
	for _, f := range p.StartFolders {
		selectedFolders[f] = true
	}
	for i, f := range profile.StartFolders {
		d.folderChecks[i] = widgets.Checkbox{Label: startFolderLabels[f], Checked: selectedFolders[f]}
	}
	return d
}

// Init is a no-op.
func (d Desktop) Init() tea.Cmd {
	return nil
}

func (d Desktop) fieldCount() int {
	n := 1 + len(d.folderChecks)
	if d.customizeIcons.Checked {
		n += len(d.iconChecks)
	}
	return n
}

func (d *Desktop) sync() {
	if d.customizeIcons.Checked {
		icons := make(map[profile.DesktopIcon]bool, len(profile.DesktopIcons))
		for i, ic := range profile.DesktopIcons {
			icons[ic] = d.iconChecks[i].Checked
		}
		d.profile.DesktopIcons = icons
	} else {
		d.profile.DesktopIcons = nil
	}

	var folders []profile.StartFolder
	for i, c := range d.folderChecks {
		if c.Checked {
			folders = append(folders, profile.StartFolders[i])
		}
	}
	d.profile.StartFolders = folders
}

// checkboxAt returns a pointer to the checkbox at overall focus index i:
// index 0 is customizeIcons, then the icon checkboxes (if visible), then
// the folder checkboxes.
func (d *Desktop) checkboxAt(i int) *widgets.Checkbox {
	if i == 0 {
		return &d.customizeIcons
	}
	i--
	if d.customizeIcons.Checked {
		if i < len(d.iconChecks) {
			return &d.iconChecks[i]
		}
		i -= len(d.iconChecks)
	}
	return &d.folderChecks[i]
}

// Update handles focus cycling, checkbox toggling and screen navigation.
func (d Desktop) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			d.focus = (d.focus + 1) % d.fieldCount()
			return d, nil
		case "shift+tab":
			d.focus = (d.focus - 1 + d.fieldCount()) % d.fieldCount()
			return d, nil
		case " ", "enter":
			c := d.checkboxAt(d.focus)
			c.Checked = !c.Checked
			d.sync()
			if d.focus >= d.fieldCount() {
				d.focus = d.fieldCount() - 1
			}
			return d, nil
		case "ctrl+n":
			d.sync()
			return d, Navigate(ScreenScripts)
		case "esc":
			d.sync()
			return d, Navigate(ScreenAccessibility)
		case "ctrl+r":
			d.sync()
			return d, Navigate(ScreenReview)
		}
	}
	return d, nil
}

// View renders the desktop icon and Start folder checkboxes.
func (d Desktop) View() string {
	out := d.customizeIcons.View(d.focus == 0)
	if d.customizeIcons.Checked {
		out += "\n\n"
		for i, c := range d.iconChecks {
			out += c.View(d.focus == 1+i) + "\n"
		}
	}
	out += "\nPin folders on Start\n\n"
	base := 1
	if d.customizeIcons.Checked {
		base += len(d.iconChecks)
	}
	for i, c := range d.folderChecks {
		out += c.View(d.focus == base+i) + "\n"
	}
	out += "\n" + d.bar.View()
	return out
}
