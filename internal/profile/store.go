package profile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProfilesDir is the conventional directory `profile init` (CLI), `profile
// list` (CLI), the TUI save command and the welcome screen all use.
const ProfilesDir = "profiles"

// DecodeProfile decodes a Profile from JSON, rejecting unknown fields (a typo
// such as "acounts" must not silently produce a profile without accounts) and
// trailing data after the JSON object. It does not validate the result.
func DecodeProfile(data []byte) (*Profile, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Profile
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data after the JSON object")
	}
	return &p, nil
}

// LoadProfile reads and decodes a Profile from a JSON file at path.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile %s: %w", path, err)
	}
	p, err := DecodeProfile(data)
	if err != nil {
		return nil, fmt.Errorf("decode profile %s: %w", path, err)
	}
	return p, nil
}

// ProfilePath returns the path profiles/<name>.json for a profile name, or an
// error if the name could escape the profiles directory.
func ProfilePath(name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid profile name %q: it must not contain path separators", name)
	}
	return filepath.Join(ProfilesDir, name+".json"), nil
}

// WriteFileSecure writes data to path atomically (temporary file in the same
// directory, then rename) with permissions 0600: profiles and answer files
// can contain passwords. Missing parent directories are created (0700).
func WriteFileSecure(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".unattend-gen-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil && !errors.Is(err, os.ErrPermission) {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

// SaveProfile encodes profile as indented JSON and writes it to path
// (permissions 0600, replacing any existing file atomically).
func SaveProfile(profile *Profile, path string) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	if err := WriteFileSecure(path, data); err != nil {
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
