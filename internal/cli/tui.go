package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui [profile.json]",
		Short: "Open the interactive TUI, optionally starting from an existing profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var initial *profile.Profile
			if len(args) == 1 {
				loaded, err := profile.LoadProfile(args[0])
				if err != nil {
					return fmt.Errorf("load profile %s: %w", args[0], err)
				}
				initial = loaded
			}

			p := tea.NewProgram(tui.NewModel(initial))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("run tui: %w", err)
			}
			return nil
		},
	}
}
