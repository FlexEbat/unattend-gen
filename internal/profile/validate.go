package profile

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationResult is the outcome of ValidateProfile.
type ValidationResult struct {
	Profile *Profile // nil if Errors is non-empty
	Errors  []string // Russian-language error messages
}

var (
	bcp47Re              = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
	productKeyRe         = regexp.MustCompile(`^[A-Za-z0-9]{5}(-[A-Za-z0-9]{5}){4}$`)
	computerNameRe       = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,13}[A-Za-z0-9])?$`)
	allDigitsRe          = regexp.MustCompile(`^[0-9]+$`)
	hexColorRe           = regexp.MustCompile(`^[0-9A-Fa-f]{6}$`)
	accountNameForbidden = `"/\[]:;|=,+*?<>`
)

var structValidator = validator.New()

// ValidateProfile decodes data as JSON and checks it against the schema-version-1
// rules in tech.md section 8. It does not touch disk and does not build XML.
func ValidateProfile(data []byte) ValidationResult {
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return ValidationResult{Errors: []string{"Профиль не является корректным JSON: " + err.Error()}}
	}

	var errs []string

	if err := structValidator.Struct(&p); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range verrs {
				errs = append(errs, structErrorMessage(fe))
			}
		} else {
			errs = append(errs, err.Error())
		}
	}

	errs = append(errs, validateLanguage(p.Language)...)
	errs = append(errs, validateEdition(p.Edition)...)
	errs = append(errs, validateComputerName(p.ComputerName)...)
	errs = append(errs, validateTimezone(p.Timezone)...)
	errs = append(errs, validateAccounts(p.Accounts)...)
	errs = append(errs, validateFirstLogon(p.FirstLogon, p.Accounts)...)
	errs = append(errs, validateWifi(p.Wifi)...)
	errs = append(errs, validateRemoveApps(p.RemoveApps)...)
	errs = append(errs, validateDefaultUserScripts(p.DefaultUserScripts)...)
	errs = append(errs, validateRemoveFeatures(p.RemoveFeatures)...)
	errs = append(errs, validateRemoveOptionalFeatures(p.RemoveOptionalFeatures)...)
	errs = append(errs, validatePersonalization(p.Personalization)...)
	errs = append(errs, validatePasswordExpiration(p.PasswordExpiration)...)
	errs = append(errs, validateAccountLockout(p.AccountLockout)...)

	if len(errs) > 0 {
		return ValidationResult{Errors: errs}
	}
	return ValidationResult{Profile: &p}
}

func structErrorMessage(fe validator.FieldError) string {
	switch fe.Field() {
	case "SchemaVersion":
		return "Поле schema_version должно быть равно 1"
	case "Name":
		return "Имя профиля обязательно"
	default:
		return fe.Field() + ": некорректное значение"
	}
}

func validateLanguage(l LanguageSettings) []string {
	var errs []string
	if !bcp47Re.MatchString(l.UILanguage) {
		errs = append(errs, "Язык интерфейса (ui_language) должен быть в формате BCP-47")
	}
	if !bcp47Re.MatchString(l.Locale) {
		errs = append(errs, "Региональный формат (locale) должен быть в формате BCP-47")
	}
	if !bcp47Re.MatchString(l.KeyboardLayout) {
		errs = append(errs, "Раскладка клавиатуры (keyboard_layout) должна быть в формате BCP-47")
	}
	return errs
}

func validateEdition(e EditionSettings) []string {
	var errs []string
	switch e.Mode {
	case EditionModeGenericKey:
		if e.Edition == nil {
			errs = append(errs, "Для режима generic_key требуется указать edition")
		}
	case EditionModeCustomKey:
		if e.ProductKey == nil || !productKeyRe.MatchString(*e.ProductKey) {
			errs = append(errs, "Для режима custom_key требуется product_key вида XXXXX-XXXXX-XXXXX-XXXXX-XXXXX")
		}
	}
	return errs
}

// validateTimezone only rejects the empty string: Windows time zone IDs
// ("Russian Standard Time", "UTC", ...) form a large, version-dependent
// list we don't try to replicate here.
func validateTimezone(tz *string) []string {
	if tz == nil {
		return nil
	}
	if *tz == "" {
		return []string{"Часовой пояс не может быть пустой строкой, используйте null"}
	}
	return nil
}

func validateComputerName(name *string) []string {
	if name == nil {
		return nil
	}
	n := *name
	var errs []string
	if len(n) < 1 || len(n) > 15 {
		errs = append(errs, "Имя компьютера длиннее 15 символов")
		return errs
	}
	if !computerNameRe.MatchString(n) {
		errs = append(errs, "Имя компьютера содержит недопустимые символы или начинается/заканчивается дефисом")
	}
	if allDigitsRe.MatchString(n) {
		errs = append(errs, "Имя компьютера не может состоять только из цифр")
	}
	return errs
}

func validateAccounts(accounts []UserAccount) []string {
	var errs []string
	if len(accounts) > 5 {
		errs = append(errs, "Учётных записей не может быть больше 5")
	}
	for _, a := range accounts {
		if a.Name == "" {
			errs = append(errs, "Имя учётной записи не может быть пустым")
			continue
		}
		if len(a.Name) > 20 {
			errs = append(errs, "Имя учётной записи "+a.Name+" длиннее 20 символов")
		}
		if strings.ContainsAny(a.Name, accountNameForbidden) {
			errs = append(errs, "Имя учётной записи "+a.Name+" содержит недопустимые символы")
		}
		if a.Password != nil && *a.Password == "" {
			errs = append(errs, "Пароль учётной записи "+a.Name+" не может быть пустой строкой, используйте null")
		}
	}
	return errs
}

func validateFirstLogon(fl FirstLogon, accounts []UserAccount) []string {
	var errs []string
	switch fl.Mode {
	case FirstLogonFirstCreatedAccount:
		if len(accounts) == 0 {
			errs = append(errs, "Режим first_created_account требует хотя бы одной учётной записи")
		}
	case FirstLogonBuiltinAdmin:
		if fl.BuiltinAdministratorPassword == nil || *fl.BuiltinAdministratorPassword == "" {
			errs = append(errs, "Пароль обязателен для встроенной учётной записи Administrator")
		}
	}
	return errs
}

func validateWifi(w *WifiSettings) []string {
	if w == nil {
		return nil
	}
	var errs []string
	if len(w.SSID) < 1 || len(w.SSID) > 32 {
		errs = append(errs, "SSID не может быть пустым и должен быть до 32 символов")
	}
	if w.Authentication != WifiOpen {
		if w.Password == nil || len(*w.Password) < 8 {
			errs = append(errs, "Пароль Wi-Fi обязателен и должен быть не короче 8 символов для WPA2/WPA3")
		}
	}
	return errs
}

func validateRemoveApps(apps []RemovableApp) []string {
	var errs []string
	for _, a := range apps {
		known := false
		for _, allowed := range RemovableApps {
			if a == allowed {
				known = true
				break
			}
		}
		if !known {
			errs = append(errs, "Неизвестное приложение для удаления: "+string(a))
		}
	}
	return errs
}

func validateDefaultUserScripts(scripts []CustomScript) []string {
	var errs []string
	for _, s := range scripts {
		if s.Format == ScriptVbs {
			errs = append(errs, "Скрипты DefaultUser не поддерживают формат .vbs")
		}
	}
	return errs
}

func validateRemoveFeatures(features []RemovableFeature) []string {
	var errs []string
	for _, f := range features {
		known := false
		for _, allowed := range RemovableFeatures {
			if f == allowed {
				known = true
				break
			}
		}
		if !known {
			errs = append(errs, "Неизвестный компонент для удаления: "+string(f))
		}
	}
	return errs
}

func validateRemoveOptionalFeatures(features []RemovableOptionalFeature) []string {
	var errs []string
	for _, f := range features {
		known := false
		for _, allowed := range RemovableOptionalFeatures {
			if f == allowed {
				known = true
				break
			}
		}
		if !known {
			errs = append(errs, "Неизвестный необязательный компонент для удаления: "+string(f))
		}
	}
	return errs
}

func validatePasswordExpiration(s PasswordExpirationSettings) []string {
	if s.Mode != PasswordExpirationCustom {
		return nil
	}
	if s.Days == nil || *s.Days < 1 {
		return []string{"Для custom-режима срока действия пароля нужно указать days >= 1"}
	}
	return nil
}

func validateAccountLockout(s AccountLockoutSettings) []string {
	if s.Mode != AccountLockoutCustom {
		return nil
	}
	var errs []string
	if s.Threshold == nil || *s.Threshold < 1 {
		errs = append(errs, "Для custom-политики блокировки нужно указать threshold >= 1")
	}
	if s.WindowMinutes == nil || *s.WindowMinutes < 1 {
		errs = append(errs, "Для custom-политики блокировки нужно указать window_minutes >= 1")
	}
	if s.DurationMinutes == nil || *s.DurationMinutes < 1 {
		errs = append(errs, "Для custom-политики блокировки нужно указать duration_minutes >= 1")
	}
	return errs
}

func validatePersonalization(s PersonalizationSettings) []string {
	var errs []string
	if s.AccentColor != nil && !hexColorRe.MatchString(*s.AccentColor) {
		errs = append(errs, "accent_color должен быть 6 hex-цифрами, например FF8800")
	}
	if s.SolidColorWallpaper != nil && !hexColorRe.MatchString(*s.SolidColorWallpaper) {
		errs = append(errs, "solid_color_wallpaper должен быть 6 hex-цифрами, например 0078D4")
	}
	return errs
}
