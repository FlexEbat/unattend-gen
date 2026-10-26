package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <profile.json>",
		Short: "Validate a profile file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := readValidProfile(cmd, args[0]); err != nil {
				return err
			}
			cmd.Println("профиль корректен")
			return nil
		},
	}
}

// readValidProfile reads and validates the profile at path. Errors are
// printed one per line to stderr and reported as errValidationFailed;
// warnings of a valid profile are printed to stderr too.
func readValidProfile(cmd *cobra.Command, path string) (*profile.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile %s: %w", path, err)
	}
	result := profile.ValidateProfile(data)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			cmd.PrintErrln(e)
		}
		return nil, errValidationFailed
	}
	for _, w := range result.Warnings {
		cmd.PrintErrln("Предупреждение: " + w)
	}
	return result.Profile, nil
}

var errValidationFailed = fmt.Errorf("profile validation failed")
