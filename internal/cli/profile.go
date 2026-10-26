package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/presets"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage answer-file profiles",
	}
	cmd.AddCommand(newProfileInitCmd())
	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileShowCmd())
	return cmd
}

func newProfileInitCmd() *cobra.Command {
	var (
		preset string
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "init <n>",
		Short: "Create profiles/<n>.json with default values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			path, err := profile.ProfilePath(name)
			if err != nil {
				return err
			}
			if err := refuseOverwrite(path, force); err != nil {
				return err
			}
			p, err := profileForInit(name, preset)
			if err != nil {
				return err
			}
			if err := profile.SaveProfile(p, path); err != nil {
				return fmt.Errorf("create profile %s: %w", path, err)
			}
			cmd.Println(path)
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "", "start from a built-in preset: "+strings.Join(presets.Names, "|"))
	cmd.Flags().BoolVar(&force, "force", false, "replace the profile if it already exists")

	return cmd
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles in ./profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			paths, err := profile.ListProfiles(profile.ProfilesDir)
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

func newProfileShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <profile.json>",
		Short: "Validate a profile and print it as JSON with passwords masked",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := readValidProfile(cmd, args[0])
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(p.Redacted(), "", "  ")
			if err != nil {
				return fmt.Errorf("encode profile: %w", err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		},
	}
}

// profileForInit builds the profile for `profile init`: from a built-in
// preset when preset is non-empty, otherwise the built-in defaults. The
// profile's Name is always set to name, replacing whatever the preset file
// carries.
func profileForInit(name, preset string) (*profile.Profile, error) {
	if preset == "" {
		return profile.Default(name), nil
	}

	data, err := presets.Load(preset)
	if err != nil {
		return nil, err
	}
	p, err := profile.DecodeProfile(data)
	if err != nil {
		return nil, fmt.Errorf("decode preset %s: %w", preset, err)
	}
	p.Name = name
	return p, nil
}
