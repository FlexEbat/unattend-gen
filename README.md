# unattend-gen

CLI and TUI tool that generates `autounattend.xml` answer files for
unattended Windows 10/11 installs. It's a terminal counterpart to
[schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/):
answer language and edition, computer name and local accounts, telemetry and
system tweaks, Wi-Fi, once, and reuse the result for every install.

A profile is a plain JSON file that captures every setting. The CLI and the
TUI both read and write the same profile format and build the XML through
the same code path, so a given profile always produces the same answer file
no matter how it was filled in.

## Features

- Language, locale, keyboard layout, Windows edition and product key
- Computer name, time zone, and up to 5 local accounts, with auto-logon control
- Express settings (telemetry) and 23 system tweaks (Windows Update, UAC,
  Windows 11 hardware-check bypass, SmartScreen, Fast Startup, System Restore,
  long paths, Remote Desktop, junction-point cleanup, Windows Update reboot
  prevention, and more)
- OOBE screens (EULA, OEM registration, network setup) are hidden automatically
  once the profile has enough non-interactive settings to complete without
  them; an optional flag skips the Microsoft-account requirement (best-effort —
  Microsoft has patched around this more than once, so it isn't guaranteed on
  every Windows build)
- Wi-Fi profile (SSID, WPA2/WPA3/open, hidden networks)
- Remove preinstalled apps (Xbox, Teams, Solitaire, Cortana, and 28 more) and
  Windows features (Internet Explorer, WordPad, OpenSSH Client, and more)
- Custom scripts (.cmd/.ps1/.reg/.vbs) at four points: System (before
  accounts exist), DefaultUser (every account, including future ones),
  FirstLogon (once) and UserOnce (once per account)
- Password expiration and account lockout policy
- File Explorer tweaks (show hidden/system files, file extensions, classic
  right-click menu, tooltips, default folder, End task in taskbar) — applied
  to every account, including future ones
- Personalization colors (light/dark theme, accent color, transparency,
  solid-color wallpaper) — same every-account mechanism
- Two built-in presets (`minimal`, `single-user`) to start from
- Interactive TUI for filling in a profile screen by screen, with a live XML
  preview before saving
- Ships as a single static binary — no server, no network calls, no config
  beyond the profile JSON file

Disk partitioning is intentionally out of scope: Windows Setup always asks
where to install, same as a normal manual install.

## Tech stack

- [Go](https://go.dev/)
- [spf13/cobra](https://github.com/spf13/cobra) — CLI commands
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea),
  [bubbles](https://github.com/charmbracelet/bubbles),
  [lipgloss](https://github.com/charmbracelet/lipgloss) — TUI
- [go-playground/validator](https://github.com/go-playground/validator) —
  profile validation

## Install

Requires Go 1.23+.

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

Cross-compile for another OS with `GOOS`/`GOARCH`:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o unattend-gen.exe ./cmd/unattend-gen
```

## Usage

Create a profile, check it, and turn it into an answer file:

```sh
unattend-gen profile init demo                    # demo.json with default values
unattend-gen profile init demo --preset minimal    # or start from a preset
unattend-gen profile list                          # list profiles in ./profiles
unattend-gen validate demo.json                     # exit code 0 or 1
unattend-gen generate demo.json                     # writes autounattend.xml next to it
unattend-gen generate demo.json -o out.xml          # or to a chosen path
```

Or fill in a profile interactively:

```sh
unattend-gen tui               # start from scratch
unattend-gen tui demo.json     # start from an existing profile
```

Drop the resulting `autounattend.xml` at the root of a Windows installation
USB drive (or mount it as a virtual floppy/CD in a VM) and Windows Setup
picks it up automatically.

## Project structure

```text
cmd/unattend-gen/     entry point
internal/profile/     Profile schema, JSON load/save, validation
internal/xmlgen/       autounattend.xml builder and its components
internal/cli/          cobra commands: profile, validate, generate, tui
internal/tui/          bubbletea app: screens and shared widgets
presets/               built-in profile presets, embedded into the binary
```

## Development

```sh
make gate   # gofmt check, go vet, golangci-lint, go test -race
```

CI runs the same gate on every push, plus a cross-compile check for Linux,
macOS and Windows.

## Testing

```sh
go test ./... -race
```

## License

[GPL-3.0](LICENSE)
