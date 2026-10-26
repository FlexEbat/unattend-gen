# Changelog

Earlier history is in the git tags (`v0.x.y`). From here on, notable changes
are listed per release.

## Unreleased

### Fixed

- Commands longer than Windows Setup's documented 259-character limit
  (`RunSynchronousCommand/Path`, `FirstLogonCommands/CommandLine`). Chains of
  `a && b` are split into one command per part; PowerShell commands are moved
  into files carried by an `<Extensions>` element and extracted by one short
  command that runs first in the specialize pass. Not yet verified on a real
  Windows installation - please test before relying on it.
- Unknown JSON keys in a profile were silently ignored (`"acounts"` produced a
  profile without accounts). They are now rejected, as is trailing data.
- Profiles and answer files, which can contain passwords, were created
  world-readable (`0644`). They are now written atomically with `0600`.
- `profile init` wrote `./<name>.json` while `profile list` and the TUI read
  `./profiles`. `init` and the TUI save command now use `profiles/<name>.json`;
  profile names containing path separators are rejected.
- Account-name length was checked in bytes by one rule and characters by
  another, and over-long or too many accounts were reported twice.

### Added

- Validation: duplicate account names (case-insensitive), names reserved by
  Windows, names ending in a period or with edge spaces, duplicate list
  entries, whitespace-only scripts, WPA2 passphrases over 63 characters and
  product keys outside the key alphabet.
- Validation errors from struct rules now name the JSON path
  (`accounts[1].group: ...`).
- Warnings for valid but risky settings (plaintext passwords, `disable_uac`,
  `disable_windows_update`, ...), printed to stderr.
- CLI: `--version`, `schema` (JSON Schema of the profile), `profile show`
  (passwords masked), `--force` for `generate` and `profile init`,
  `generate -o -` (stdout).
- Tests: golden answer files, a fuzz target over validate-and-build, and
  checks that every generated command fits the length limit.
- `SECURITY.md`, a release workflow, Dependabot, and `govulncheck` in CI.

### Changed

- `NewDeployment` and `NewShellSetupOOBE` take a `*profile.Profile` instead of
  17-24 positional parameters.
- The sensitive-files cleanup also deletes the generated `unattend-cmd-*`
  helper files.
- CI: read-only token, pinned action SHAs, pinned golangci-lint v2.15.0 with
  `errorlint`, `gosec` and `misspell` enabled; the `go mod tidy` auto-commit
  was removed in favour of `go mod verify` and `go mod tidy -diff`. Release
  binaries are built for arm64 as well.
- The generic validation loops are one generic helper.
