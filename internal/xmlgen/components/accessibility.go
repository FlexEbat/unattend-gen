package components

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 19 (tech.md backlog group C): Sticky Keys and Lock Key settings —
// two accessibility-related sections entirely missing before this slice.
// Mechanisms sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs),
// not invented from memory.

var stickyKeysFlagBits = map[profile.StickyKeysFlag]int{
	profile.StickyKeysHotKeyActive:    0x00000004,
	profile.StickyKeysIndicator:       0x00000020,
	profile.StickyKeysTriState:        0x00000080,
	profile.StickyKeysTwoKeysOff:      0x00000100,
	profile.StickyKeysAudibleFeedback: 0x00000040,
	profile.StickyKeysHotKeySound:     0x00000010,
}

// stickyKeysFlagsValue computes the Control Panel\Accessibility\StickyKeys
// "Flags" value: SKF_AVAILABLE (0x2) | SKF_CONFIRMHOTKEY (0x8), OR'd with
// whichever flags are selected. StickyKeysModeDisabled passes no flags,
// which — notably — clears SKF_HOTKEYACTIVE (0x4) along with everything
// else, disabling the 5x-Shift activation shortcut itself.
func stickyKeysFlagsValue(flags []profile.StickyKeysFlag) int {
	result := 0x00000002 | 0x00000008
	for _, f := range flags {
		result |= stickyKeysFlagBits[f]
	}
	return result
}

// StickyKeysDefaultUserCommand returns one specialize-pass command that
// mounts the default user hive and writes the Sticky Keys Flags value
// there, so every future account gets the same setting. Returns "" for
// StickyKeysModeDefault (Windows' own default applies, nothing to write).
func StickyKeysDefaultUserCommand(s profile.StickyKeysSettings) string {
	if s.Mode != profile.StickyKeysModeDisabled && s.Mode != profile.StickyKeysModeCustom {
		return ""
	}
	value := stickyKeysFlagsValue(s.Flags)
	key := defaultUserHiveKey + `\Control Panel\Accessibility\StickyKeys`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v Flags /t REG_SZ /d %d /f`, key, value),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// StickyKeysSystemDefaultCommand returns one specialize-pass command that
// writes the same Flags value to HKU\.DEFAULT — the hive Windows uses for
// the logon screen and any session before a user hive is loaded. Unlike
// DefaultUser, .DEFAULT is always mounted, so this needs no load/unload
// cycle. Returns "" for StickyKeysModeDefault.
func StickyKeysSystemDefaultCommand(s profile.StickyKeysSettings) string {
	if s.Mode != profile.StickyKeysModeDisabled && s.Mode != profile.StickyKeysModeCustom {
		return ""
	}
	value := stickyKeysFlagsValue(s.Flags)
	return fmt.Sprintf(`cmd.exe /c reg add "HKU\.DEFAULT\Control Panel\Accessibility\StickyKeys" /v Flags /t REG_SZ /d %d /f`, value)
}

// lockKeyScancodes are the make-scancodes Windows assigns to Caps Lock,
// Num Lock and Scroll Lock — the same values the "Scancode Map" keyboard
// remap mechanism (HKLM\SYSTEM\CurrentControlSet\Control\Keyboard Layout)
// uses to identify a physical key.
const (
	scancodeCapsLock   = 0x3A
	scancodeNumLock    = 0x45
	scancodeScrollLock = 0x46
)

// LockKeyIndicatorsCommand returns one specialize-pass command setting the
// initial on/off state of Caps/Num/Scroll Lock. It writes the same decimal
// bitmask (Caps=1, Num=2, Scroll=4) to InitialKeyboardIndicators under
// Control Panel\Keyboard in both HKU\.DEFAULT (direct, always mounted) and
// HKU\DefaultUser (load/unload, for future accounts) — matching the
// reference implementation's foreach over both roots. Returns "" if s is
// nil (Windows' own defaults apply).
func LockKeyIndicatorsCommand(s *profile.LockKeySettings) string {
	if s == nil {
		return ""
	}
	indicators := 0
	if s.CapsLock.Initial == profile.LockKeyOn {
		indicators |= 1
	}
	if s.NumLock.Initial == profile.LockKeyOn {
		indicators |= 2
	}
	if s.ScrollLock.Initial == profile.LockKeyOn {
		indicators |= 4
	}
	defaultKey := `HKU\.DEFAULT\Control Panel\Keyboard`
	userKey := defaultUserHiveKey + `\Control Panel\Keyboard`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe add "%s" /v InitialKeyboardIndicators /t REG_SZ /d %d /f`, defaultKey, indicators),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v InitialKeyboardIndicators /t REG_SZ /d %d /f`, userKey, indicators),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// buildScancodeMap builds the binary "Scancode Map" registry value that
// remaps each given source scancode to 0x0000 (i.e. disables it): a
// 4-byte Version (0), a 4-byte Flags (0), a 4-byte little-endian Count
// (len(scancodes)+1 — the map itself counts as an entry, per the
// documented format), one 4-byte [target=0x0000, source] entry per
// scancode, and a 4-byte null terminator entry. Returns nil if scancodes
// is empty (nothing to disable).
func buildScancodeMap(scancodes []uint16) []byte {
	if len(scancodes) == 0 {
		return nil
	}
	buf := make([]byte, 0, 12+4*len(scancodes)+4)
	buf = append(buf, 0, 0, 0, 0) // Version
	buf = append(buf, 0, 0, 0, 0) // Flags
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, uint32(len(scancodes)+1))
	buf = append(buf, count...)
	for _, sc := range scancodes {
		buf = append(buf, 0x00, 0x00, byte(sc&0xFF), byte(sc>>8)) // target 0x0000, source sc
	}
	buf = append(buf, 0, 0, 0, 0) // terminating null entry
	return buf
}

// LockKeyScancodeMapCommand returns one specialize-pass command that
// installs a Scancode Map disabling every lock key whose Behavior is
// LockKeyIgnore (takes effect after a reboot, same as upstream). Returns
// "" if s is nil or no key is set to Ignore.
func LockKeyScancodeMapCommand(s *profile.LockKeySettings) string {
	if s == nil {
		return ""
	}
	var scancodes []uint16
	if s.CapsLock.Behavior == profile.LockKeyIgnore {
		scancodes = append(scancodes, scancodeCapsLock)
	}
	if s.NumLock.Behavior == profile.LockKeyIgnore {
		scancodes = append(scancodes, scancodeNumLock)
	}
	if s.ScrollLock.Behavior == profile.LockKeyIgnore {
		scancodes = append(scancodes, scancodeScrollLock)
	}
	mapBytes := buildScancodeMap(scancodes)
	if mapBytes == nil {
		return ""
	}
	return fmt.Sprintf(`cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\Keyboard Layout" /v "Scancode Map" /t REG_BINARY /d %s /f`, hex.EncodeToString(mapBytes))
}
