package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
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

			// Every start asks for the language unless --lang or
			// UNATTEND_GEN_LANG already chose one. Screens are built after
			// this point, so they come out in the chosen language.
			if i18n.Current() == "" {
				l, err := tui.PickLanguage(i18n.FromEnv())
				if err != nil {
					return err
				}
				i18n.Set(l)
			}

			var model tea.Model = tui.NewModel(initial)
			// UNATTEND_GEN_TRACE=<file> logs message timings (never key text)
			// to diagnose input lag; see tui.Tracer.
			if path := os.Getenv("UNATTEND_GEN_TRACE"); path != "" {
				// The path is chosen by the user on purpose (like -o); mode is 0600.
				f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // user-chosen trace path
				if err != nil {
					return fmt.Errorf("open trace file %s: %w", path, err)
				}
				defer func() { _ = f.Close() }()
				model = tui.NewTracer(model, f)
			}

			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("run tui: %w", err)
			}
			return nil
		},
	}
}
