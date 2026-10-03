# unattend-gen — Usage Guide

English | [Русский](USAGE.ru.md) | [简体中文](USAGE.zh-CN.md) | [Español](USAGE.es.md) | [हिन्दी](USAGE.hi.md)

This is the full usage reference for `unattend-gen`: a Go CLI/TUI tool that
generates Windows 10/11 `autounattend.xml` answer files. You build a
**profile** (a JSON file describing how you want Windows installed and
configured) once, then generate an answer file from it any time you
reinstall — no web form, no network access, ships as a single static
binary.

For a short feature overview see [README.md](../README.md). This guide
covers every command, every TUI screen, and every field a profile can
contain.

If something here and the generated XML ever disagree, the code is
authoritative. This guide is written by hand from the same source of
truth the tool itself uses (`internal/profile/schema.go`), but mistakes are
possible. File an issue if you find one.

## Table of contents

- [What this tool does — and deliberately doesn't](#what-this-tool-does--and-deliberately-doesnt)
- [Installing / building](#installing--building)
- [Quick start](#quick-start)
- [CLI reference](#cli-reference)
- [The TUI, screen by screen](#the-tui-screen-by-screen)
- [Profile JSON reference](#profile-json-reference)
- [Built-in presets](#built-in-presets)
- [Example profiles](#example-profiles)
- [Verifying a generated answer file](#verifying-a-generated-answer-file)
- [Troubleshooting and known limitations](#troubleshooting-and-known-limitations)

## What this tool does — and deliberately doesn't

`unattend-gen` builds a single `autounattend.xml` file. You put it on a USB
stick (or the root of your install media) next to the Windows installer,
and Windows Setup runs it automatically, skipping every prompt you've
configured away.

**It does not partition or format your disk.** Windows Setup always asks
interactively where to install Windows — this is intentional, not a
missing feature. The tool covers everything *after* you've told Setup
which disk/partition to use: language, accounts, tweaks, app removal,
personalization, and so on.

**It does not require network access to run.** No web form, no server —
it's a local binary that reads a JSON profile and writes an XML file.

## Installing / building

You need Go 1.23+ to build from source (there are no pre-built binaries
published by this guide — check the repository's Releases page if one
exists for your platform).

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

This produces a single static binary, `unattend-gen` (or `unattend-gen.exe`
on Windows), with no runtime dependencies.

## Quick start

The fastest path is the interactive TUI:

```sh
./unattend-gen tui
```

Walk through the screens (arrow through them with **Ctrl+N**/**Esc**, jump
straight to the summary with **Ctrl+R**), then save. This writes a JSON
profile you can regenerate from later, or hand-edit.

If you'd rather work from JSON directly:

```sh
# Start a new profile with default (mostly interactive) settings
./unattend-gen profile init my-pc

# ...edit my-pc.json by hand, or continue in the TUI:
./unattend-gen tui my-pc.json

# Check it's valid
./unattend-gen validate my-pc.json

# Generate the answer file (writes autounattend.xml next to the profile
# by default)
./unattend-gen generate my-pc.json
```

Copy the resulting `autounattend.xml` to the root of your Windows install
USB (alongside `setup.exe`, or the root of the ISO if you're building
custom media) and boot from it — Windows Setup finds and applies it
automatically.

## CLI reference

Every command validates its input up front and fails loudly (non-zero
exit code, error text on stderr) rather than generating a partial or
guessed-at file.

### `profile init <name> [--preset minimal|single-user]`

Creates `<name>.json` in the current directory. Without `--preset`, the
profile starts from built-in defaults (language/locale/keyboard `en-US`,
everything else close to "ask me during setup" — see
[`profile.Default`](#defaults) below). With `--preset`, it starts from one
of the two built-in presets (see [Built-in presets](#built-in-presets)),
with `name` always set to whatever you passed on the command line
regardless of what the preset file says.

```sh
./unattend-gen profile init laptop --preset single-user
# -> laptop.json
```

### `profile list`

Lists every `*.json` file in `./profiles` (one path per line, nothing
else — this output is meant to be scriptable). The directory has to exist
and contain profiles; there's no other configuration for where profiles
live.

```sh
./unattend-gen profile list
```

### `validate <profile.json>`

Runs the same validation `generate` does, without building an answer
file. Prints one error per line to stderr and exits non-zero if anything's
wrong; prints `профиль корректен` (validation messages and this success
line are in Russian — see [Text conventions](#text-conventions) below) and
exits 0 if the profile is valid.

```sh
./unattend-gen validate laptop.json
```

### `generate <profile.json> [-o path]`

Validates the profile, then builds and writes the answer file. Without
`-o`/`--output`, the file is written as `autounattend.xml` in the same
directory as the profile. Prints the path it wrote on success.

```sh
./unattend-gen generate laptop.json
./unattend-gen generate laptop.json -o /media/usb/autounattend.xml
```

### `tui [profile.json]`

Opens the interactive terminal UI. With no argument, it starts from
built-in defaults. With a profile path, it loads that profile first (the
same loader `validate`/`generate` use) — so you can round-trip: edit in
the TUI, save, hand-tweak the JSON, reopen in the TUI, and so on.

```sh
./unattend-gen tui
./unattend-gen tui laptop.json
```

<a id="text-conventions"></a>
**A note on language**: code, comments, commit messages, and plain console
text (command help, success/failure messages you see when piping output)
are in English. Text the TUI shows you while you're filling in a profile,
and every validation error message, are in Russian. This is a project
convention, not a bug.

## The TUI, screen by screen

Screens appear in this order; **Ctrl+N** moves to the next one, **Esc**
goes back, **Ctrl+R** jumps straight to Review from anywhere, **Tab** /
**Shift+Tab** move between fields on a screen, and **Space** toggles
checkboxes. Every screen keeps the same in-memory profile in sync as you
move around, so nothing is lost by navigating back and forth.

1. **Welcome** — landing screen, nothing to configure.
2. **Language** — UI language / locale / keyboard layout (BCP-47 codes),
   Windows edition mode (interactive / generic key / custom key / key
   stored in BIOS-UEFI firmware), a separate activation-only product key,
   and target processor architecture.
3. **Accounts** — computer name (or leave blank for Windows to generate
   one), time zone, up to 5 local accounts (name/display name/password/
   group) in an editable table, first-logon behavior, and the
   best-effort "skip requiring a Microsoft account" checkbox.
4. **Tweaks** — express settings (telemetry/diagnostics) plus every one of
   the 33 system tweaks (see the [full list](#system-tweaks) below),
   password-expiration and account-lockout policy, and File Explorer
   tweaks — all on one screen since they're all "flip this default"
   settings.
5. **Wifi** — configure a Wi-Fi network to auto-connect on first boot,
   either by filling in SSID/security type/password/hidden, or by pasting
   in a raw exported WLAN profile XML instead.
6. **Apps** — three checkbox groups: apps to remove (Appx packages),
   Windows features to remove (DISM capabilities), and legacy optional
   features to remove (a third, separate removal mechanism) — see the
   [full lists](#remove-apps) below.
7. **Personalization** — light/dark theme (system and apps separately),
   accent color, where the accent color shows (Start/taskbar, title
   bars), transparency, and a solid desktop background color.
8. **Accessibility** — Sticky Keys (off/disabled/custom flag combination)
   and Caps/Num/Scroll Lock initial state + whether pressing them does
   anything.
9. **Desktop** — which desktop icons are shown (This PC, Recycle Bin,
   etc. — 13 total), and which special folders are pinned on the Start
   menu next to the power button (Windows 11).
10. **VisualEffects** — Windows' "Performance Options": a preset (best
    appearance / best performance) or 17 individual animation/appearance
    toggles.
11. **Taskbar** — Start menu and taskbar tweaks: disable widgets,
    left-align the taskbar (Windows 11), hide the Task View button,
    disable Bing results in search, always show every tray icon, the
    taskbar search box's display mode, Start pins (Windows 11 JSON) /
    tiles (Windows 10 XML), and pinned taskbar icons (empty or a custom
    layout XML).
12. **Advanced** — install VM guest tools (VirtualBox/VMware/VirtIO/
    Parallels), a raw AppLocker policy XML, a PowerShell script that
    computes a dynamic computer name, and three small checkboxes: keep
    the answer file after setup instead of deleting it, auto-start
    Narrator, and obscure account passwords in the generated XML.
13. **Scripts** — one custom script per category (System/DefaultUser/
    FirstLogon/UserOnce — see [Custom scripts](#custom-scripts) below)
    editable in a multi-line text area, plus a checkbox to restart
    Explorer after scripts run.
14. **Review** — a live-updating summary of the whole profile and,
    assuming it validates, the generated XML preview.

Not every field on every screen maps to a top-level JSON key one-to-one —
several screens edit nested objects (e.g. `system_tweaks`,
`personalization`). The [Profile JSON reference](#profile-json-reference)
below is organized by JSON structure, with a note on which screen edits
each piece, so you can find a field either way.

## Profile JSON reference

A profile is one JSON object, `schema_version: 1`. Every field below is
optional unless marked **required**; omitting an optional field (or
setting it to `null`/leaving a bool `false`) means "leave Windows' own
default behavior alone" — with one deliberate exception noted where it
comes up.

### Top level

| Field | Type | Notes |
|---|---|---|
| `schema_version` | int | **Required**, must be `1`. |
| `name` | string | **Required**. Free text — not written into the XML, just a label for the profile file itself. |

### Language & edition — *(Language screen)*

```json
"language": {
  "ui_language": "en-US",
  "locale": "en-US",
  "keyboard_layout": "en-US"
}
```

All three are **required** BCP-47 codes (e.g. `en-US`, `de-DE`, `ru-RU`).
`keyboard_layout` maps to the input locale; `locale` sets both the system
locale and the user locale.

```json
"edition": {
  "mode": "interactive",
  "edition": null,
  "product_key": null
}
```

- `mode`: one of
  - `"interactive"` — Windows Setup asks you which edition/key during
    install.
  - `"generic_key"` — requires `edition` to be one of `"Home"`, `"Pro"`,
    `"Education"`, `"Enterprise"`; Setup uses Microsoft's public generic
    (KMS client setup) key for that edition, still asks you to activate
    later.
  - `"custom_key"` — requires `product_key`, a real 25-character key in
    `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX` form. This key is also reused for
    activation later unless you set `activation_key` (below) explicitly.
  - `"firmware"` — uses whatever product key is already embedded in the
    device's BIOS/UEFI firmware (typical of OEM Windows preinstalls); no
    key is ever asked for or written.

```json
"activation_key": null,
"processor_architecture": ""
```

- `activation_key`: a separate key used **only** for activation
  (`Microsoft-Windows-Shell-Setup/ProductKey`), independent of whichever
  key (if any) `edition` uses to select what gets installed. Leave it
  `null` to fall back to the `edition.product_key` (only when
  `edition.mode` is `"custom_key"`) or to activate with nothing at all.
- `processor_architecture`: one of `"amd64"` (default when empty/absent),
  `"x86"`, `"arm64"`. Only a single architecture is supported per profile
  — unlike some other unattend generators, this tool does not build one
  answer file that installs on multiple architectures.

### Computer name & time zone — *(Accounts screen)*

```json
"computer_name": null,
"computer_name_script": null,
"timezone": null
```

- `computer_name`: a static hostname (1–15 chars, letters/digits/hyphens,
  can't start/end with a hyphen or be all digits). `null` lets Windows
  generate a random one.
- `computer_name_script`: a PowerShell script, run during setup, whose
  output becomes the computer name — for generating names dynamically
  (e.g. from serial number or a naming scheme). **Mutually exclusive**
  with `computer_name` — setting both is a validation error. Renaming
  happens via a background process that keeps re-applying the name for a
  short window after setup, working around Windows re-writing the name
  during its own specialize pass.
- `timezone`: a Windows time zone ID string, e.g. `"Russian Standard
  Time"`, `"Pacific Standard Time"`, `"UTC"`. `null` lets Windows
  auto-detect it. The tool does not validate that the string names a real
  time zone (the list is large and version-dependent) — only that it
  isn't an empty string if you set it at all.

### Accounts & first logon — *(Accounts screen)*

```json
"accounts": [
  {
    "name": "alice",
    "display_name": null,
    "password": "Sup3rSecret!",
    "group": "Administrators"
  }
],
"first_logon": { "mode": "first_created_account" }
```

- `accounts`: up to 5 entries.
  - `name` — **required**, ≤20 chars, no `" / \ [ ] : ; | = , + * ? < >`.
  - `display_name` — optional friendly name.
  - `password` — `null` means no password; an empty string is invalid
    (use `null` instead).
  - `group` — **required**, `"Administrators"` or `"Users"`.
- `first_logon.mode`: one of
  - `"none"` — no auto-logon; the first real boot shows the normal logon
    screen.
  - `"first_created_account"` — auto-logs into the first entry in
    `accounts` (which must be non-empty).
  - `"builtin_administrator"` — auto-logs into the hidden built-in
    Administrator account; requires
    `first_logon.builtin_administrator_password` to be set.

### Express settings & bypasses

```json
"express_settings": { "mode": "interactive" },
"bypass_online_account_requirement": false
```

- `express_settings.mode`: `"interactive"` (Setup asks about telemetry
  etc.), `"all_enabled"`, or `"all_disabled"`. Anything other than
  `"interactive"` also hides several OOBE screens automatically (EULA,
  OEM registration, network setup).
- `bypass_online_account_requirement`: best-effort attempt to let Setup
  finish with a local account instead of requiring a Microsoft account
  sign-in (writes the well-known `BypassNRO` registry value). **Not
  guaranteed** — Microsoft has patched around this more than once across
  2025–2026; don't rely on it for unattended fleets without testing on
  your current Windows build. Local accounts already skip this prompt
  reliably once you have at least one entry in `accounts`, regardless of
  this flag.

<a id="system-tweaks"></a>
### System tweaks — *(Tweaks screen)*

All 33 fields inside `"system_tweaks": { ... }`, every one a plain
boolean defaulting to `false` (no change) **except** `keep_sensitive_files`,
noted below:

| Field | What it does |
|---|---|
| `disable_windows_update` | Pauses/disables Windows Update. |
| `disable_uac` | Disables User Account Control prompts. |
| `bypass_win11_requirements` | Skips the TPM/Secure Boot/RAM hardware checks (windowsPE, before Setup evaluates them). |
| `disable_smart_app_control` | Disables Smart App Control. |
| `disable_smart_screen` | Disables SmartScreen (system + Edge). |
| `disable_fast_startup` | Disables Fast Startup (hybrid boot). |
| `disable_system_restore` | Disables System Restore. |
| `enable_long_paths` | Enables NTFS long-path support beyond 260 chars. |
| `enable_remote_desktop` | Enables Remote Desktop + firewall rule. |
| `allow_powershell_scripts` | Sets the PowerShell execution policy to `RemoteSigned`. |
| `disable_last_access_timestamp` | `fsutil behavior set disablelastaccess 1`. |
| `prevent_device_encryption` | Prevents automatic BitLocker device encryption. |
| `disable_auto_sign_on_last_user` | Disables automatic sign-on of the last interactive user after a restart. |
| `disable_wpbt` | Disables Windows Platform Binary Table execution. |
| `audit_process_creation` | Enables process-creation auditing, including command line. |
| `hide_edge_first_run` | Skips Edge's first-run experience. |
| `disable_edge_startup_boost` | Disables Edge's startup boost / background mode. |
| `delete_hidden_junctions` | Removes legacy NTFS junction points (e.g. `C:\Documents and Settings`) — applies to the setup account and every future account. |
| `prevent_automatic_reboot` | Stops Windows Update from rebooting a machine that's in active use (registers a scheduled task that keeps nudging "active hours" to the current time). |
| `turn_off_system_sounds` | Sets the sound scheme to "No Sounds" — for the setup account and every future account. |
| `disable_app_suggestions` | Disables Content Delivery Manager's silently-installed suggested apps. |
| `disable_pointer_precision` | Disables "Enhance pointer precision" (mouse acceleration). |
| `prevent_device_apps` | Prevents Windows from downloading/installing apps associated with specific hardware devices. |
| `harden_system_drive_acl` | Removes the "Authenticated Users" group's write access to `C:\`. |
| `make_edge_uninstallable` | Flips the internal policy flag that lets Edge show an "Uninstall" option. |
| `delete_windows_old` | Deletes `C:\Windows.old` (only relevant on in-place upgrades; a harmless no-op on a fresh install). |
| `disable_core_isolation` | Disables Memory Integrity / virtualization-based security (useful in some VM guests, or with older drivers). |
| `delete_edge_desktop_icon` | Removes the Microsoft Edge desktop shortcut — for the setup account and every future account. |
| `disable_widgets` | Disables the Widgets panel. |
| `left_taskbar` | Left-aligns the taskbar (Windows 11; default is centered). |
| `hide_task_view_button` | Hides the Task View button on the taskbar. |
| `disable_bing_results` | Disables Bing web results appearing in taskbar search. |
| `show_all_tray_icons` | Always shows every notification-area icon instead of collapsing inactive ones (mechanism differs between Windows 10 and 11 — handled automatically). |

`keep_sensitive_files` lives at the top level of the profile, not inside
`system_tweaks` — see the note right below, because its default behavior
is the one deliberate exception to "false/absent means unchanged" in this
whole schema.

```json
"keep_sensitive_files": false
```

By default (`false`, i.e. **absent from the JSON too**), the tool deletes
`C:\Windows\Panther\unattend.xml` / `unattend-original.xml` (the copies
Windows Setup keeps of your answer file, including any plaintext
passwords in it, unless you also set `obscure_passwords`) and its own
Wi-Fi profile temp file after setup finishes. Set this to `true` to leave
those files in place. This is the one field in the whole schema where the
default *does something* rather than *changing nothing* — the tradeoff
being "your passwords sit in plaintext on disk after setup" is worse than
breaking the usual convention.

### Password expiration & account lockout — *(Tweaks screen)*

```json
"password_expiration": { "mode": "default", "days": null },
"account_lockout": {
  "mode": "default",
  "threshold": null,
  "window_minutes": null,
  "duration_minutes": null
}
```

- `password_expiration.mode`: `"default"` (Windows' own default, 42
  days — no command emitted), `"never"` (passwords never expire), or
  `"custom"` (requires `days` ≥ 1).
- `account_lockout.mode`: `"default"` (Windows' own default — 10 failed
  attempts / 10 min window / 10 min duration), `"disabled"` (lockout
  turned off), or `"custom"` (requires all three numeric fields, each
  ≥ 1).

### File Explorer tweaks — *(Tweaks screen)*

```json
"file_explorer": {
  "hidden_files": "default",
  "show_file_extensions": false,
  "classic_context_menu": false,
  "hide_folder_tooltips": false,
  "open_to_this_pc": false,
  "show_end_task_in_taskbar": false
}
```

- `hidden_files`: `"default"`, `"show_hidden"` (show hidden files), or
  `"show_all"` (show hidden + protected operating system files).
- The remaining booleans: show file extensions; restore the classic
  (Windows 10-style) right-click menu on Windows 11; hide folder/desktop
  icon tooltips; open File Explorer to "This PC" instead of "Quick
  access"/"Home"; show "End task" directly in the taskbar's right-click
  menu.

All of these apply both to the account created during setup and to every
future account on the machine.

### Wi-Fi — *(Wifi screen)*

```json
"wifi": {
  "ssid": "MyNetwork",
  "authentication": "WPA2Personal",
  "password": "hunter2000",
  "connect_hidden": false,
  "raw_profile_xml": null
}
```

`wifi` as a whole is optional — omit it (or leave it `null`) to not
configure Wi-Fi at all.

- `authentication`: `"Open"`, `"WPA2Personal"`, or `"WPA3Personal"`.
  `password` is required (≥8 chars) unless `authentication` is `"Open"`.
- `raw_profile_xml`: if set, this raw WLAN profile XML (exported via
  `netsh wlan export profile key=clear`) is used verbatim instead of
  building one from `ssid`/`authentication`/`password`/`connect_hidden` —
  those four become optional in that case.

<a id="remove-apps"></a>
### Remove apps, features & optional features — *(Apps screen)*

Three separate lists, using three different underlying removal
mechanisms — kept separate because a name in one list isn't valid in
another.

```json
"remove_apps": ["OneDrive", "Terminal", "Store"],
"remove_features": ["InternetExplorer"],
"remove_optional_features": ["Recall"]
```

**`remove_apps`** (Appx packages, removed via
`Remove-AppxProvisionedPackage`) — any of:

`3DViewer`, `BingSearch`, `Calculator`, `Camera`, `Clipchamp`, `Clock`,
`Copilot`, `Cortana`, `DevHome`, `Family`, `FeedbackHub`, `GameAssist`,
`GetHelp`, `MailAndCalendar`, `Maps`, `MediaPlayerModern`, `MixedReality`,
`MoviesAndTV`, `News`, `Notepad`, `Office`, `OneDrive`, `OneNote`,
`Outlook`, `Paint`, `Paint3D`, `People`, `PhoneLink`, `PowerAutomate`,
`QuickAssist`, `Skype`, `SnippingTool`, `SolitaireCollection`,
`StickyNotes`, `Store`, `Teams`, `Terminal`, `Tips`, `ToDo`,
`VoiceRecorder`, `Wallet`, `Weather`, `XboxApps`.

`OneDrive` is handled differently under the hood from the rest (it isn't
distributed as an Appx package — the tool deletes its leftover shortcut
and setup executables and removes its autorun entry instead), but you use
it the same way, just by name in this same list.

**`remove_features`** (Windows optional *capabilities*, removed via
`Remove-WindowsCapability`) — any of: `InternetExplorer`, `WordPad`,
`PowerShellISE`, `OpenSSHClient`, `MediaPlayer`, `Speech`, `Handwriting`,
`WindowsHello`, `MathInputPanel`, `OneSync`, `StepsRecorder`.

**`remove_optional_features`** (legacy Windows optional features, removed
via `Disable-WindowsOptionalFeature` — a third, distinct mechanism) — any
of: `Recall`, `MediaFeatures`, `RemoteDesktopClient`.

### Personalization — *(Personalization screen)*

```json
"personalization": {
  "system_theme": "",
  "apps_theme": "",
  "accent_color": null,
  "show_accent_on_start_taskbar": false,
  "show_accent_on_title_bars": false,
  "disable_transparency": false,
  "solid_color_wallpaper": null
}
```

- `system_theme` / `apps_theme`: `"light"` or `"dark"`; empty/absent
  leaves Windows' default.
- `accent_color` / `solid_color_wallpaper`: 6-digit hex color strings
  (e.g. `"FF8800"`), no `#` prefix. `solid_color_wallpaper` replaces the
  desktop background with a flat color; an actual image-file wallpaper
  isn't supported.
- These apply to every future account, not just the one created during
  setup.

### Accessibility — *(Accessibility screen)*

```json
"sticky_keys": { "mode": "default", "flags": [] },
"lock_keys": null
```

- `sticky_keys.mode`: `"default"` (Windows' own default — on, 5×Shift
  activates it), `"disabled"` (turns off the 5×Shift shortcut itself,
  not just visual extras), or `"custom"` (use `flags`, any subset of
  `HotKeyActive`, `Indicator`, `TriState`, `TwoKeysOff`,
  `AudibleFeedback`, `HotKeySound`).
- `lock_keys`: `null` leaves Caps/Num/Scroll Lock behavior untouched.
  Set it to configure all three:

  ```json
  "lock_keys": {
    "caps_lock":   { "initial": "off", "behavior": "toggle" },
    "num_lock":    { "initial": "on",  "behavior": "toggle" },
    "scroll_lock": { "initial": "off", "behavior": "ignore" }
  }
  ```

  `initial` is `"off"` or `"on"` (state right after Windows starts);
  `behavior` is `"toggle"` (normal) or `"ignore"` (pressing the key does
  nothing at all — takes effect after a reboot).

### Desktop icons & Start folders — *(Desktop screen)*

```json
"desktop_icons": { "ThisPC": true, "RecycleBin": false },
"start_folders": ["Settings", "Documents", "Downloads"]
```

- `desktop_icons`: a map; keys you don't mention are left at Windows'
  default, keys you do mention are set explicitly `true` (shown) or
  `false` (hidden). Valid keys: `ThisPC`, `UserFiles`, `Network`,
  `RecycleBin`, `ControlPanel`, `Desktop`, `Documents`, `Downloads`,
  `Music`, `Pictures`, `Videos`, `Gallery`, `Home`.
- `start_folders`: an ordered list of special folders pinned next to the
  power button on the Windows 11 Start menu — the order in the list is
  the order they're pinned in. Valid values: `Settings`, `FileExplorer`,
  `Documents`, `Downloads`, `Music`, `Pictures`, `Videos`, `Network`,
  `PersonalFolder`. An empty/absent list means "leave Windows' own
  default set alone" — there's no way to express "pin exactly zero
  folders" with this field.

Both apply to every future account, not just the one created during
setup.

### Visual effects — *(VisualEffects screen)*

```json
"visual_effects": { "mode": "default", "custom": {} }
```

- `mode`: `"default"` (unchanged), `"best_appearance"` (all effects on),
  `"best_performance"` (all effects off), or `"custom"` (use `custom`).
- `custom`: a map of effect name → `true`/`false`; effects you don't
  mention keep Windows' default. Valid keys: `ControlAnimations`,
  `AnimateMinMax`, `TaskbarAnimations`, `DWMAeroPeekEnabled`,
  `MenuAnimation`, `TooltipAnimation`, `SelectionFade`,
  `DWMSaveThumbnailEnabled`, `CursorShadow`, `ListviewShadow`,
  `ThumbnailsOrIcon`, `ListviewAlphaSelect`, `DragFullWindows`,
  `ComboBoxAnimation`, `FontSmoothing`, `ListBoxSmoothScrolling`,
  `DropShadow`.

Applies to every future account, not just the one created during setup.

### Start menu & taskbar — *(Taskbar screen)*

```json
"taskbar_search": "",
"start_pins": { "mode": "default", "json": null },
"start_tiles": { "mode": "default", "xml": null },
"taskbar_icons": { "mode": "default", "xml": null }
```

- `taskbar_search`: `""`/absent (Windows default, search box shown),
  `"hide"`, `"icon"` (icon only, no box), `"box"` (explicit, same as
  default), or `"label"` (icon with text label).
- `start_pins` (Windows 11 only, no effect on Windows 10): `mode` is
  `"default"`, `"empty"` (no apps pinned), or `"custom"` (requires
  `json` — a raw `{"pinnedList": [...]}` payload in the format Windows'
  `ConfigureStartPins` policy expects).
- `start_tiles` (Windows 10 only, no effect on Windows 11): `mode` is
  `"default"`, `"empty"` (no tile groups), or `"custom"` (requires
  `xml` — a raw `LayoutModification.xml` document).
- `taskbar_icons`: icons pinned to the taskbar for every future account.
  `mode` is `"default"`, `"empty"` (no pinned icons) or `"custom"` (requires
  `xml` — a raw `LayoutModification.xml` containing a
  `CustomTaskbarLayoutCollection`). It works through a locked Start layout
  that is unlocked again at each account's first logon, so users can
  rearrange the taskbar afterwards; the generated answer file therefore also
  registers a small `UnlockStartLayout` scheduled task. Only well-formedness
  of the XML is checked, not Windows' taskbar-layout schema.

### Advanced — *(Advanced screen)*

```json
"install_vm_guest_tools": [],
"applocker_policy_xml": null,
"use_narrator": false,
"obscure_passwords": false
```

- `install_vm_guest_tools`: any of `VBoxGuestAdditions`, `VMwareTools`,
  `VirtIoGuestTools`, `ParallelsTools`. Each runs a silent installer that
  looks for its guest-tools ISO on drive letters D–Z and no-ops with a
  log message if it isn't attached — safe to list all four "just in
  case" if you don't know in advance which hypervisor you're on.
- `applocker_policy_xml`: a raw AppLocker policy XML. Only checked for
  being well-formed XML, not validated against AppLocker's full schema —
  a well-formed-but-invalid policy will only surface as an error on the
  target machine.
- `use_narrator`: auto-starts Narrator during the install itself and at
  every future account's first logon.
- `obscure_passwords`: Base64-obscures account passwords in the generated
  XML instead of writing them in plain text. This is **obfuscation, not
  encryption** — the salt is a fixed, publicly documented value
  (Microsoft's own unattend convention), so it only keeps passwords from
  being trivially grep-able in the raw file, nothing more.

### Global setup switch

```json
"hide_powershell_windows": false
```

When `true`, every PowerShell window this tool spawns during setup
(built-in tweaks and your own custom scripts alike) runs hidden instead
of visible. Doesn't affect `cmd`/`reg`/`vbs`-based steps, which don't show
a window either way.

<a id="custom-scripts"></a>
### Custom scripts — *(Scripts screen)*

```json
"system_scripts": [],
"default_user_scripts": [],
"first_logon_scripts": [],
"user_once_scripts": [],
"restart_explorer_after_scripts": false
```

Four categories, each a list of `{ "format": "...", "content": "..." }`
objects. `format` is one of `cmd`, `ps1`, `reg`, `vbs` (`default_user_scripts`
doesn't support `vbs`). Limits: `system_scripts` ≤4, `default_user_scripts`
≤3, `first_logon_scripts` ≤4, `user_once_scripts` ≤4.

- **`system_scripts`** — run once, in the system context, *before* any
  user account exists.
- **`default_user_scripts`** — applied to the default-user registry
  template, so they affect every account created afterward, including
  ones that don't exist yet — not the account created during setup
  itself (unless it's also created against this template, which it
  normally is).
- **`first_logon_scripts`** — run once, when the very first account logs
  on.
- **`user_once_scripts`** — run once *per account*, including future
  ones, the first time each one logs on.

`restart_explorer_after_scripts`: if any `first_logon_scripts` ran,
restart Explorer afterward (useful if a script changed something Explorer
caches).

<a id="defaults"></a>
### What `profile init` (no preset) gives you

`schema_version: 1`, language/locale/keyboard all `"en-US"`,
`edition.mode: "interactive"`, no accounts, `first_logon.mode: "none"`,
`express_settings.mode: "interactive"` — everything else is the JSON zero
value (empty string/false/null/empty list), i.e. "ask me during setup" or
"leave Windows' default" for every setting this guide describes.

## Built-in presets

Two presets ship embedded in the binary; use them with
`profile init <name> --preset <preset-name>`.

- **`minimal`** — same as no preset at all, essentially: interactive
  edition, no accounts, interactive express settings, explicit `false`
  for the three original system tweaks. A safe, do-nothing-extra
  starting point to build up from.
- **`single-user`** — one local Administrator account named `admin` (no
  password set — add one before using this for real), auto-logs into
  that account (`first_logon.mode: "first_created_account"`), and
  `express_settings.mode: "all_disabled"` (skip the telemetry/diagnostics
  prompts entirely). A reasonable starting point for a personal machine
  you're the only user of.

Both presets predate several of the fields documented above (they were
written against an earlier, smaller schema) — anything they don't
mention just takes the schema's zero-value default, same as an unset
field anywhere else.

## Example profiles

A fuller profile, combining several of the sections above — a
single-user laptop with some privacy/performance tweaks and a couple of
apps removed:

```json
{
  "schema_version": 1,
  "name": "laptop",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "interactive" },
  "computer_name": "LAPTOP-01",
  "timezone": "UTC",
  "accounts": [
    {
      "name": "alice",
      "display_name": "Alice",
      "password": "Sup3rSecret!",
      "group": "Administrators"
    }
  ],
  "first_logon": { "mode": "first_created_account" },
  "express_settings": { "mode": "all_disabled" },
  "bypass_online_account_requirement": true,
  "system_tweaks": {
    "disable_windows_update": false,
    "disable_uac": false,
    "bypass_win11_requirements": true,
    "disable_smart_screen": false,
    "enable_remote_desktop": true,
    "delete_hidden_junctions": true,
    "prevent_automatic_reboot": true,
    "left_taskbar": true,
    "hide_task_view_button": true
  },
  "file_explorer": {
    "hidden_files": "show_all",
    "show_file_extensions": true,
    "open_to_this_pc": true
  },
  "remove_apps": ["Cortana", "Skype", "Teams", "SolitaireCollection"],
  "remove_features": ["InternetExplorer"],
  "personalization": {
    "system_theme": "dark",
    "apps_theme": "dark",
    "accent_color": "0078D4"
  },
  "keep_sensitive_files": false,
  "obscure_passwords": true
}
```

A no-frills unattended minimal install (no accounts of your own, just
skip the hardware checks and update pause) using a generic key:

```json
{
  "schema_version": 1,
  "name": "minimal-vm",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "generic_key", "edition": "Pro" },
  "accounts": [],
  "first_logon": { "mode": "none" },
  "express_settings": { "mode": "interactive" },
  "system_tweaks": {
    "bypass_win11_requirements": true,
    "disable_windows_update": true
  },
  "install_vm_guest_tools": ["VBoxGuestAdditions", "VMwareTools"]
}
```

## Verifying a generated answer file

`validate`/`generate` only check the JSON profile against the schema —
they can't confirm Windows Setup will actually accept the resulting XML
end to end, because that requires a real Windows Setup run. There's no
substitute for actually burning the result to a USB stick (or mounting it
in a VM) and watching it install. If something in the XML is wrong,
Windows Setup will usually tell you which element via its own error
dialog or `setupact.log`/`setuperr.log` under
`C:\Windows\Panther` (or `X:\Windows\Panther` while still in the PE
environment).

## Troubleshooting and known limitations

- **Disk partitioning is out of scope, on purpose.** Windows Setup will
  always stop and ask you which disk/partition to install to. This tool
  configures everything else around that interactive step.
- **`bypass_online_account_requirement` is best-effort.** Microsoft has
  changed how this works more than once; if it stops working on a future
  Windows build, that's a Microsoft-side change, not a bug report against
  this tool (though an issue about it is still welcome).
- **Multi-architecture answer files aren't supported.** Pick one
  `processor_architecture` per profile.
- **Only one language and one keyboard layout** can be configured per
  profile; additional languages and layouts aren't supported.
- **A raw XML "escape hatch" for arbitrary unattend components** isn't
  implemented — if you need a component this tool doesn't expose a field
  for, you'll need to hand-edit the generated `autounattend.xml` after
  the fact.
- **Image-file desktop wallpaper** isn't supported — only a flat solid
  color (`personalization.solid_color_wallpaper`).
- **Lock-screen image** isn't supported — Microsoft's mechanism for it is
  gated to Enterprise/Education/Pro-SharedPC editions, which makes it
  unreliable for this tool's mostly Home/Pro audience, so it was left out
  rather than shipped half-working.
