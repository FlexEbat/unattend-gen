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
			path := args[0]
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read profile %s: %w", path, err)
			}
			result := profile.ValidateProfile(data)
			if len(result.Errors) > 0 {
				for _, e := range result.Errors {
					cmd.PrintErrln(e)
				}
				return errValidationFailed
			}
			cmd.Println("профиль корректен")
			return nil
		},
	}
}

var errValidationFailed = fmt.Errorf("profile validation failed")
