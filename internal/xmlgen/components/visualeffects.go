package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 25 (tech.md backlog, Visual effects — outside the original 6 audit
// groups). Mechanism sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs),
// not invented from memory.

// visualEffectsPresetValues returns the per-effect true/false map for a
// preset mode: all effects on (best appearance), all off (best
// performance), or the user's own Custom map. Returns nil for
// VisualEffectsModeDefault (nothing to write).
func visualEffectsPresetValues(s profile.VisualEffectsSettings) map[profile.VisualEffect]bool {
	switch s.Mode {
	case profile.VisualEffectsModeBestAppearance:
		values := make(map[profile.VisualEffect]bool, len(profile.VisualEffects))
		for _, e := range profile.VisualEffects {
			values[e] = true
		}
		return values
	case profile.VisualEffectsModeBestPerformance:
		values := make(map[profile.VisualEffect]bool, len(profile.VisualEffects))
		for _, e := range profile.VisualEffects {
			values[e] = false
		}
		return values
	case profile.VisualEffectsModeCustom:
		return s.Custom
	default:
		return nil
	}
}

// visualFXSettingValue is the HKCU VisualFXSetting value that matches each
// mode: 1 = best appearance, 2 = best performance, 3 = custom (the same
// values the Performance Options dialog itself writes).
func visualFXSettingValue(mode profile.VisualEffectsMode) int {
	switch mode {
	case profile.VisualEffectsModeBestAppearance:
		return 1
	case profile.VisualEffectsModeBestPerformance:
		return 2
	default:
		return 3
	}
}

// VisualEffectsSpecializeCommand returns one specialize-pass command that
// seeds HKLM's per-effect DefaultValue template — the values Windows
// copies into a NEW account's own settings the first time that account's
// Explorer profile is created — so this applies to every future account,
// not just the one created during setup. Returns "" for
// VisualEffectsModeDefault.
func VisualEffectsSpecializeCommand(s profile.VisualEffectsSettings) string {
	values := visualEffectsPresetValues(s)
	if values == nil {
		return ""
	}
	var statements []string
	for _, e := range profile.VisualEffects {
		on, ok := values[e]
		if !ok {
			continue
		}
		v := 0
		if on {
			v = 1
		}
		key := `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects\` + string(e)
		statements = append(statements, fmt.Sprintf(`reg.exe add "%s" /v DefaultValue /t REG_DWORD /d %d /f`, key, v))
	}
	if len(statements) == 0 {
		return ""
	}
	return wrapCommand(statements)
}

// VisualEffectsUserOnceCommand returns one specialize-pass command that
// mounts the default user hive just long enough to register a RunOnce
// entry; that entry sets the live account's own VisualFXSetting (which
// preset the Performance Options dialog shows as selected) at first logon
// — the same RunOnce-to-live-HKCU mechanism DesktopIconsUserOnceCommand
// uses. Returns "" for VisualEffectsModeDefault.
func VisualEffectsUserOnceCommand(s profile.VisualEffectsSettings, hidePowerShellWindows bool) string {
	if s.Mode == profile.VisualEffectsModeDefault || s.Mode == "" {
		return ""
	}
	script := fmt.Sprintf(`Set-ItemProperty -LiteralPath 'Registry::HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects' -Name 'VisualFXSetting' -Type 'DWord' -Value %d -Force;`, visualFXSettingValue(s.Mode))
	scriptPath := scriptsDir + `\unattend-visual-effects-uo.ps1`
	wrapperPath := scriptsDir + `\unattend-visual-effects-uo-run.cmd`
	wrapperContent := "@echo off\r\n" + invokeCommand(profile.ScriptPs1, scriptPath, hidePowerShellWindows) + "\r\n"
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(scriptPath, []byte(script)),
		writeFileStatement(wrapperPath, []byte(wrapperContent)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendVisualEffects /d "%s" /f`, defaultUserRunOnceKey, wrapperPath),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}
