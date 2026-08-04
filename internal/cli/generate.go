package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen"
)

const defaultAnswerFileName = "autounattend.xml"

func newGenerateCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "generate <profile.json>",
		Short: "Validate a profile and write an autounattend.xml answer file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profilePath := args[0]
			data, err := os.ReadFile(profilePath)
			if err != nil {
				return fmt.Errorf("read profile %s: %w", profilePath, err)
			}

			result := profile.ValidateProfile(data)
			if len(result.Errors) > 0 {
				for _, e := range result.Errors {
					cmd.PrintErrln(e)
				}
				return errValidationFailed
			}

			xmlStr, err := xmlgen.BuildAnswerFile(result.Profile)
			if err != nil {
				return fmt.Errorf("build answer file: %w", err)
			}

			outputPath := output
			if outputPath == "" {
				outputPath = filepath.Join(filepath.Dir(profilePath), defaultAnswerFileName)
			}
			if err := os.WriteFile(outputPath, []byte(xmlStr), 0o644); err != nil {
				return fmt.Errorf("write answer file %s: %w", outputPath, err)
			}

			cmd.Println(outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "path to write the answer file (default: autounattend.xml next to the profile)")

	return cmd
}
