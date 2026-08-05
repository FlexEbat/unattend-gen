package widgets

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// MaxAccounts is the maximum number of rows AccountsTable accepts, matching
// the profile validation rule (section 8: max 5 accounts).
const MaxAccounts = 5

// AccountsTable is an editable table of up to MaxAccounts local accounts.
// Editing individual fields is done by the screen (via a small sub-form);
// AccountsTable only displays rows and reports which one is selected.
type AccountsTable struct {
	accounts []profile.UserAccount
	table    table.Model
}

// NewAccountsTable builds an empty accounts table.
func NewAccountsTable() AccountsTable {
	cols := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Display name", Width: 20},
		{Title: "Group", Width: 16},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(MaxAccounts+1),
	)
	return AccountsTable{table: t}
}

// Accounts returns the current rows.
func (a AccountsTable) Accounts() []profile.UserAccount {
	return a.accounts
}

// SetAccounts replaces all rows.
func (a *AccountsTable) SetAccounts(accounts []profile.UserAccount) {
	a.accounts = accounts
	a.refresh()
}

// Add appends an account if there is room. Returns false if the table is
// already at MaxAccounts.
func (a *AccountsTable) Add(account profile.UserAccount) bool {
	if len(a.accounts) >= MaxAccounts {
		return false
	}
	a.accounts = append(a.accounts, account)
	a.refresh()
	return true
}

// ReplaceSelected overwrites the currently selected account, if any.
func (a *AccountsTable) ReplaceSelected(account profile.UserAccount) {
	i := a.table.Cursor()
	if i < 0 || i >= len(a.accounts) {
		return
	}
	a.accounts[i] = account
	a.refresh()
}

// RemoveSelected deletes the currently selected account, if any.
func (a *AccountsTable) RemoveSelected() {
	i := a.table.Cursor()
	if i < 0 || i >= len(a.accounts) {
		return
	}
	a.accounts = append(a.accounts[:i], a.accounts[i+1:]...)
	a.refresh()
}

// Selected returns the currently selected account and whether one exists.
func (a AccountsTable) Selected() (profile.UserAccount, bool) {
	i := a.table.Cursor()
	if i < 0 || i >= len(a.accounts) {
		return profile.UserAccount{}, false
	}
	return a.accounts[i], true
}

func (a *AccountsTable) refresh() {
	rows := make([]table.Row, len(a.accounts))
	for i, acc := range a.accounts {
		display := ""
		if acc.DisplayName != nil {
			display = *acc.DisplayName
		}
		rows[i] = table.Row{acc.Name, display, string(acc.Group)}
	}
	a.table.SetRows(rows)
}

// Update forwards msg to the underlying table.
func (a AccountsTable) Update(msg tea.Msg) (AccountsTable, tea.Cmd) {
	var cmd tea.Cmd
	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

// View renders the table.
func (a AccountsTable) View() string {
	return a.table.View()
}
