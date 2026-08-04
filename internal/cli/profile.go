package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const profilesDir = "profiles"

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage answer-file profiles",
	}
	cmd.AddCommand(newProfileInitCmd())
	cmd.AddCommand(newProfileListCmd())
	return cmd
}

func newProfileInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new profile with default values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			p := defaultProfile(name)
			path := name + ".json"
			if err := profile.SaveProfile(p, path); err != nil {
				return fmt.Errorf("create profile %s: %w", path, err)
			}
			cmd.Println(path)
			return nil
		},
	}
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles in ./profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			paths, err := profile.ListProfiles(profilesDir)
			if err != nil {
				return fmt.Errorf("list profiles: %w", err)
			}
			for _, p := range paths {
				cmd.Println(p)
			}
			return nil
		},
	}
}

// defaultProfile returns a profile that passes ValidateProfile with no
// accounts and interactive express settings, per tech.md section 7.
func defaultProfile(name string) *profile.Profile {
	return &profile.Profile{
		SchemaVersion: 1,
		Name:          name,
		Language: profile.LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition: profile.EditionSettings{
			Mode: profile.EditionModeInteractive,
		},
		Accounts: []profile.UserAccount{},
		FirstLogon: profile.FirstLogon{
			Mode: profile.FirstLogonNone,
		},
		ExpressSettings: profile.ExpressSettings{
			Mode: profile.ExpressInteractive,
		},
	}
}
