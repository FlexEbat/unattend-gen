package profile

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// ValidationResult is the outcome of ValidateProfile.
type ValidationResult struct {
	Profile  *Profile // nil if Errors is non-empty
	Errors   []string // Russian-language error messages
	Warnings []string // Russian-language notes about risky but valid settings; only set when Errors is empty
}

var (
	bcp47Re = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
	// Product keys use a 25-character alphabet (no vowels, no 0/1/5/L/I/O/S/U/Z;
	// N is the flag character Windows 8+ keys embed), so a typo such as O for 0
	// is caught here instead of during setup.
	productKeyRe         = regexp.MustCompile(`^(?i)[BCDFGHJKMNPQRTVWXY2346789]{5}(-[BCDFGHJKMNPQRTVWXY2346789]{5}){4}$`)
	computerNameRe       = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,13}[A-Za-z0-9])?$`)
	allDigitsRe          = regexp.MustCompile(`^[0-9]+$`)
	hexColorRe           = regexp.MustCompile(`^[0-9A-Fa-f]{6}$`)
	accountNameForbidden = `"/\[]:;|=,+*?<>`
)

// reservedAccountNames are names Windows reserves for built-in accounts and
// security principals (compared case-insensitively).
var reservedAccountNames = []string{
	"administrator", "guest", "defaultaccount", "wdagutilityaccount",
	"system", "localservice", "networkservice",
}

var structValidator = newStructValidator()

func newStructValidator() *validator.Validate {
	v := validator.New()
	// Report errors by JSON field name (accounts[0].group), not Go field name.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
	return v
}

// ValidateProfile decodes data as JSON and checks it against the schema-version-1
// rules. It does not touch disk and does not build XML.
func ValidateProfile(data []byte) ValidationResult {
	p, err := DecodeProfile(data)
	if err != nil {
		if msg := err.Error(); strings.HasPrefix(msg, "json: unknown field ") {
			return ValidationResult{Errors: []string{"Неизвестное поле в профиле (опечатка?): " + strings.TrimPrefix(msg, "json: unknown field ")}}
		}
		return ValidationResult{Errors: []string{"Профиль не является корректным JSON: " + err.Error()}}
	}

	var errs []string

	if err := structValidator.Struct(p); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				errs = append(errs, structErrorMessage(fe))
			}
		} else {
			errs = append(errs, err.Error())
		}
	}

	errs = append(errs, validateLanguage(p.Language)...)
	errs = append(errs, validateEdition(p.Edition)...)
	errs = append(errs, validateActivationKey(p.ActivationKey)...)
	errs = append(errs, validateComputerName(p.ComputerName)...)
	errs = append(errs, validateTimezone(p.Timezone)...)
	errs = append(errs, validateAccounts(p.Accounts)...)
	errs = append(errs, validateFirstLogon(p.FirstLogon, p.Accounts)...)
	errs = append(errs, validateWifi(p.Wifi)...)
	errs = append(errs, validateComputerNameScript(p.ComputerName, p.ComputerNameScript)...)
	errs = append(errs, checkEnumList(p.RemoveApps, RemovableApps, "remove_apps", "Неизвестное приложение для удаления: ")...)
	errs = append(errs, validateDefaultUserScripts(p.DefaultUserScripts)...)
	errs = append(errs, validateScripts("system_scripts", p.SystemScripts)...)
	errs = append(errs, validateScripts("default_user_scripts", p.DefaultUserScripts)...)
	errs = append(errs, validateScripts("first_logon_scripts", p.FirstLogonScripts)...)
	errs = append(errs, validateScripts("user_once_scripts", p.UserOnceScripts)...)
	errs = append(errs, checkEnumList(p.RemoveFeatures, RemovableFeatures, "remove_features", "Неизвестный компонент для удаления: ")...)
	errs = append(errs, checkEnumList(p.RemoveOptionalFeatures, RemovableOptionalFeatures, "remove_optional_features", "Неизвестный необязательный компонент для удаления: ")...)
	errs = append(errs, validatePersonalization(p.Personalization)...)
	errs = append(errs, validatePasswordExpiration(p.PasswordExpiration)...)
	errs = append(errs, validateAccountLockout(p.AccountLockout)...)
	errs = append(errs, validateStickyKeys(p.StickyKeys)...)
	errs = append(errs, checkEnumMapKeys(p.DesktopIcons, DesktopIcons, "Неизвестный значок рабочего стола: ")...)
	errs = append(errs, validateVisualEffects(p.VisualEffects)...)
	errs = append(errs, validateStartPins(p.StartPins)...)
	errs = append(errs, validateStartTiles(p.StartTiles)...)
	errs = append(errs, validateTaskbarIcons(p.TaskbarIcons)...)
	errs = append(errs, checkEnumList(p.StartFolders, StartFolders, "start_folders", "Неизвестная папка для Start: ")...)
	errs = append(errs, checkEnumList(p.InstallVMGuestTools, VMGuestTools, "install_vm_guest_tools", "Неизвестный набор гостевых дополнений ВМ: ")...)
	errs = append(errs, validateAppLockerPolicyXML(p.AppLockerPolicyXML)...)

	if len(errs) > 0 {
		return ValidationResult{Errors: errs}
	}
	return ValidationResult{Profile: p, Warnings: profileWarnings(p)}
}

// structErrorMessage turns a go-playground/validator failure into a Russian
// message that names the offending field by its JSON path.
func structErrorMessage(fe validator.FieldError) string {
	path := strings.TrimPrefix(fe.Namespace(), "Profile.")
	switch path {
	case "schema_version":
		return "Поле schema_version должно быть равно 1"
	case "name":
		return "Имя профиля обязательно"
	}
	switch fe.Tag() {
	case "required":
		return path + ": обязательное поле"
	case "oneof":
		return path + ": допустимые значения: " + strings.Join(strings.Fields(fe.Param()), ", ")
	case "max":
		if fe.Kind() == reflect.Slice {
			return path + ": не больше " + fe.Param() + " элементов"
		}
		return path + ": не больше " + fe.Param() + " символов"
	case "min":
		return path + ": не меньше " + fe.Param() + " символов"
	default:
		return path + ": некорректное значение"
	}
}

// checkEnumList reports values that are not in allowed, and values listed more
// than once. field is the JSON field name, unknownMsg the message prefix for an
// unknown value.
func checkEnumList[T ~string](values, allowed []T, field, unknownMsg string) []string {
	var errs []string
	seen := make(map[T]bool, len(values))
	for _, v := range values {
		if !slices.Contains(allowed, v) {
			errs = append(errs, unknownMsg+string(v))
			continue
		}
		if seen[v] {
			errs = append(errs, field+": значение "+string(v)+" указано больше одного раза")
		}
		seen[v] = true
	}
	return errs
}

// checkEnumMapKeys reports map keys that are not in allowed, in sorted order.
func checkEnumMapKeys[K ~string, V any](m map[K]V, allowed []K, unknownMsg string) []string {
	var unknown []string
	for k := range m {
		if !slices.Contains(allowed, k) {
			unknown = append(unknown, string(k))
		}
	}
	sort.Strings(unknown)
	errs := make([]string, len(unknown))
	for i, k := range unknown {
		errs[i] = unknownMsg + k
	}
	return errs
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

func validateActivationKey(activationKey *string) []string {
	if activationKey == nil {
		return nil
	}
	if !productKeyRe.MatchString(*activationKey) {
		return []string{"activation_key должен быть вида XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"}
	}
	return nil
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
	seen := make(map[string]bool, len(accounts))
	for _, a := range accounts {
		// An empty name and a name over 20 characters are reported by the
		// struct tags (required, max=20, counted in characters).
		if a.Name == "" {
			continue
		}
		if strings.ContainsAny(a.Name, accountNameForbidden) || strings.ContainsFunc(a.Name, unicode.IsControl) {
			errs = append(errs, "Имя учётной записи "+a.Name+" содержит недопустимые символы")
		}
		if strings.HasSuffix(a.Name, ".") {
			errs = append(errs, "Имя учётной записи "+a.Name+" не может заканчиваться точкой")
		}
		if strings.TrimSpace(a.Name) != a.Name {
			errs = append(errs, "Имя учётной записи "+a.Name+" не может начинаться или заканчиваться пробелом")
		}
		lower := strings.ToLower(a.Name)
		if slices.Contains(reservedAccountNames, lower) {
			errs = append(errs, "Имя учётной записи "+a.Name+" зарезервировано Windows")
		}
		if seen[lower] {
			errs = append(errs, "Имя учётной записи "+a.Name+" указано больше одного раза (регистр не учитывается)")
		}
		seen[lower] = true
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
	if w.RawProfileXML != nil {
		if strings.TrimSpace(*w.RawProfileXML) == "" {
			return []string{"raw_profile_xml не может быть пустой строкой (используйте null)"}
		}
		return validateWellFormedXML(*w.RawProfileXML, "raw_profile_xml")
	}
	var errs []string
	if len(w.SSID) < 1 || len(w.SSID) > 32 {
		errs = append(errs, "SSID не может быть пустым и должен быть до 32 символов")
	}
	if w.Authentication == "" {
		errs = append(errs, "authentication обязателен, если raw_profile_xml не задан")
	}
	if w.Authentication != WifiOpen {
		if w.Password == nil || len(*w.Password) < 8 {
			errs = append(errs, "Пароль Wi-Fi обязателен и должен быть не короче 8 символов для WPA2/WPA3")
		} else if w.Authentication == WifiWPA2Personal && utf8.RuneCountInString(*w.Password) > 63 {
			errs = append(errs, "Пароль Wi-Fi для WPA2 не может быть длиннее 63 символов")
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

func validateStickyKeys(s StickyKeysSettings) []string {
	if s.Mode != StickyKeysModeCustom {
		return nil
	}
	return checkEnumList(s.Flags, StickyKeysFlags, "sticky_keys.flags", "Неизвестный флаг Sticky Keys: ")
}

func validateVisualEffects(s VisualEffectsSettings) []string {
	if s.Mode != VisualEffectsModeCustom {
		return nil
	}
	return checkEnumMapKeys(s.Custom, VisualEffects, "Неизвестный визуальный эффект: ")
}

func validateStartPins(s StartPinsSettings) []string {
	if s.Mode != StartPinsModeCustom {
		return nil
	}
	if s.JSON == nil || strings.TrimSpace(*s.JSON) == "" {
		return []string{"Для режима custom у start_pins требуется непустой json"}
	}
	if !json.Valid([]byte(*s.JSON)) {
		return []string{"start_pins.json не является валидным JSON"}
	}
	return nil
}

func validateStartTiles(s StartTilesSettings) []string {
	if s.Mode != StartTilesModeCustom {
		return nil
	}
	if s.XML == nil || strings.TrimSpace(*s.XML) == "" {
		return []string{"Для режима custom у start_tiles требуется непустой xml"}
	}
	return validateWellFormedXML(*s.XML, "start_tiles.xml")
}

func validateTaskbarIcons(s TaskbarIconsSettings) []string {
	if s.Mode != TaskbarIconsModeCustom {
		return nil
	}
	if s.XML == nil || strings.TrimSpace(*s.XML) == "" {
		return []string{"Для режима custom у taskbar_icons требуется непустой xml"}
	}
	return validateWellFormedXML(*s.XML, "taskbar_icons.xml")
}

func validateComputerNameScript(computerName, computerNameScript *string) []string {
	if computerNameScript == nil {
		return nil
	}
	var errs []string
	if computerName != nil {
		errs = append(errs, "computer_name и computer_name_script нельзя задавать одновременно")
	}
	if strings.TrimSpace(*computerNameScript) == "" {
		errs = append(errs, "computer_name_script не может быть пустой строкой (используйте null)")
	}
	return errs
}

func validateWellFormedXML(xmlContent string, fieldName string) []string {
	dec := xml.NewDecoder(strings.NewReader(xmlContent))
	for {
		_, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return []string{fieldName + " не является валидным XML: " + err.Error()}
		}
	}
}

func validateAppLockerPolicyXML(policyXML *string) []string {
	if policyXML == nil {
		return nil
	}
	if strings.TrimSpace(*policyXML) == "" {
		return []string{"applocker_policy_xml не может быть пустой строкой (используйте null)"}
	}
	return validateWellFormedXML(*policyXML, "applocker_policy_xml")
}

// validateScripts rejects scripts whose content is empty or only whitespace
// (the struct tag only catches the empty string).
func validateScripts(field string, scripts []CustomScript) []string {
	var errs []string
	for i, sc := range scripts {
		if sc.Content != "" && strings.TrimSpace(sc.Content) == "" {
			errs = append(errs, fmt.Sprintf("%s[%d]: содержимое скрипта не может состоять только из пробелов", field, i))
		}
	}
	return errs
}
