package components

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	scriptsDir            = `C:\Windows\Setup\Scripts`
	defaultUserHivePath   = `C:\Users\Default\NTUSER.DAT`
	defaultUserHiveKey    = `HKU\DefaultUser`
	defaultUserRunOnceKey = defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\RunOnce`
)

func scriptExtension(f profile.ScriptFormat) string {
	return "." + string(f)
}

// invokeCommand returns the command that runs the script at path, quoted
// with double quotes throughout: that quoting works unescaped both as a
// PowerShell statement and, for UserOnce, inside a batch (.cmd) file — the
// two contexts this function's output ends up in.
func invokeCommand(format profile.ScriptFormat, path string) string {
	switch format {
	case profile.ScriptCmd:
		return fmt.Sprintf(`cmd.exe /c "%s"`, path)
	case profile.ScriptPs1:
		return fmt.Sprintf(`powershell.exe -WindowStyle Normal -ExecutionPolicy Unrestricted -NoProfile -File "%s"`, path)
	case profile.ScriptReg:
		return fmt.Sprintf(`reg.exe import "%s"`, path)
	case profile.ScriptVbs:
		return fmt.Sprintf(`cscript.exe //E:vbscript "%s"`, path)
	default:
		return ""
	}
}

// escapeForOuterCommand escapes double quotes for embedding a PowerShell
// statement inside our outer `powershell -NoProfile -Command "..."` string.
func escapeForOuterCommand(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func wrapCommand(statements []string) string {
	return `powershell -NoProfile -Command "` + escapeForOuterCommand(strings.Join(statements, ";")) + `"`
}

func writeFileStatement(path string, content []byte) string {
	encoded := base64.StdEncoding.EncodeToString(content)
	return fmt.Sprintf(`[IO.File]::WriteAllBytes("%s",[Convert]::FromBase64String('%s'))`, path, encoded)
}

func ensureScriptsDirStatement() string {
	return fmt.Sprintf(`New-Item -ItemType Directory -Force -Path "%s" | Out-Null`, scriptsDir)
}

// SystemScriptsCommands returns one command per script, each writing the
// script to disk and running it. They run in the system context before
// user accounts are created (Microsoft-Windows-Deployment, specialize).
func SystemScriptsCommands(scripts []profile.CustomScript) []string {
	var cmds []string
	for i, s := range scripts {
		path := fmt.Sprintf(`%s\unattend-sys-%02d%s`, scriptsDir, i+1, scriptExtension(s.Format))
		cmds = append(cmds, wrapCommand([]string{
			ensureScriptsDirStatement(),
			writeFileStatement(path, []byte(s.Content)),
			invokeCommand(s.Format, path),
		}))
	}
	return cmds
}

// FirstLogonScriptsCommands returns one command per script, each writing
// the script to disk and running it. They run once, when the first user
// logs on (Microsoft-Windows-Shell-Setup/FirstLogonCommands, oobeSystem).
func FirstLogonScriptsCommands(scripts []profile.CustomScript) []string {
	var cmds []string
	for i, s := range scripts {
		path := fmt.Sprintf(`%s\unattend-fl-%02d%s`, scriptsDir, i+1, scriptExtension(s.Format))
		cmds = append(cmds, wrapCommand([]string{
			ensureScriptsDirStatement(),
			writeFileStatement(path, []byte(s.Content)),
			invokeCommand(s.Format, path),
		}))
	}
	return cmds
}

// DefaultUserScriptCommand returns one command that mounts
// C:\Users\Default\NTUSER.DAT (the hive every new account is created from),
// writes and runs each script, then unmounts it. Empty when scripts is
// empty. Runs in the specialize pass, after SystemScripts and before
// accounts are created.
func DefaultUserScriptCommand(scripts []profile.CustomScript) string {
	if len(scripts) == 0 {
		return ""
	}
	statements := []string{ensureScriptsDirStatement(), fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath)}
	for i, s := range scripts {
		path := fmt.Sprintf(`%s\unattend-du-%02d%s`, scriptsDir, i+1, scriptExtension(s.Format))
		statements = append(statements, writeFileStatement(path, []byte(s.Content)), invokeCommand(s.Format, path))
	}
	statements = append(statements, fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey))
	return wrapCommand(statements)
}

// UserOnceScriptCommand returns one command that mounts the default user
// hive and, for each script, writes it plus a small .cmd wrapper and
// registers the wrapper in the hive's RunOnce key — so it runs once for
// every account created from that template, not just the current one.
// Empty when scripts is empty.
func UserOnceScriptCommand(scripts []profile.CustomScript) string {
	if len(scripts) == 0 {
		return ""
	}
	statements := []string{ensureScriptsDirStatement(), fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath)}
	for i, s := range scripts {
		n := i + 1
		scriptPath := fmt.Sprintf(`%s\unattend-uo-%02d%s`, scriptsDir, n, scriptExtension(s.Format))
		wrapperPath := fmt.Sprintf(`%s\unattend-uo-%02d-run.cmd`, scriptsDir, n)
		wrapperContent := "@echo off\r\n" + invokeCommand(s.Format, scriptPath) + "\r\n"
		valueName := fmt.Sprintf("UnattendUserOnce%02d", n)
		statements = append(statements,
			writeFileStatement(scriptPath, []byte(s.Content)),
			writeFileStatement(wrapperPath, []byte(wrapperContent)),
			fmt.Sprintf(`reg.exe add "%s" /v %s /d "%s" /f`, defaultUserRunOnceKey, valueName, wrapperPath),
		)
	}
	statements = append(statements, fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey))
	return wrapCommand(statements)
}

// RestartExplorerCommand restarts explorer.exe so Start menu/taskbar
// changes made by FirstLogon or UserOnce scripts take effect immediately.
func RestartExplorerCommand() string {
	return `powershell -NoProfile -Command "Stop-Process -Name explorer -Force -ErrorAction SilentlyContinue;Start-Process explorer.exe"`
}
