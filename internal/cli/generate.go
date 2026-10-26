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
	var (
		output string
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "generate <profile.json>",
		Short: "Validate a profile and write an autounattend.xml answer file",
		Long: "Validate a profile and write an autounattend.xml answer file.\n\n" +
			"The answer file can contain passwords, so it is written with permissions 0600 and\n" +
			"an existing file is only replaced with --force. Use -o - to print the XML to stdout instead.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profilePath := args[0]
			p, err := readValidProfile(cmd, profilePath)
			if err != nil {
				return err
			}

			xmlStr, err := xmlgen.BuildAnswerFile(p)
			if err != nil {
				return fmt.Errorf("build answer file: %w", err)
			}

			if output == "-" {
				_, err := fmt.Fprint(cmd.OutOrStdout(), xmlStr)
				return err
			}
			outputPath := output
			if outputPath == "" {
				outputPath = filepath.Join(filepath.Dir(profilePath), defaultAnswerFileName)
			}
			if err := refuseOverwrite(outputPath, force); err != nil {
				return err
			}
			if err := profile.WriteFileSecure(outputPath, []byte(xmlStr)); err != nil {
				return fmt.Errorf("write answer file %s: %w", outputPath, err)
			}

			cmd.Println(outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "path to write the answer file, or - for stdout (default: autounattend.xml next to the profile)")
	cmd.Flags().BoolVar(&force, "force", false, "replace the output file if it already exists")

	return cmd
}

// refuseOverwrite returns an error if path exists and force is false.
func refuseOverwrite(path string, force bool) error {
	if force {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists (use --force to replace it)", path)
	}
	return nil
}
