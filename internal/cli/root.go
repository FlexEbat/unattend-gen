// Package cli wires up the unattend-gen cobra commands.
package cli

import "github.com/spf13/cobra"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "unattend-gen",
		Short:         "Generate Windows autounattend.xml answer files",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newProfileCmd())
	root.AddCommand(newValidateCmd())
	return root
}

// Execute runs the unattend-gen CLI.
func Execute() error {
	return newRootCmd().Execute()
}
