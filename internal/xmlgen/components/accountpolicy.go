package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// PasswordExpirationCommand returns the `net accounts /maxpwage` command
// for s, or "" when nothing needs to change (Mode is "" or "default":
// Windows' own default of 42 days already applies).
func PasswordExpirationCommand(s profile.PasswordExpirationSettings) string {
	switch s.Mode {
	case profile.PasswordExpirationNever:
		return `net accounts /maxpwage:UNLIMITED`
	case profile.PasswordExpirationCustom:
		if s.Days == nil {
			return ""
		}
		return fmt.Sprintf(`net accounts /maxpwage:%d`, *s.Days)
	default:
		return ""
	}
}

// AccountLockoutCommand returns the `net accounts /lockout...` command for
// s, or "" when nothing needs to change (Mode is "" or "default": Windows'
// own default policy already applies).
func AccountLockoutCommand(s profile.AccountLockoutSettings) string {
	switch s.Mode {
	case profile.AccountLockoutDisabled:
		return `net accounts /lockoutthreshold:0`
	case profile.AccountLockoutCustom:
		if s.Threshold == nil || s.WindowMinutes == nil || s.DurationMinutes == nil {
			return ""
		}
		return fmt.Sprintf(`net accounts /lockoutthreshold:%d /lockoutwindow:%d /lockoutduration:%d`,
			*s.Threshold, *s.WindowMinutes, *s.DurationMinutes)
	default:
		return ""
	}
}
