package components

import (
	"fmt"
	"strconv"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	personalizeKey = defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	dwmKey         = defaultUserHiveKey + `\Software\Microsoft\Windows\DWM`
	colorsKey      = defaultUserHiveKey + `\Control Panel\Colors`
	desktopKey     = defaultUserHiveKey + `\Control Panel\Desktop`
)

// PersonalizationCommand returns one command that mounts the default user
// hive, applies s, and unmounts it — the same DefaultUser mechanism as
// FileExplorerCommand, so every account (including future ones) gets these
// colors. Returns "" when s is the zero value: nothing to change.
func PersonalizationCommand(s profile.PersonalizationSettings) string {
	var ops []string

	if s.SystemTheme != "" {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v SystemUsesLightTheme /t REG_DWORD /d %d /f`,
			personalizeKey, lightThemeValue(s.SystemTheme)))
	}
	if s.AppsTheme != "" {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v AppsUseLightTheme /t REG_DWORD /d %d /f`,
			personalizeKey, lightThemeValue(s.AppsTheme)))
	}
	if s.ShowAccentOnStartTaskbar {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v ColorPrevalence /t REG_DWORD /d 1 /f`, personalizeKey))
	}
	if s.ShowAccentOnTitleBars {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v ColorPrevalence /t REG_DWORD /d 1 /f`, dwmKey))
	}
	if s.DisableTransparency {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v EnableTransparency /t REG_DWORD /d 0 /f`, personalizeKey))
	}
	if s.AccentColor != nil {
		packed := accentColorToPackedABGR(*s.AccentColor)
		ops = append(ops,
			fmt.Sprintf(`reg.exe add "%s" /v AccentColor /t REG_DWORD /d 0x%s /f`, dwmKey, packed),
			fmt.Sprintf(`reg.exe add "%s" /v ColorizationColor /t REG_DWORD /d 0x%s /f`, dwmKey, packed),
		)
	}
	if s.SolidColorWallpaper != nil {
		ops = append(ops,
			fmt.Sprintf(`reg.exe add "%s" /v Background /t REG_SZ /d "%s" /f`, colorsKey, hexColorToDecimalRGB(*s.SolidColorWallpaper)),
			fmt.Sprintf(`reg.exe add "%s" /v Wallpaper /t REG_SZ /d "" /f`, desktopKey),
			fmt.Sprintf(`reg.exe add "%s" /v WallpaperStyle /t REG_SZ /d 0 /f`, desktopKey),
			fmt.Sprintf(`reg.exe add "%s" /v TileWallpaper /t REG_SZ /d 0 /f`, desktopKey),
		)
	}

	if len(ops) == 0 {
		return ""
	}

	statements := append([]string{fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath)}, ops...)
	statements = append(statements, fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey))
	return wrapCommand(statements)
}

func lightThemeValue(t profile.ColorTheme) int {
	if t == profile.ColorThemeDark {
		return 0
	}
	return 1
}

// accentColorToPackedABGR converts a "RRGGBB" hex string to the packed
// "AABBGGRR" DWORD (alpha FF) DWM's AccentColor/ColorizationColor expect.
func accentColorToPackedABGR(rrggbb string) string {
	r, g, b := rrggbb[0:2], rrggbb[2:4], rrggbb[4:6]
	return "FF" + b + g + r
}

// hexColorToDecimalRGB converts a "RRGGBB" hex string to the space-separated
// decimal "R G B" format the classic Control Panel\Colors\Background value
// expects.
func hexColorToDecimalRGB(rrggbb string) string {
	var r, g, b int64
	r, _ = strconv.ParseInt(rrggbb[0:2], 16, 0)
	g, _ = strconv.ParseInt(rrggbb[2:4], 16, 0)
	b, _ = strconv.ParseInt(rrggbb[4:6], 16, 0)
	return fmt.Sprintf("%d %d %d", r, g, b)
}
