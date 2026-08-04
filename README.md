# unattend-gen

CLI and TUI tool that builds `autounattend.xml` answer files for unattended
Windows 10/11 installs. Goal is a full-featured terminal counterpart to
[schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/).

A profile (JSON file) captures every setting once. The same profile drives
both the interactive TUI and the `generate` CLI command, and both go through
the same XML builder, so a profile always produces the same answer file
regardless of how it was filled in.

The tool ships as a single static binary for Linux, macOS and Windows. It
runs offline: no server, no network calls, no authentication. Profiles are
plain JSON files on disk.

## Status

Under active development, built slice by slice per `tech.md`. Current stage:
scaffold (module layout, `Profile` schema, validation, minimal XML skeleton,
`profile init`/`profile list`/`validate` commands).

## Build

```sh
go build -o unattend-gen ./cmd/unattend-gen
```

Cross-compile with `GOOS`/`GOARCH`, `CGO_ENABLED=0`:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o unattend-gen.exe ./cmd/unattend-gen
```

## Usage

```sh
unattend-gen profile init demo   # writes demo.json with default values
unattend-gen validate demo.json  # checks a profile, exit code 0 or 1
```

## Development

```sh
make gate   # gofmt check, go vet, golangci-lint, go test -race
```

See `tech.md` for the full contract: data model, package layout, CLI
commands, TUI screens, validation rules and the slice-by-slice roadmap.
