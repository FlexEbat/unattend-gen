package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Print the JSON Schema of a profile file (for editor autocompletion)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data, err := profile.SchemaJSON()
			if err != nil {
				return fmt.Errorf("build schema: %w", err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		},
	}
}
