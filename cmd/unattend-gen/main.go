// Command unattend-gen generates Windows autounattend.xml answer files.
package main

import (
	"fmt"
	"os"

	"github.com/FlexEbat/unattend-gen/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
