// Package cli wires up the unattend-gen cobra commands.
package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the unattend-gen command tree. Exported so tests can
// call it directly and redirect SetArgs/SetOut/SetErr without going through
// os.Args or the process's real stdout/stderr.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "unattend-gen",
		Short:         "Generate Windows autounattend.xml answer files",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newProfileCmd())
	root.AddCommand(newValidateCmd())
	root.AddCommand(newGenerateCmd())
	root.AddCommand(newTUICmd())
	return root
}

// Execute runs the unattend-gen CLI.
func Execute() error {
	return NewRootCmd().Execute()
}
