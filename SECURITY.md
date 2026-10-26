# Security

## What the generated files contain

Profiles (`profiles/*.json`) and the generated `autounattend.xml` can contain
account, administrator and Wi-Fi passwords.

- Both are written with permissions `0600`, and `generate` never replaces an
  existing file without `--force`.
- `.gitignore` excludes `profiles/*.json`, root-level `*.json` and
  `autounattend.xml`. Do not commit real profiles or answer files.
- `obscure_passwords` only Base64-encodes passwords inside the XML. It is not
  encryption: anyone holding the file can read them. `validate`, `generate` and
  `profile show` warn about this.
- Unless `keep_sensitive_files` is set, a first-logon command deletes the copy
  of the answer file that Windows Setup keeps in `C:\Windows\Panther`, the
  temporary Wi-Fi profile and the generated `unattend-cmd-*` helper files.
  The file on your installation media is not touched - remove it yourself.

## Reporting a vulnerability

Please report security problems privately through GitHub's "Report a
vulnerability" button on the repository's Security tab rather than a public
issue. Expect an acknowledgement within a week.
