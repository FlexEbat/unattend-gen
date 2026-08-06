package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ProfilesDir is the conventional directory `profile list` (CLI) and the
// welcome screen (TUI) both read profiles from.
const ProfilesDir = "profiles"

// LoadProfile reads and decodes a Profile from a JSON file at path.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile %s: %w", path, err)
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decode profile %s: %w", path, err)
	}
	return &p, nil
}

// SaveProfile encodes profile as indented JSON and writes it to path.
func SaveProfile(profile *Profile, path string) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write profile %s: %w", path, err)
	}
	return nil
}

// ListProfiles returns the paths of *.json files in dir, sorted by name.
func ListProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read profiles dir %s: %w", dir, err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}
