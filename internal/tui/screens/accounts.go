package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var firstLogonOptions = []widgets.SelectOption{
	{Value: string(profile.FirstLogonNone), Label: "None"},
	{Value: string(profile.FirstLogonFirstCreatedAccount), Label: "First created account"},
	{Value: string(profile.FirstLogonBuiltinAdmin), Label: "Built-in Administrator"},
}

var groupOptions = []widgets.SelectOption{
	{Value: string(profile.GroupAdministrators), Label: "Administrators"},
	{Value: string(profile.GroupUsers), Label: "Users"},
}

// Accounts is the computer name / local accounts / first logon screen.
type Accounts struct {
	profile       *profile.Profile
	computerName  widgets.LabeledInput
	table         widgets.AccountsTable
	firstLogon    widgets.LabeledSelect
	adminPassword widgets.PasswordInput
	focus         int
	bar           widgets.ConfirmBar
	err           string

	formMode     bool
	formEditing  bool
	formName     widgets.LabeledInput
	formDisplay  widgets.LabeledInput
	formPassword widgets.PasswordInput
	formGroup    widgets.LabeledSelect
	formFocus    int
}

// NewAccounts builds the accounts screen backed by profile.
func NewAccounts(p *profile.Profile) Accounts {
	a := Accounts{
		profile:       p,
		computerName:  widgets.NewLabeledInput("Computer name (optional)", "leave empty for a random name"),
		table:         widgets.NewAccountsTable(),
		firstLogon:    widgets.NewLabeledSelect("First logon", firstLogonOptions),
		adminPassword: widgets.NewPasswordInput("Administrator password"),
		bar:           widgets.NewConfirmBar("Tab: focus", "a: add", "e: edit", "d: delete", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
		formName:      widgets.NewLabeledInput("Name", "alice"),
		formDisplay:   widgets.NewLabeledInput("Display name (optional)", ""),
		formPassword:  widgets.NewPasswordInput("Password (optional)"),
		formGroup:     widgets.NewLabeledSelect("Group", groupOptions),
	}
	if p.ComputerName != nil {
		a.computerName.SetValue(*p.ComputerName)
	}
	a.table.SetAccounts(p.Accounts)
	a.firstLogon.SetValue(string(p.FirstLogon.Mode))
	if p.FirstLogon.BuiltinAdministratorPassword != nil {
		a.adminPassword.SetValue(*p.FirstLogon.BuiltinAdministratorPassword)
	}
	return a
}

// Init focuses the computer name field.
func (a Accounts) Init() tea.Cmd {
	return a.computerName.Focus()
}

func (a Accounts) fieldCount() int {
	if profile.FirstLogonMode(a.firstLogon.Value()) == profile.FirstLogonBuiltinAdmin {
		return 4
	}
	return 3
}

func (a *Accounts) setFocus(i int) tea.Cmd {
	a.computerName.Blur()
	a.adminPassword.Blur()
	a.focus = i
	if i == 0 {
		return a.computerName.Focus()
	}
	if i == 3 {
		return a.adminPassword.Focus()
	}
	return nil
}

func (a *Accounts) sync() {
	name := a.computerName.Value()
	if name == "" {
		a.profile.ComputerName = nil
	} else {
		a.profile.ComputerName = &name
	}
	a.profile.Accounts = a.table.Accounts()
	a.profile.FirstLogon.Mode = profile.FirstLogonMode(a.firstLogon.Value())
	if a.profile.FirstLogon.Mode == profile.FirstLogonBuiltinAdmin {
		pwd := a.adminPassword.Value()
		a.profile.FirstLogon.BuiltinAdministratorPassword = &pwd
	} else {
		a.profile.FirstLogon.BuiltinAdministratorPassword = nil
	}
}

func (a *Accounts) openForm(editing bool) {
	a.formMode = true
	a.formEditing = editing
	a.formFocus = 0
	if editing {
		acc, ok := a.table.Selected()
		if ok {
			a.formName.SetValue(acc.Name)
			if acc.DisplayName != nil {
				a.formDisplay.SetValue(*acc.DisplayName)
			} else {
				a.formDisplay.SetValue("")
			}
			if acc.Password != nil {
				a.formPassword.SetValue(*acc.Password)
			} else {
				a.formPassword.SetValue("")
			}
			a.formGroup.SetValue(string(acc.Group))
		}
	} else {
		a.formName.SetValue("")
		a.formDisplay.SetValue("")
		a.formPassword.SetValue("")
		a.formGroup.SetValue(string(profile.GroupUsers))
	}
}

func (a *Accounts) submitForm() {
	name := a.formName.Value()
	if name == "" {
		a.err = "имя учётной записи не может быть пустым"
		return
	}
	var display *string
	if v := a.formDisplay.Value(); v != "" {
		display = &v
	}
	var password *string
	if v := a.formPassword.Value(); v != "" {
		password = &v
	}
	account := profile.UserAccount{
		Name:        name,
		DisplayName: display,
		Password:    password,
		Group:       profile.AccountGroup(a.formGroup.Value()),
	}
	if a.formEditing {
		a.table.ReplaceSelected(account)
	} else if !a.table.Add(account) {
		a.err = "не более 5 учётных записей"
		return
	}
	a.formMode = false
	a.err = ""
}

// Update handles the main screen and, while open, the add/edit account form.
func (a Accounts) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.formMode {
		return a.updateForm(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			cmd := a.setFocus((a.focus + 1) % a.fieldCount())
			a.sync()
			return a, cmd
		case "shift+tab":
			cmd := a.setFocus((a.focus - 1 + a.fieldCount()) % a.fieldCount())
			a.sync()
			return a, cmd
		case "a":
			if a.focus == 1 {
				a.openForm(false)
				return a, nil
			}
		case "e":
			if a.focus == 1 {
				if _, ok := a.table.Selected(); ok {
					a.openForm(true)
				}
				return a, nil
			}
		case "d":
			if a.focus == 1 {
				a.table.RemoveSelected()
				a.sync()
				return a, nil
			}
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenTweaks)
		case "esc":
			a.sync()
			return a, Navigate(ScreenLanguage)
		case "ctrl+r":
			a.sync()
			return a, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch a.focus {
	case 0:
		a.computerName, cmd = a.computerName.Update(msg)
	case 1:
		a.table, cmd = a.table.Update(msg)
	case 2:
		a.firstLogon, cmd = a.firstLogon.Update(msg)
	case 3:
		a.adminPassword, cmd = a.adminPassword.Update(msg)
	}
	a.sync()
	return a, cmd
}

func (a Accounts) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	const formFieldCount = 4
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			a.formFocus = (a.formFocus + 1) % formFieldCount
			return a, nil
		case "shift+tab":
			a.formFocus = (a.formFocus - 1 + formFieldCount) % formFieldCount
			return a, nil
		case "enter":
			a.submitForm()
			a.sync()
			return a, nil
		case "esc":
			a.formMode = false
			return a, nil
		}
	}

	var cmd tea.Cmd
	switch a.formFocus {
	case 0:
		a.formName, cmd = a.formName.Update(msg)
	case 1:
		a.formDisplay, cmd = a.formDisplay.Update(msg)
	case 2:
		a.formPassword, cmd = a.formPassword.Update(msg)
	case 3:
		a.formGroup, cmd = a.formGroup.Update(msg)
	}
	return a, cmd
}

// View renders the main screen or, while open, the add/edit account form.
func (a Accounts) View() string {
	if a.formMode {
		out := a.formName.View() + "\n\n" + a.formDisplay.View() + "\n\n" + a.formPassword.View() + "\n\n" + a.formGroup.View()
		out += "\n\nEnter: save | Esc: cancel"
		if a.err != "" {
			out += "\n" + a.err
		}
		return out
	}

	out := a.computerName.View() + "\n\n" + a.table.View() + "\n\n" + a.firstLogon.View()
	if profile.FirstLogonMode(a.firstLogon.Value()) == profile.FirstLogonBuiltinAdmin {
		out += "\n\n" + a.adminPassword.View()
	}
	if a.err != "" {
		out += "\n" + a.err
	}
	out += "\n\n" + a.bar.View()
	return out
}
