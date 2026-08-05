// Package presets embeds the built-in profile presets into the unattend-gen
// binary, so `profile init --preset NAME` works without the presets/
// directory being present at runtime.
package presets

import (
	"embed"
	"fmt"
)

//go:embed minimal.json single-user.json
var files embed.FS

// Names lists the valid preset names, in the order shown to the user.
var Names = []string{"minimal", "single-user"}

// Load returns the raw JSON content of the named preset.
func Load(name string) ([]byte, error) {
	data, err := files.ReadFile(name + ".json")
	if err != nil {
		return nil, fmt.Errorf("unknown preset %q: %w", name, err)
	}
	return data, nil
}
