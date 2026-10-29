// Package cli wires up the unattend-gen cobra commands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
)

// version is overridden at build time:
//
//	go build -ldflags "-X github.com/FlexEbat/unattend-gen/internal/cli.version=v0.29.0"
var version = "dev"

// NewRootCmd builds the unattend-gen command tree. Exported so tests can
// call it directly and redirect SetArgs/SetOut/SetErr without going through
// os.Args or the process's real stdout/stderr.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "unattend-gen",
		Short:         "Generate Windows autounattend.xml answer files",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		// --lang / UNATTEND_GEN_LANG select the language of messages. Without
		// either, the CLI keeps its default texts and `tui` asks.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			code, _ := cmd.Flags().GetString("lang")
			if code == "" {
				code = os.Getenv("UNATTEND_GEN_LANG")
			}
			if code == "" {
				return nil
			}
			l, ok := i18n.Parse(code)
			if !ok {
				return fmt.Errorf("unknown language %q (supported: en, ru, zh, es, hi)", code)
			}
			i18n.Set(l)
			return nil
		},
	}
	root.PersistentFlags().String("lang", "", "interface language: en, ru, zh, es, hi (env UNATTEND_GEN_LANG; the tui asks if unset)")
	root.AddCommand(newProfileCmd())
	root.AddCommand(newValidateCmd())
	root.AddCommand(newGenerateCmd())
	root.AddCommand(newSchemaCmd())
	root.AddCommand(newTUICmd())
	return root
}

// Execute runs the unattend-gen CLI.
func Execute() error {
	return NewRootCmd().Execute()
}
