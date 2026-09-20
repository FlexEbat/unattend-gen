# tech.md — unattend-gen

**Версия: v28** (2026-09-11)

Changelog:
- v28 — слайс 26: Start menu/taskbar, достижимая часть (8 из 9 под-пунктов, раздел сайта вне исходных 6 групп) — новый файл `internal/xmlgen/components/taskbar.go`, новый экран `screens.ScreenTaskbar` (между VisualEffects и Advanced). `SystemTweaks` 28→33 (5 новых), `Profile.TaskbarSearch`/`StartPins`/`StartTiles`. Единственный оставшийся под-пункт (кастомные закреплённые иконки таскбара, locked-layout+scheduled-task-unlock) сознательно не взят — существенно сложнее остальных 8 вместе взятых. Раздел 4/9.3/9.16.
- v27 — слайс 25: Visual Effects (раздел сайта вне исходных 6 групп) — новый файл `internal/xmlgen/components/visualeffects.go`, новый экран `screens.ScreenVisualEffects` (между Desktop и Advanced), `Profile.VisualEffects VisualEffectsSettings` (4 режима, 17 эффектов). Побутно найден и исправлен баг слайса 21: `ScreenAdvanced` отсутствовал в `rebuildScreen`, переход туда всегда показывал Review. Раздел 4/9.15.
- v26 — слайс 24: достижимая часть группы D — `ActivationKey`, `EditionModeFirmware` (BIOS/UEFI ключ), `ProcessorArchitecture` (единственное значение, не мульти-архитектура). Важные находки: «без интернета» дублировал уже реализованный `BypassOnlineAccountRequirement`; 7 других под-пунктов группы D закрыты как «неприменимо при текущем архитектурном решении» (все — части одного механизма замены PE-этапа своим `.cmd`-скриптом, конфликтующего с исключением диск-партиционирования). `UserData.ProductKey` стал указателем. Раздел 4/9.14.
- v25 — слайс 23: `SystemTweaks.DeleteEdgeDesktopIcon` (28-е поле, побочная находка слайса 20) + `Profile.HidePowerShellWindows` (последний, 6-й пункт группы E — ЗАКРЫТА ПОЛНОСТЬЮ). `invokeCommand` получил параметр `hidden bool`; протащен через 13 функций в 5 файлах (`scripts.go`, `apps.go`, `misc.go`, `optimizations.go`, `vmapplocker.go`) + сигнатуры `NewDeployment`/`NewShellSetupOOBE`. Раздел 9.3/9.13.
- v24 — слайс 22 (tech.md backlog group E, 5/6, ЗАКРЫТА почти полностью): новый файл `internal/xmlgen/components/misc.go`. `Profile.KeepSensitiveFiles` — единственное поле в схеме, где zero value вызывает действие (удаление), а не «ничего не менять» — сознательное отступление от общей конвенции, задокументировано и защищено тестовой фикстурой `baseProfile()` (там явно `true`). `UseNarrator`, `ComputerNameScript` (взаимоисключимо с `ComputerName`), `ObscurePasswords` (новый helper `newPasswordElement`, заменил 4 места конструирования пароля), `WifiSettings.RawProfileXML`. Экран Advanced расширен, экран Wifi получил raw-режим.
- v23 — слайс 21 (tech.md backlog group C, +3/9, ЗАКРЫТА ПОЛНОСТЬЮ 9/9): новый файл `internal/xmlgen/components/vmapplocker.go` (VM guest tools + AppLocker), новый тип `Profile.InstallVMGuestTools []VMGuestTool` + `Profile.AppLockerPolicyXML *string`, `SystemTweaks.DisableCoreIsolation` (27-е поле), новый экран `screens.ScreenAdvanced` между Desktop и Scripts. AppLocker валидируется только на well-formedness XML, без XSD-схемы (сознательное упрощение, задокументировано в 9.11). Замечено, но не взято: `TaskbarAl`/left-align таскбара (простой reg add, в бэклоге группы Start menu/taskbar).
- v22 — слайс 20 (tech.md backlog group C, +4/9, суммарно 6/9 закрыто): новый файл `internal/xmlgen/components/startmenu.go` (Desktop Icons + Folders on Start), новый тип `Profile.DesktopIcons map[DesktopIcon]bool` + `Profile.StartFolders []StartFolder`, новый экран `screens.ScreenDesktop` между Accessibility и Scripts. Оба механизма впервые используют RunOnce (не default-user-hive-контент напрямую) для правки живого HKCU нового аккаунта — новый паттерн, задокументирован в 9.10. Замечен, но сознательно не взят в этот слайс: `DeleteEdgeDesktopIcon` (отдельный простой твик, в бэклоге).
- v21 — слайс 19 (tech.md backlog group C, 2 из 9 пунктов): новый файл `internal/xmlgen/components/accessibility.go` (Sticky Keys + Lock Keys), новый тип `StickyKeysSettings`/`LockKeySettings`+`Profile.LockKeys *LockKeySettings`, новый экран `screens.ScreenAccessibility` между Personalization и Scripts. Оба механизма пишут в `HKU\DefaultUser` (будущие аккаунты) и `HKU\.DEFAULT` (текущая сессия/экран блокировки) параллельно. Раздел 3/4/7 обновлены, новый раздел 9.9.
- v20 — слайс 18 (tech.md backlog group B, остаток — ЗАКРЫТА полностью): `SystemTweaks` 23→26, новые — `HardenSystemDriveACL`, `MakeEdgeUninstallable`, `DeleteWindowsOld`. Сигнатура `NewShellSetupOOBE` выросла ещё на один параметр (`deleteWindowsOld` — единственный из троицы, который живёт в FirstLogonCommands, а не в Deployment/specialize). Раздел 9.3 дополнен.
- v19 — слайс 17 (tech.md backlog group A, largely closed): `RemovableApp` 32→42 (10 новых Appx + `AppOneDrive` через новый custom-механизм), `RemovableFeature` 7→11 (4 новых DISM capability), новый тип `RemovableOptionalFeature` (3 значения) + новый файл `optionalfeatures.go` — третий механизм удаления (`Disable-WindowsOptionalFeature`), закрывший Media Features/Recall/Remote Desktop Client. Побочный фикс: `FeatureSpeech` получил второй capability-селектор. Раздел 4/9.6 переписаны.
- v18 — слайс 16 (tech.md backlog group B, частично): `SystemTweaks` 17→23 поля, новые — `DeleteHiddenJunctions`, `PreventAutomaticReboot`, `TurnOffSystemSounds`, `DisableAppSuggestions`, `DisablePointerPrecision`, `PreventDeviceApps`. Новый файл `internal/xmlgen/components/optimizations.go`, раздел 9.3 расписан подробно.
- v17 — раздел 15 (бэклог) переписан по итогам полной постраничной сверки с schneegans.de (commit `88d81f0`): раскрыт по группам A–F с приоритетом, найдено ~30 новых незадокументированных гэпов (расширенный Windows PE stage, Activation, Processor architectures, Start menu/taskbar целиком, Lock key settings, Sticky keys, Folders on Start, VM host core isolation, ~18 недостающих приложений в Remove bloatware, ещё 6 System tweaks). Уточнено: Windows Fax and Scan больше не в списке сайта, убран из бэклога.
- v16 — восстановлен после удаления из репозитория. Актуализирован по фактическому коду: слайсы 0–15 сделаны (последний — v0.15.0, solid-цвет обоев). Заморожены: `Profile` (внутри — `SystemTweaks`, `WifiSettings`, `RemovableApp`/`RemovableFeature`, `CustomScript`, `PasswordExpirationSettings`, `AccountLockoutSettings`, `FileExplorerSettings`, `PersonalizationSettings`), API `profile.ValidateProfile`/`xmlgen.BuildAnswerFile`, структура папок, слои `cli → tui/xmlgen ← profile`.
- v1 — первая версия контракта (слайсы 0–6: каркас, язык/издание, компьютер/аккаунты, CLI end-to-end, express settings + tweaks, TUI + пресеты, Wi-Fi).

---

## 0. Как читать этот файл

Это источник истины проекта. Читай его перед каждой задачей и подчиняйся дословно.

- Имена полей, типов, функций и путей берутся отсюда. Свои варианты не придумывай.
- Нужного контракта здесь нет → СТОП, выдай блок `CONTRACT GAP` (раздел 13). Код с выдуманным XML-элементом или полем не пиши.
- Файл меняет только владелец проекта. Каждое изменение контракта поднимает версию сверху файла.
- Работаешь один слайс за заход. История слайсов и бэклог — раздел 15.
- Раздел 4 описывает данные как они есть в коде сейчас (`internal/profile/schema.go`), а не как задумывались изначально — при расхождении верен код, а этот файл поднимает версию и подгоняется под код.

---

## 1. Проект

CLI/TUI-инструмент на Go, генерирующий Windows 10/11 `autounattend.xml` (unattended-install answer file). Цель — полный (1:1, по функциональности) паритет с [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/), но в виде локального CLI/TUI вместо веб-формы: пользователь один раз собирает профиль под себя, дальше Windows ставится без диалогов.

Репозиторий: `github.com/FlexEbat/unattend-gen`. Собирается в один статический бинарник, без сети и без сервера в рантайме.

Явное исключение из скоупа (решение владельца, не пробел, который надо закрывать): разметка диска (`DiskConfiguration`) не реализуется — Windows Setup всегда спрашивает, куда ставить, интерактивно.

---

## 2. Стек

- Go 1.23
- `github.com/spf13/cobra` — CLI
- `github.com/charmbracelet/bubbletea` + `bubbles` + `lipgloss` — TUI
- `github.com/go-playground/validator/v10` — часть валидации профиля (структурные теги), поверх — ручные Russian-language проверки
- Тесты: `go test ./...`, стандартная библиотека (`testing`), пакетные `builder_*_test.go` в `xmlgen`, `charmbracelet/x/exp/teatest` для TUI
- Проверки: `gofmt`, `go vet`, `golangci-lint` (только `govet`, `staticcheck`, `unused`, `errcheck` — см. `.golangci.yml`)

Без ORM (нет базы), без внешних HTTP-клиентов, без сети в рантайме. Пресеты профилей встраиваются в бинарник через `go:embed`.

---

## 3. Структура папок

```
cmd/
  unattend-gen/
    main.go                          вызывает cli.Execute()
internal/
  cli/
    root.go                          NewRootCmd/Execute, регистрирует подкоманды
    profile.go                       `profile init`, `profile list`
    validate.go                      `validate <profile.json>`
    generate.go                      `generate <profile.json>`
    tui.go                           `tui [profile.json]`
    *_test.go
  profile/
    schema.go                        типы: Profile и все вложенные *Settings/enum'ы
    validate.go                      ValidateProfile(data []byte) ValidationResult
    store.go                         LoadProfile/SaveProfile/ListProfiles, ProfilesDir="profiles"
    *_test.go
  tui/
    app.go                           NewModel, Model (bubbletea), таблица screens.ID → tea.Model
    screens/
      welcome.go, language.go, accounts.go, tweaks.go, wifi.go,
      apps.go, personalization.go, accessibility.go, desktop.go, visualeffects.go, taskbar.go, advanced.go, scripts.go, review.go
      nav.go                         screens.ID, навигационные сообщения
    widgets/
      labeled_input.go, password_input.go, labeled_select.go,
      labeled_textarea.go, checkbox.go, accounts_table.go, confirm_bar.go
  xmlgen/
    builder.go                       BuildAnswerFile(*profile.Profile) (string, error)
    components/
      international.go               International-Core (windowsPE + specialize)
      setup.go                       Microsoft-Windows-Setup: UserData, BypassWin11Requirements
      shellsetup.go                  Shell-Setup: ComputerName/TimeZone (specialize), OOBE+FirstLogonCommands (oobeSystem)
      wlan.go                        Wi-Fi через netsh WLAN-профиль (base64) в FirstLogonCommands
      apps.go                        удаление Appx через Get/Remove-AppxProvisionedPackage
      features.go                    удаление DISM-компонентов через Get/Remove-WindowsCapability
      optionalfeatures.go            удаление legacy optional features через Get/Disable-WindowsOptionalFeature
      scripts.go                     System/DefaultUser/FirstLogon/UserOnce скрипты, wrapCommand/escapeForOuterCommand
      accountpolicy.go               `net accounts` (истечение пароля, блокировка)
      fileexplorer.go                реестр в default-user hive (видимость файлов, контекстное меню и т.д.)
      personalization.go             реестр в default-user hive (тема, акцентный цвет, обои)
      optimizations.go               SystemTweaks-твики со слайсов 16/18 (junctions, active-hours, sounds, ACL, Edge uninstallable и т.д.)
      accessibility.go               Sticky Keys + Lock Keys (default-user hive + HKU\.DEFAULT + Scancode Map)
      startmenu.go                   Desktop Icons + Folders on Start (RunOnce → живой HKCU нового аккаунта)
      vmapplocker.go                 VM guest tools (4 ps1-скрипта) + AppLocker (Set-AppLockerPolicy)
      visualeffects.go               Visual effects presets (best appearance/performance/custom, слайс 25)
      taskbar.go                     Start menu/taskbar достижимая часть (слайс 26)
      misc.go                        Keep sensitive files, Narrator, ComputerNameScript (слайс 22)
    builder_*_test.go                по одному файлу тестов на каждый компонент/срез функциональности
presets/
  presets.go                         go:embed, Names, Load(name)
  minimal.json, single-user.json
go.mod, go.sum, Makefile, .golangci.yml, README.md
```

Правила размещения (слои):

- `internal/xmlgen` никогда не читает диск и не импортирует `cobra`/`bubbletea`. Вход — валидный `*profile.Profile`, выход — строка XML либо ошибка сериализации.
- CLI и TUI никогда не строят XML напрямую — только через `xmlgen.BuildAnswerFile`.
- `internal/profile/validate.go` не импортирует `xmlgen`.
- `internal/profile/store.go` только читает/пишет JSON на диск, валидацию не делает.
- Новые файлы за пределами этого списка — только под новый функциональный срез (например, новый `components/*.go` под новый раздел настроек Windows), не под рефакторинг ради рефакторинга.

---

## 4. Контракт данных (заморожен)

Корневой тип — `profile.Profile` (`internal/profile/schema.go`), JSON, `schema_version: 1`.

```go
type Profile struct {
    SchemaVersion                  int                         // = 1
    Name                           string                      // required
    Language                       LanguageSettings
    Edition                        EditionSettings
    ComputerName                   *string                     // nil = Windows сам сгенерирует
    Timezone                       *string                     // nil = автоопределение; иначе Windows time zone ID ("Russian Standard Time")
    Accounts                       []UserAccount               // max 5
    FirstLogon                     FirstLogon
    ExpressSettings                ExpressSettings
    SystemTweaks                   SystemTweaks
    Wifi                           *WifiSettings               // nil = Wi-Fi не настраивается
    BypassOnlineAccountRequirement bool
    RemoveApps                     []RemovableApp
    RemoveFeatures                 []RemovableFeature
    RemoveOptionalFeatures         []RemovableOptionalFeature
    PasswordExpiration             PasswordExpirationSettings
    AccountLockout                 AccountLockoutSettings
    FileExplorer                   FileExplorerSettings
    Personalization                PersonalizationSettings
    SystemScripts                  []CustomScript              // max 4, контекст system, до создания аккаунтов
    DefaultUserScripts             []CustomScript               // max 3, без .vbs, пишутся в default-user hive
    FirstLogonScripts              []CustomScript               // max 4, oobeSystem FirstLogonCommands, один раз
    UserOnceScripts                []CustomScript               // max 4, RunOnce в default-user hive — на каждого нового пользователя
    RestartExplorerAfterScripts    bool
}
```

Ключевые вложенные типы (полные определения — в `schema.go`, здесь только контракт, который нельзя менять без ревизии версии файла):

- `LanguageSettings{UILanguage, Locale, KeyboardLayout string}` — все три обязательны, формат BCP-47.
- `EditionSettings{Mode: generic_key|custom_key|interactive|firmware, Edition *WindowsEdition, ProductKey *string}` — `firmware` (слайс 24) использует ключ, уже встроенный в BIOS/UEFI прошивку устройства (типично для OEM-предустановок), без каких-либо доп. полей.
- `Profile.ActivationKey *string` (слайс 24) — отдельный ключ ТОЛЬКО для активации (`Microsoft-Windows-Shell-Setup/ProductKey`, specialize), независимый от установочного ключа выше (`UserData/ProductKey`, windowsPE). `nil` + `Edition.Mode=custom_key` переиспользует тот ключ для активации тоже (как было и раньше, теперь явно через `components.ResolveActivationKey`); `nil` при остальных режимах — ничего не пишется.
- `Profile.ProcessorArchitecture ProcessorArchitecture` (слайс 24) — `amd64|x86|arm64`, `""` = `amd64` по умолчанию. Сознательное упрощение относительно сайта-эталона: тот поддерживает НЕСКОЛЬКО архитектур в одном XML (весь документ дублируется на каждую) — у нас XML собирается из типизированных Go-структур, а не пост-обработкой DOM, так что честная поддержка мульти-архитектуры потребовала бы отдельного крупного рефакторинга сериализации; выбрано единственное значение, покрывающее подавляющее большинство реальных сценариев (один образ — одна архитектура).
- `VisualEffectsSettings{Mode: default|best_appearance|best_performance|custom, Custom map[VisualEffect]bool}` (слайс 25) — 17 значений `VisualEffect` (`profile.VisualEffects`); `Custom` значим только при `Mode=custom`, необозначенные эффекты сохраняют дефолт Windows.
- `Profile.TaskbarSearch TaskbarSearchMode` (слайс 26) — `hide|icon|box|label`, `""` = `box` (дефолт Windows).
- `StartPinsSettings{Mode: default|empty|custom, JSON *string}` (слайс 26) — влияет только на Win11 (`ConfigureStartPins` policy); `JSON` обязателен при `Mode=custom`, валидируется на well-formed JSON.
- `StartTilesSettings{Mode: default|empty|custom, XML *string}` (слайс 26) — влияет только на Win10 (`LayoutModification.xml`); `XML` обязателен при `Mode=custom`, валидируется на well-formed XML (переиспользует `validateWellFormedXML`).
- `UserAccount{Name string (≤20), DisplayName *string, Password *string (nil=без пароля, "" запрещено), Group: Administrators|Users}`.
- `FirstLogon{Mode: first_created_account|builtin_administrator|none, BuiltinAdministratorPassword *string}`.
- `ExpressSettings{Mode: all_disabled|all_enabled|interactive}`.
- `SystemTweaks` — 33 булевых поля (17 из слайса 8 + 6 из слайса 16 + 3 из слайса 18 + 1 из слайса 21 + 1 из слайса 23 + 5 из слайса 26, см. раздел 9.3), все опциональны, zero value = ничего не меняется.
- `WifiSettings{SSID (≤32), Authentication: Open|WPA2Personal|WPA3Personal, Password *string, ConnectHidden bool, RawProfileXML *string}` — `RawProfileXML` (слайс 22), если задан, используется как есть, остальные поля становятся необязательными (проверка в `validateWifi`, не в struct-тегах).
- `RemovableApp` — строковый enum, 42 значения (31 из слайса 9 + 10 простых из слайса 17 + `OneDrive`, у которого другой механизм, см. раздел 9.6; список — `profile.RemovableApps`, порядок = порядок в TUI).
- `RemovableFeature` — строковый enum, 11 значений (`profile.RemovableFeatures`): InternetExplorer, WordPad, PowerShellISE, OpenSSHClient, MediaPlayer, Speech, Handwriting (слайс 11) + WindowsHello, MathInputPanel, OneSync, StepsRecorder (слайс 17).
- `RemovableOptionalFeature` — строковый enum, 3 значения (`profile.RemovableOptionalFeatures`): Recall, MediaFeatures, RemoteDesktopClient (слайс 17) — третий механизм удаления, см. раздел 9.6.
- `CustomScript{Format: cmd|ps1|reg|vbs, Content string}`.
- `PasswordExpirationSettings{Mode: default|never|custom (""=default), Days *int}`.
- `AccountLockoutSettings{Mode: default|disabled|custom (""=default), Threshold/WindowMinutes/DurationMinutes *int}`.
- `FileExplorerSettings{HiddenFiles: default|show_hidden|show_all, ShowFileExtensions, ClassicContextMenu, HideFolderTooltips, OpenToThisPC, ShowEndTaskInTaskbar bool}` — zero value ничего не меняет.
- `PersonalizationSettings{SystemTheme/AppsTheme: light|dark, AccentColor *string (RRGGBB), ShowAccentOnStartTaskbar/ShowAccentOnTitleBars/DisableTransparency bool, SolidColorWallpaper *string (RRGGBB)}` — только цвета; файл обоев и экран блокировки не входят (см. раздел 15, бэклог).
- `StickyKeysSettings{Mode: default|disabled|custom (""=default), Flags []StickyKeysFlag}` (слайс 19) — `Flags` только при `Mode=custom`, 6 значений (`profile.StickyKeysFlags`).
- `LockKeySettings{CapsLock/NumLock/ScrollLock: LockKeySetting{Initial: off|on, Behavior: toggle|ignore}}`, поле `Profile.LockKeys *LockKeySettings` (слайс 19) — `nil` = поведение Windows не трогается (как `SkipLockKeySettings` у сайта-эталона).
- `Profile.DesktopIcons map[DesktopIcon]bool` (слайс 20) — 13 значений (`profile.DesktopIcons`); `nil`/пустая карта = поведение Windows не трогается; ключ есть = явно show(`true`)/hide(`false`), ключа нет = соответствующий значок не трогается (частичная карта допустима).
- `Profile.StartFolders []StartFolder` (слайс 20) — 9 значений (`profile.StartFolders`); порядок в списке = порядок закрепления папок на Start; пустой список = не трогать (сознательное упрощение — «закрепить ровно ноль папок» этим полем не выразить, см. раздел 9.10).
- `Profile.InstallVMGuestTools []VMGuestTool` (слайс 21) — 4 значения (`profile.VMGuestTools`): VBoxGuestAdditions, VMwareTools, VirtIoGuestTools, ParallelsTools.
- `Profile.AppLockerPolicyXML *string` (слайс 21) — `nil` = не настраивать; сырой XML политики AppLocker, валидируется только на well-formedness (без XSD-схемы, см. раздел 9.11).
- `Profile.KeepSensitiveFiles bool` (слайс 22) — единственное поле в схеме, где zero value (`false`) вызывает действие, а не «ничего не менять»: `false` (по умолчанию) означает удаление чувствительных файлов после установки, как у сайта-эталона; `true` — не удалять. См. раздел 9.12.
- `Profile.UseNarrator bool` (слайс 22) — автозапуск Narrator на этапах windowsPE/specialize/при первом входе каждого будущего аккаунта.
- `Profile.ComputerNameScript *string` (слайс 22) — динамическое имя компьютера через PowerShell-скрипт; взаимоисключимо с `Profile.ComputerName` (ошибка валидации, если заданы оба).
- `Profile.ObscurePasswords bool` (слайс 22) — Base64-обфускация паролей аккаунтов в сгенерированном XML (не шифрование, см. раздел 9.12).
- `Profile.HidePowerShellWindows bool` (слайс 23) — глобальный переключатель: `-WindowStyle Hidden` вместо `Normal` для КАЖДОГО PowerShell-скрипта, генерируемого проектом (встроенные твики и пользовательские скрипты — все 13 мест, где строится ps1-команда, см. раздел 9.13). Не влияет на cmd/reg/vbs.

`profile.Default(name string) *Profile` — профиль по умолчанию, которым стартуют `profile init` без `--preset` и свежая TUI-сессия: `schema_version=1`, язык en-US/en-US/en-US, `Edition.Mode=interactive`, `Accounts=[]`, `FirstLogon.Mode=none`, `ExpressSettings.Mode=interactive`, всё остальное — нулевые значения.

Полей `updatedAt`, `disk configuration`, второго/третьего языка (multi-language) в v1 нет. Понадобилось — `CONTRACT GAP`.

---

## 5. API (заморожен)

`internal/profile`:

```go
func ValidateProfile(data []byte) ValidationResult
// ValidationResult{ Profile *Profile /* nil если есть ошибки */; Errors []string /* тексты на русском */ }

func Default(name string) *Profile
func LoadProfile(path string) (*Profile, error)
func SaveProfile(profile *Profile, path string) error
func ListProfiles(dir string) ([]string, error)

const ProfilesDir = "profiles"
```

`internal/xmlgen`:

```go
func BuildAnswerFile(p *profile.Profile) (string, error)
```

Правила:

- `ValidateProfile` сама делает `json.Unmarshal`; на вход — сырые байты файла, не структура.
- `BuildAnswerFile` НЕ валидирует профиль повторно — ожидает уже прошедший `ValidateProfile`. Единственная ошибка, которую он может вернуть — сбой сериализации XML.
- Ошибки валидации — только человекочитаемые русскоязычные строки, без кодов и без исключений/паник.
- `LoadProfile`/`SaveProfile` не валидируют — это отдельный слой (`cli/generate.go`, `cli/validate.go` вызывают `ValidateProfile` сами).

---

## 6. CLI-команды (заморожен)

Корень: `unattend-gen` (`internal/cli/root.go`, `NewRootCmd`/`Execute`).

| команда | что делает |
| --- | --- |
| `profile init <name> [--preset minimal\|single-user]` | создаёт `<name>.json` — из `profile.Default(name)` либо из встроенного пресета (`presets.Load`), с `Name` всегда переписанным на `<name>` |
| `profile list` | печатает пути `*.json` из `./profiles`, по одному на строку |
| `validate <profile.json>` | `ValidateProfile`; при ошибках печатает их в stderr и возвращает ненулевой код; при успехе печатает `профиль корректен` |
| `generate <profile.json> [-o путь]` | валидирует, затем `BuildAnswerFile`, пишет `autounattend.xml` рядом с профилем (или по `-o`) |
| `tui [profile.json]` | запускает bubbletea-приложение; если аргумент передан — стартует с `profile.LoadProfile` вместо `profile.Default` |

Правила:

- `generate` и `validate` печатают ошибки построчно в stderr и завершаются с ошибкой `profile validation failed`, а не паникой.
- Успешный вывод команд, читаемых скриптами (`profile init`, `generate`), — только путь к результату, ничего больше.

---

## 7. TUI: экраны и виджеты (заморожен)

Порядок экранов (`internal/tui/app.go`, `screens.ID`): **Welcome → Language → Accounts → Tweaks → Wifi → Apps → Personalization → Accessibility → Desktop → VisualEffects → Taskbar → Advanced → Scripts → Review**.

- Каждый экран — отдельный `tea.Model` в `internal/tui/screens/*.go`, общий `*profile.Profile` передаётся через `rebuildScreen` при каждой навигации, так экран всегда синхронизирован с последним состоянием.
- `screens.ScreenApps` — совмещённый экран: чекбоксы удаляемых приложений (`RemoveApps`), удаляемых DISM-компонентов (`RemoveFeatures`) и удаляемых legacy optional features (`RemoveOptionalFeatures`, слайс 17) — три разных механизма, одна таблица фокуса (`internal/tui/screens/apps.go`, `checkboxAt`).
- `screens.ScreenAccessibility` (слайс 19) — Sticky Keys (select + до 6 чекбоксов, только когда `mode=custom`) и Lock Keys (чекбокс «настраивать» + 6 select'ов, видны только если включён — `nil` `LockKeys` иначе, как на сайте-эталоне). Между Personalization и Scripts.
- `screens.ScreenDesktop` (слайс 20) — видимость значков рабочего стола (мастер-чекбокс «настраивать»: выключен = `nil` `DesktopIcons`, включён = все 13 значков получают явный чекбокс) + закреплённые папки на Start (`StartFolders`, простой список без мастер-чекбокса — пустой список сам по себе уже значит «не трогать», доп. переключатель не нужен). Между Accessibility и Scripts.
- `screens.ScreenVisualEffects` (слайс 25) — select режима (default/best_appearance/best_performance/custom) + 17 чекбоксов, видны только при `mode=custom`. Выбор режима custom всегда пишет явное значение для ВСЕХ 17 эффектов (не частичную карту) — тот же принцип «выбор режима подразумевает весь набор», что и у видимости значков рабочего стола на экране Desktop. Между Desktop и Advanced.
- `screens.ScreenTaskbar` (слайс 26) — 5 чекбоксов (`DisableWidgets`/`LeftTaskbar`/`HideTaskViewButton`/`DisableBingResults`/`ShowAllTrayIcons`), select `TaskbarSearch`, select+текстовое поле для `StartPins` (Win11 JSON) и `StartTiles` (Win10 XML) — текстовое поле видно только при `mode=custom`. Между VisualEffects и Advanced.
- `screens.ScreenAdvanced` (слайс 21, расширен в слайсе 22) — чекбоксы 4 наборов гостевых дополнений ВМ (`InstallVMGuestTools`) + `widgets.LabeledTextArea` для сырого AppLocker policy XML + `widgets.LabeledTextArea` для `ComputerNameScript` + 3 чекбокса (`KeepSensitiveFiles`, `UseNarrator`, `ObscurePasswords`). Пусто/не отмечено = `nil`/`false`, как везде в проекте. Между Taskbar и Scripts.
- `screens.ScreenWifi` (расширен в слайсе 22) — чекбокс «настроить Wi-Fi», затем чекбокс «raw XML вместо ручных полей» (`rawMode`): включён — показывает только `widgets.LabeledTextArea` для `WifiSettings.RawProfileXML`, ручные поля (SSID/auth/password/hidden) скрыты и не участвуют; выключен — старое поведение без изменений.
- `screens.ScreenTweaks` — самый нагруженный экран: express settings, 17 чекбоксов `SystemTweaks`, политика истечения пароля, политика блокировки аккаунта, настройки File Explorer. Число полей и индекс фокуса вычисляются динамически (условные блоки появляются только когда соответствующий Mode = custom).
- `screens.ScreenAccounts` — также несёт `Timezone` и `BypassOnlineAccountRequirement`, не только таблицу аккаунтов.
- Виджеты (`internal/tui/widgets/`), экраны не пишут свой ввод/таблицы напрямую:

| виджет | назначение |
| --- | --- |
| `LabeledInput` | однострочное текстовое поле с подписью и слотом под ошибку |
| `PasswordInput` | обёртка над `LabeledInput` с `EchoPassword` |
| `LabeledSelect` | компактный single-select для enum-полей |
| `LabeledTextArea` | многострочное поле (содержимое скриптов) |
| `Checkbox` | булево поле `[x]`/`[ ]` |
| `AccountsTable` | редактируемая таблица до 5 строк аккаунтов (`widgets.MaxAccounts = 5`) |
| `ConfirmBar` | нижняя строка подсказок горячих клавиш + сообщение об ошибке навигации |

Правило размещения новой функциональности: прежде чем добавлять поля на существующий экран, спросить — «это та же тема экрана, или нужен свой» (Personalization в слайсе 14 получил отдельный экран именно по этой причине, после того как Tweaks в слайсе 13 был признан перегруженным).

---

## 8. Валидация (заморожен)

Источник истины по допустимым значениям — `internal/profile/validate.go`. Совмещает структурные теги `go-playground/validator` (`required`, `max`, `oneof`, `dive`) с ручными проверками; каждая ручная проверка возвращает срез русскоязычных строк.

Основные правила (за точным текстом сообщений — код, здесь только смысл):

- `ui_language`/`locale`/`keyboard_layout` — формат BCP-47 (`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`).
- `edition.mode=generic_key` требует `edition`; `custom_key` требует `product_key` вида `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX`.
- `timezone`, если задан, не может быть пустой строкой (используй `null`), реальные Windows time zone ID не проверяются — их список слишком большой и version-dependent.
- `computer_name`, если задан: 1–15 символов, разрешённые символы, не начинается/заканчивается дефисом, не состоит только из цифр.
- Аккаунтов не больше 5; имя аккаунта непусто, ≤20 символов, без символов `"/\[]:;|=,+*?<>`; пароль `nil` (без пароля) или непустая строка, но не `""`.
- `first_logon.mode=first_created_account` требует хотя бы один аккаунт; `builtin_administrator` требует непустой `builtin_administrator_password`.
- Wi-Fi: SSID 1–32 символа; для `WPA2Personal`/`WPA3Personal` пароль обязателен и ≥8 символов.
- `remove_apps`/`remove_features` — только значения из `profile.RemovableApps`/`profile.RemovableFeatures`, неизвестные значения — ошибка.
- `default_user_scripts` не поддерживает формат `vbs`.
- `password_expiration.mode=custom` требует `days ≥ 1`; `account_lockout.mode=custom` требует все три числовых поля ≥1.
- `personalization.accent_color`/`solid_color_wallpaper`, если заданы — ровно 6 hex-цифр.

`ValidateProfile` не трогает диск и не строит XML — только декодирует JSON и проверяет.

---

## 9. Механизмы генерации XML

`xmlgen.BuildAnswerFile` собирает `<unattend>` с тремя `<settings pass="...">`: `windowsPE`, `specialize`, `oobeSystem` (пасс опускается целиком, если для него нет ни одного компонента). Внутри пасса компоненты — гетерогенный список `[]interface{}`, сериализуется вручную через `MarshalXML` (`settingsPass`), потому что `encoding/xml` не умеет сериализовать срез разнотипных структур под одним тегом `<component>` из коробки.

### 9.1 Общие соглашения

- У всех компонентов пространство имён `xmlns:wcm` и `wcm:action="add"` на элементах списков.
- Два разных конструктора для `International-Core`: `NewInternationalCoreWinPE` (имя компонента оканчивается на `-WinPE`) и `NewInternationalCoreSpecialize` (без суффикса) — это два разных компонента Windows, не один переиспользуемый.
- `AutoLogon` всегда `LogonCount=1`.
- `UserData.AcceptEula=true` всегда, когда задан продукт-ключ.
- Схема/имена элементов XML никогда не берутся по памяти для функционально значимых полей — только по проверенному источнику (Microsoft Learn или рабочий пример), с последующей проверкой реального сгенерированного вывода (см. раздел 12).

### 9.2 Два разных механизма для «выполнить один раз при первом входе»

- **FirstLogonCommands** (`Microsoft-Windows-Shell-Setup`, oobeSystem) — команда выполняется один раз, для аккаунта первого входа. Используется для: Wi-Fi (`wlan.go`), удаления приложений (`apps.go`), удаления компонентов (`features.go`), `FirstLogonScripts`. Все команды идут в один список с нарастающим `Order`: сначала Wi-Fi, затем apps, затем features, затем FirstLogon-скрипты.
- **Default-user hive mount** (`C:\Users\Default\NTUSER.DAT` монтируется как `HKU\DefaultUser`, правится, размонтируется) — правки применяются к **каждому** будущему аккаунту, включая ещё не созданные. Используется для: `DefaultUserScripts`, `UserOnceScripts` (через запись `RunOnce`-значения в этот куст), `FileExplorer`, `Personalization`. Реализовано через `Microsoft-Windows-Deployment` / `RunSynchronousCommand` в specialize.

### 9.3 SystemTweaks → RunSynchronousCommand (specialize, Microsoft-Windows-Deployment)

Каждый флаг `SystemTweaks` — одна или несколько команд в общем списке `RunSynchronousCommand` этого компонента (тот же компонент несёт также команды из `PasswordExpiration`/`AccountLockout`/скриптов System и DefaultUser):

Простые (одна reg.exe-команда): `DisableWindowsUpdate`, `DisableUAC`, `BypassWin11Requirements` (единственный tweak через `Microsoft-Windows-Setup/RunSynchronous` в windowsPE, а не Deployment/specialize — раньше в загрузке), `DisableSmartAppControl`, `DisableSmartScreen`, `DisableFastStartup`, `DisableSystemRestore`, `EnableLongPaths`, `EnableRemoteDesktop`, `AllowPowerShellScripts`, `DisableLastAccessTimestamp`, `PreventDeviceEncryption`, `DisableAutoSignOnLastUser`, `DisableWPBT`, `AuditProcessCreation`, `HideEdgeFirstRun`, `DisableEdgeStartupBoost`, `PreventDeviceApps` (слайс 16), `HardenSystemDriveACL` (слайс 18, `icacls.exe C:\ /remove:g "*S-1-5-11"`, снимает права Authenticated Users на корень системного диска), `DisableCoreIsolation` (слайс 21, 4 reg add на `DeviceGuard`/`HypervisorEnforcedCodeIntegrity` — отключает Memory Integrity/virtualization-based security, нужно некоторым гостевым ОС и старым драйверам).

`DeleteEdgeDesktopIcon` (слайс 23, побочная находка слайса 20) — 2 команды: простая (удаляет `C:\Users\Public\Desktop\Microsoft Edge.lnk`, доступен сразу после применения образа, без монтирования куста) + `DeleteEdgeDesktopIconUserOnceCommand` (RunOnce→живой-HKCU для будущих аккаунтов, тот же паттерн, что Desktop Icons в слайсе 20).

Составные (slice 16, `internal/xmlgen/components/optimizations.go`, механизмы сверены с исходником github.com/cschneegans/unattend-generator, `modifier/Optimizations.cs`, не по памяти):

- `TurnOffSystemSounds` — 3 команды: простая (BootAnimation/EditionOverrides, специализируется системно), `TurnOffSystemSoundsDefaultUserCommand` (мont default-user hive, чистит `AppEvents\Schemes` для будущих аккаунтов), `TurnOffSystemSoundsUserOnceCommand` (RunOnce-запись, ставит `.None` в живом `HKCU\AppEvents\Schemes` текущего аккаунта при первом входе).
- `DisableAppSuggestions` — 2 команды: простая (`CloudContent\DisableWindowsConsumerFeatures`), `DisableAppSuggestionsDefaultUserCommand` (обнуляет 17 значений `ContentDeliveryManager` в default-user hive).
- `DisablePointerPrecision` — 1 команда, только default-user hive (`Control Panel\Mouse`: MouseSpeed/MouseThreshold1/MouseThreshold2 = REG_SZ "0").
- `PreventAutomaticReboot` — 1 команда: 2 reg.exe (`WindowsUpdate\AU`: AUOptions=4, NoAutoRebootWithLoggedOnUsers=1) + `Register-ScheduledTask` с embedded XML задачи `MoveActiveHours` (сдвигает "активные часы" на текущее время каждые 4 часа, чтобы Windows не считала машину простаивающей).
- `DeleteHiddenJunctions` — 2 команды в разных pass'ах: `DeleteJunctionsFirstLogonCommand` (oobeSystem FirstLogonCommands, чистит reparse-point'ы вроде `C:\Documents and Settings` для аккаунта из установки) + `DeleteJunctionsUserOnceCommand` (specialize, тот же RunOnce-механизм что и `UserOnceScriptCommand`, для будущих аккаунтов).
- `DeleteWindowsOld` (слайс 18) — 1 команда, `FirstLogonCommands` (oobeSystem, не Deployment/specialize): `cmd.exe /c "rmdir C:\Windows.old"`. Точно как у сайта-эталона — без `/s`/`/q`, поэтому на непустой директории тихо ничего не делает; для чистой установки (без апгрейда) это безвредный no-op.
- `MakeEdgeUninstallable` (слайс 18) — 1 команда, `MakeEdgeUninstallableCommand`, specialize: запускает ps1, который правит `defaultState` политики Edge (`{1bca278a-5d11-4acf-ad2f-f9ab6d7f93a6}`) в `C:\Windows\System32\IntegratedServicesRegionPolicySet.json` с `disabled` на `enabled` — только так у Edge появляется реальная кнопка "Удалить" в "Приложениях и компонентах". Скрипт скопирован дословно из `resource/MakeEdgeUninstallable.ps1` эталона.

Точные реестровые пути/утилиты — в `internal/xmlgen/components/setup.go`/`optimizations.go` и комментариях к каждому tweak; при добавлении нового tweak путь проверяется заново, не копируется по аналогии вслепую.

### 9.4 OOBE hide-flags (производные, без отдельных полей профиля)

`Microsoft-Windows-Shell-Setup/OOBE` (oobeSystem): `HideEULAPage`, `HideOEMRegistrationScreen`, `HideOnlineAccountScreens`, `HideWirelessSetupInOOBE`, `HideLocalAccountScreen`, `NetworkLocation="Work"` включаются пачкой, когда `ExpressSettings.Mode != interactive`. Отдельно `HideOnlineAccountScreens` включается ещё и при `len(Accounts) > 0` вне зависимости от express settings. `BypassOnlineAccountRequirement=true` дополнительно ставит `HideOnlineAccountScreens` и добавляет команду записи `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\OOBE\BypassNRO=1` — задокументировано everywhere как best-effort, Microsoft неоднократно патчила обход этого ключа.

### 9.5 Скрипты (`scripts.go`)

Реальные команды запуска по формату (взяты с сайта-эталона, не придуманы):

| формат | команда |
| --- | --- |
| `cmd` | `cmd.exe /c "путь"` |
| `ps1` | `powershell.exe -WindowStyle Normal -ExecutionPolicy Unrestricted -NoProfile -File "путь"` |
| `reg` | `reg.exe import "путь"` |
| `vbs` | `cscript.exe //E:vbscript "путь"` |

Содержимое встраивается как base64 через `powershell -NoProfile -Command "..."`. Кавычки — единообразно двойные (не одинарные), потому что `UserOnce` дополнительно пишет .cmd-обёртку, а в `cmd.exe`/batch одинарная кавычка не работает как кавычка.

### 9.6 Приложения и компоненты

- `apps.go`: сопоставление `RemovableApp → подстроки DisplayName` (не `PackageFamilyName` — те меняются между версиями Windows), команда `Get-AppxProvisionedPackage | Remove-AppxProvisionedPackage`. Все селекторы (включая слайс 17) сверены с `resource/Bloatware.json` из github.com/cschneegans/unattend-generator, не по памяти. Осторожность: `AppPaint` использует паттерн `Microsoft.Paint` (не короткое `Paint`), чтобы не задеть `AppPaint3D` (`Paint3D`/старый пакет `Microsoft.MSPaint`) — покрыто тестом на коллизию.
- `apps.go`: `AppOneDrive` — единственное исключение из паттерн-механизма выше: OneDrive не Appx-пакет, а отдельный инсталлятор. `RemoveOneDriveFilesCommand` (specialize, без монтирования куста) удаляет ярлык и `OneDriveSetup.exe`/`OneDriveSetup.exe` (SysWOW64); `RemoveOneDriveDefaultUserCommand` (specialize, монтирует default-user hive) удаляет автозапуск из `...\CurrentVersion\Run`. Обе команды — отдельные `RunSynchronousCommand` в `Microsoft-Windows-Deployment`, не FirstLogonCommands.
- `features.go`: сопоставление `RemovableFeature → префикс(ы) DISM capability Name` (обратная ситуация: `Name` стабилен, суффикс `~~~lang~version` — нет), `Get-WindowsCapability -Online | Where Name -like "Prefix*" | Remove-WindowsCapability -Online`. Одна `RemovableFeature` может маппиться на несколько capability (`WindowsHello` → 3 значения; `Speech` — 2, `Language.Speech`+`Language.TextToSpeech`, исправлено в слайсе 17 — раньше было только первое).
- `optionalfeatures.go` (слайс 17) — третий, отдельный механизм удаления: `RemovableOptionalFeature → точное имя FeatureName` (не префикс — эти имена не несут version-суффикса), `Get-WindowsOptionalFeature -Online | Where FeatureName -eq Name | Disable-WindowsOptionalFeature -Online -Remove -NoRestart`. Та же FirstLogonCommands-семья (oobeSystem), что Wi-Fi/apps/features, третья по порядку команда. PowerShell 2.0 остаётся в бэклоге, хотя механизм для него теперь есть — не взят в слайс 17 сознательно (не входил в scope «недостающие приложения»).

### 9.7 Политика паролей/блокировки (`accountpolicy.go`)

Единственный механизм без PowerShell/base64-обёртки — статичная команда `net accounts`: `/maxpwage:UNLIMITED|<days>`, `/lockoutthreshold:0` (disabled) или `/lockoutthreshold:N /lockoutwindow:M /lockoutduration:D`. Дефолты Windows (не переопределяются при `Mode=default`): 42 дня истечения пароля, 10 попыток / 10 мин окно / 10 мин блокировка.

### 9.8 Персонализация (`personalization.go`)

`AccentColor` (вход — `RRGGBB`) упаковывается в DWORD `AABBGGRR` (альфа `FF`, байты RGB в обратном порядке) для `DWM\AccentColor`/`ColorizationColor` — единственная содержательная конвертация в этом компоненте, покрыта тестом с известной парой вход/выход. `SolidColorWallpaper` пишет `Control Panel\Colors\Background` как `"R G B"` decimal (тоже конвертация из hex) и очищает `Control Panel\Desktop\Wallpaper`, чтобы показывался цвет, а не картинка.

### 9.9 Sticky Keys и Lock Keys (`accessibility.go`, слайс 19)

Оба раздела отсутствовали в проекте полностью до слайса 19; механизмы сверены с `modifier/Optimizations.cs` эталона.

- **Sticky Keys** — значение `Flags` под `Control Panel\Accessibility\StickyKeys`: база `SKF_AVAILABLE(0x2) | SKF_CONFIRMHOTKEY(0x8)`, ИЛИ выбранные флаги (`HotKeyActive=0x4`, `Indicator=0x20`, `TriState=0x80`, `TwoKeysOff=0x100`, `AudibleFeedback=0x40`, `HotKeySound=0x10`). `Mode=disabled` — те же база-флаги без `HotKeyActive`, то есть отключается сама активация по 5×Shift, а не просто визуальные эффекты. Пишется в ДВА места: `HKU\DefaultUser` (для будущих аккаунтов, load/unload) и `HKU\.DEFAULT` (экран блокировки и любая сессия до загрузки пользовательского куста — этот куст всегда смонтирован, load/unload не нужен). `Mode=default`/`""` — команд не добавляется.
- **Lock Keys** — `nil` `Profile.LockKeys` = поведение Windows не трогается, `SkipLockKeySettings` у эталона. Если задан:
  - **Initial** (начальное состояние) — один decimal bitmask (Caps=1, Num=2, Scroll=4) в `InitialKeyboardIndicators` (REG_SZ) под `Control Panel\Keyboard`, тоже в оба места — `HKU\.DEFAULT` напрямую и `HKU\DefaultUser` через load/unload, одной комбинированной командой.
  - **Behavior=ignore** — бинарный `Scancode Map` (`HKLM\SYSTEM\CurrentControlSet\Control\Keyboard Layout`, REG_BINARY, вступает в силу после перезагрузки): 4 байта Version(0) + 4 байта Flags(0) + 4 байта little-endian Count(N+1) + по 4 байта на каждую отображаемую клавишу (`[0x00,0x00,scancode_lo,scancode_hi]` — target=0x0000 отключает клавишу) + 4 байта нулевой терминатор. Scancode'ы клавиш: CapsLock=`0x3A`, NumLock=`0x45`, ScrollLock=`0x46`. Команда добавляется только если хотя бы одна клавиша имеет `Behavior=ignore` — если у всех `toggle`, второй команды нет вообще.

### 9.10 Desktop Icons и Folders on Start (`startmenu.go`, слайс 20)

Оба раздела отсутствовали в проекте полностью до слайса 20. В отличие от 9.8/9.9 (которые правят статичный контент default-user hive или всегда-смонтированный `HKU\.DEFAULT`), эти два таргетят **живой** `HKCU` только что созданного аккаунта — единственный способ применить это к будущим аккаунтам через наш стек — тот же RunOnce-механизм, что уже есть в `scripts.go` для `UserOnceScripts` (монтируем default-user hive **только чтобы** прописать RunOnce-запись, сама команда выполняется потом, в сессии нового аккаунта, и пишет прямо в его `HKCU`).

- **Desktop Icons** — булева карта `DesktopIcons` (13 значений, GUID-ы взяты из `resource/DesktopIcon.json` эталона) пишется в ДВА подраздела `HideDesktopIcons` (`ClassicStartMenu` и `NewStartPanel` — Windows проверяет оба в зависимости от активного стиля меню Пуск) через RunOnce-обёрнутый `.cmd`-скрипт: `0` = показать, `1` = скрыть, только для ключей, реально присутствующих в карте. Скрипт завершается перезапуском `explorer.exe` (`taskkill /f /im explorer.exe && start explorer.exe`), иначе изменение не подхватится без выхода из сессии — как у эталона (`UserOnceScript.RestartExplorer()`).
- **Folders on Start** (закреплённые папки у кнопки питания, Win11) — `StartFolders []StartFolder` (9 значений, 16-байтовые GUID взяты из `resource/StartFolder.json`, декодированы из base64 в hex-константы в коде) конкатенируются В ПОРЯДКЕ СПИСКА и пишутся одним REG_BINARY значением `VisiblePlaces` под `...\CurrentVersion\Start`, тоже через RunOnce/`.cmd`, без PowerShell (тот же hex-подход, что и Scancode Map в 9.9). Пустой список = не трогать (сознательное упрощение, задокументировано в исходном коде `StartFoldersUserOnceCommand`): выразить «закрепить ровно ноль папок» этим полем нельзя, только «оставить дефолт Windows».

### 9.11 VM guest tools и AppLocker (`vmapplocker.go`, слайс 21)

- **VM guest tools** — `InstallVMGuestTools []VMGuestTool` (4 значения). Каждый инструмент — отдельная `FirstLogonCommands`-команда (не объединены в одну, как Wi-Fi/apps/features — если ISO одного гостевого пакета не примонтирован, это не должно мешать остальным). Содержимое всех 4 ps1-скриптов скопировано дословно из `resource/*.ps1` эталона: каждый перебирает буквы дисков D-Z в поисках своего инсталлятора на примонтированном ISO и молча завершается с сообщением, если ISO не подключён — это единственный способ найти гостевые дополнения, т.к. буква диска CD-привода не предсказуема заранее. Сознательное упрощение: у эталона VBox/VMware/Parallels (но не VirtIO — тот всегда идёт через FirstLogon) запускаются в specialize вместо FirstLogon, если у пользователя отключён Defender (`IsDefenderDisabled`) — в проекте нет эквивалентного сигнала «Defender отключён», поэтому все 4 инструмента всегда идут через FirstLogonCommands (это и есть собственный fallback-путь эталона, не отклонение от него).
- **AppLocker** — `AppLockerPolicyXML *string`, `nil` = не настраивать. Заданный XML встраивается как файл (как пользовательские скрипты) и применяется в specialize: сперва `Get-Service AppIDSvc | Set-Service -StartupType Automatic` + `Start-Service` (без этой службы AppLocker молча не работает), затем `Set-AppLockerPolicy -XmlPolicy`. Валидация (`profile.ValidateProfile`) проверяет только well-formedness XML через `encoding/xml` — полной проверки по XSD-схеме AppLocker (как у эталона, `Util.ValidateAgainstSchema`) в проекте нет: в стандартной библиотеке Go нет XSD-валидатора, а тащить его отдельной зависимостью ради одного поля сочли неоправданным для этого слайса. Синтаксически кривая, но «не-XML-невалидная» политика будет обнаружена только на целевой машине при ошибке `Set-AppLockerPolicy`.

### 9.12 Keep sensitive files, Narrator, динамическое имя компьютера, обфускация паролей (`misc.go`, `wlan.go`, слайс 22)

- **Keep sensitive files** (`internal/xmlgen/components/misc.go`) — единственное поле в схеме, ведущее себя не по конвенции «zero value ничего не меняет»: `KeepSensitiveFiles=false` (значение по умолчанию) удаляет `C:\Windows\Panther\unattend.xml`/`unattend-original.xml` (куда Windows Setup копирует answer file для собственного логирования — это и есть основной риск утечки паролей) через `FirstLogonCommands`; `true` — оставляет файлы. Это сознательное отступление от общей конвенции проекта, сделанное намеренно: молчаливое оставление на диске файла с открытым текстом пароля — худший дефолт, чем нарушение конвенции. В `baseProfile()` (тестовая фикстура `internal/xmlgen/builder_language_test.go`) поле явно выставлено в `true`, чтобы не плодить лишнюю команду во всех существующих «пустой профиль = 0 команд» тестах прошлых слайсов — сам дефолт (`false` = удалять) не тронут. В отличие от эталона, нет `C:\Windows\Setup\Scripts\Wifi.xml` для удаления (наш Wi-Fi пишет профиль в `%TEMP%\wifi-profile.xml`, см. `wlan.go`) — чистится этот путь вместо него.
- **Narrator** (`UseNarrator bool`) — 3 части: windowsPE (`Microsoft-Windows-Setup/RunSynchronous`, `NewSetup` получил параметр), specialize (запуск + `HKLM\...\Accessibility\Configuration=narrator`), UserOnce для будущих аккаунтов (тот же RunOnce→живой-HKCU паттерн, что и Desktop Icons в слайсе 20).
- **Динамическое имя компьютера** (`ComputerNameScript *string`) — взаимоисключимо с `ComputerName` (ошибка валидации при обоих). `ComputerName` в XML ставится в заглушку `"TEMPNAME"`. Механизм: пользовательский скрипт пишет новое имя в `ComputerName.txt`, затем стартует СКРЫТЫЙ фоновый PowerShell-процесс (`SetComputerName.ps1`, скопирован дословно из эталона — пути внутри уже совпадают с `scriptsDir` проекта), который каждые 50мс переприменяет имя к реестру бесконечно — это обходит особенность Windows, которая сама переписывает эти ключи во время specialize уже ПОСЛЕ старта скрипта.
- **Base64-обфускация паролей** (`ObscurePasswords bool`) — `newPasswordElement(raw, element, obscure)` в `shellsetup.go`: `Base64(UTF16LE(raw+element))`, где `element` — имя XML-элемента (`"Password"` или `"AdministratorPassword"`) как соль, `PlainText=false`. Стандартная конвенция Windows unattend (Microsoft Learn), НЕ шифрование — соль фиксирована и общеизвестна, обфусцированное значение восстанавливается тем же способом, что и открытое; единственная практическая польза — пароль не виден grep'ом по сырому XML. Реализовано на стандартной библиотеке (`unicode/utf16`+`encoding/binary`), без новых зависимостей.
- **Raw WLAN profile XML** (`WifiSettings.RawProfileXML *string`) — если задан, используется как есть вместо построения профиля из `SSID`/`Authentication`/`Password`/`ConnectHidden` (те становятся необязательными на уровне struct-тегов, проверка перенесена в `validateWifi`). Экспортируется пользователем через `netsh wlan export profile key=clear`.

### 9.13 DeleteEdgeDesktopIcon и глобальный Hide PowerShell windows (слайс 23)

- **DeleteEdgeDesktopIcon** — см. раздел 9.3, `optimizations.go`.
- **HidePowerShellWindows** (`Profile.HidePowerShellWindows bool`) — довёл до конца 6-й пункт группы E, отложенный в слайсе 22. `internal/xmlgen/components/scripts.go`: `invokeCommand(format, path, hidden bool)` — теперь для `ScriptPs1` выбирает `-WindowStyle Hidden`/`Normal`; `cmd`/`reg`/`vbs`-инвокации не затронуты (у сайта-эталона видимыми во время установки бывают только PowerShell-окна). Параметр `hidePowerShellWindows` протащен через все 13 мест вызова `invokeCommand` в 5 файлах (`scripts.go` — `SystemScriptsCommands`, `FirstLogonScriptsCommands`, `DefaultUserScriptCommand`, `UserOnceScriptCommand`; `apps.go` — `RemoveOneDriveFilesCommand`; `misc.go` — `KeepSensitiveFilesFirstLogonCommand`; `optimizations.go` — `DeleteJunctionsFirstLogonCommand`, `DeleteJunctionsUserOnceCommand`, `TurnOffSystemSoundsDefaultUserCommand`, `TurnOffSystemSoundsUserOnceCommand`, `MakeEdgeUninstallableCommand`, `DeleteEdgeDesktopIconUserOnceCommand`; `vmapplocker.go` — `VMGuestToolsFirstLogonCommands`) и далее в сигнатуры `NewDeployment`/`NewShellSetupOOBE`. Затрагивает как встроенные твики-скрипты, так и пользовательские `SystemScripts`/`FirstLogonScripts`/`DefaultUserScripts`/`UserOnceScripts` — единый переключатель на все PowerShell-инвокации проекта, как и задумано у эталона.

Группа E бэклога закрыта полностью (6/6).

### 9.14 Group D — достижимая часть (слайс 24): активационный ключ, BIOS/UEFI ключ, архитектура процессора

**Важная находка при исследовании этого слайса, меняющая прошлый аудит:** «Allow Windows 11 to be installed without internet connection» из группы D — это НЕ отдельный, нереализованный механизм. У эталона (`modifier/Bypass.cs`, `Configuration.BypassNetworkCheck`) он пишет ровно тот же `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\OOBE\BypassNRO`, что наш `BypassOnlineAccountRequirement` уже делает с самого слайса 7. Пункт был задвоен в первоначальном аудите — исправлено здесь, не отдельная задача.

**Ещё одна находка, сузившая скоуп группы D до трёх пунктов.** Изучение `modifier/Disk.cs` показало: `Disable 8.3 filenames`, `Disable Defender in PE`, паузы перед разметкой/перезагрузкой, `compact`-режим, `skip integrity check`, выбор образа по имени/индексу и `$OEM$` distribution share — ВСЕ они являются частью ОДНОГО механизма эталона (`GeneratePESettings`/`CustomPESettings`): полной замены windowsPE-этапа собственным `.cmd`-скриптом, который сам делает `diskpart`-разметку, `dism /Apply-Image` и `bcdboot`. Это прямо конфликтует с явным исключением диск-партиционирования из скоупа проекта (раздел 1, «Windows Setup всегда спрашивает интерактивно, куда ставить»). Ни один из этих под-пунктов не достижим независимо от остальных без сначала реализации полной замены PE-этапа — что осталось явно исключено. Это НЕ рекатегоризация одного пункта, а закрытие сразу семи пунктов бэклога группы D как «неприменимо при данном архитектурном решении», а не «сделать позже».

Реально достижимые и реализованные в этом слайсе:

- **Activation Key** (`Profile.ActivationKey *string`) — `Microsoft-Windows-Shell-Setup/ProductKey` (specialize) — отдельный XML-элемент, независимый от `UserData/ProductKey` (windowsPE), который используется только для выбора образа/показа UI на этапе установки. `components.ResolveActivationKey(edition, activationKey)`: явный `ActivationKey` побеждает; иначе, если `Edition.Mode=custom_key`, переиспользуется тот же ключ (совпадает с прежним поведением, теперь оформлено явной функцией); иначе — пусто, элемент не пишется (`ShellSetupSpecialize.ProductKey` с `omitempty`).
- **BIOS/UEFI stored key** (`EditionSettings.Mode=firmware`) — `UserData` без `ProductKey`-элемента вообще, `WillShowUI=Never`, `AcceptEula=true`. `UserData.ProductKey` стал указателем (`*ProductKey`, `omitempty`) для этого случая — раньше был обязательным полем.
- **Processor architecture** (`Profile.ProcessorArchitecture`) — `newStandardAttrs(arch)` теперь принимает архитектуру вместо жёстко зашитого `"amd64"`; протащено через 6 конструкторов, несущих `standardAttrs` (`NewInternationalCoreWinPE`/`Specialize`, `NewSetup`, `NewShellSetupSpecialize`, `NewDeployment`, `NewShellSetupOOBE`). Мульти-архитектура (несколько значений в одном XML, как у эталона) сознательно не реализована — см. раздел 4.

### 9.15 Visual Effects (`visualeffects.go`, слайс 25)

Раздел сайта вне исходных 6 групп аудита, реализован полностью. Механизм сверен с `modifier/Optimizations.cs` эталона.

- 4 режима: `default` (ничего не менять), `best_appearance` (все 17 эффектов включены), `best_performance` (все выключены), `custom` (карта `map[VisualEffect]bool`, только перечисленные эффекты).
- Две команды, обе — `Microsoft-Windows-Deployment` (specialize): `VisualEffectsSpecializeCommand` пишет `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects\{Имя}\DefaultValue` (0/1) для каждого перечисленного эффекта — это шаблон, который Windows копирует в настройки НОВОГО аккаунта при первом создании его профиля, поэтому действует на все будущие аккаунты, не только на созданный при установке. `VisualEffectsUserOnceCommand` — тот же RunOnce→живой-HKCU паттерн, что и Desktop Icons (слайс 20): ставит `HKCU\...\Explorer\VisualEffects\VisualFXSetting` = 1 (best appearance) / 2 (best performance) / 3 (custom) — значение, которое диалог «Параметры быстродействия» показывает как выбранный пресет для ТЕКУЩЕГО аккаунта в момент первого входа.
- 17 значений `VisualEffect` — имена ключей реестра совпадают с именами констант дословно (`ControlAnimations`, `AnimateMinMax`, ..., `DropShadow`), взяты из `enum Effect` эталона без изменений.

Побочная находка при работе над этим слайсом: `screens.ScreenAdvanced` отсутствовал как `case` в `rebuildScreen` (`internal/tui/app.go`) с самого слайса 21 — переход на этот экран (`Ctrl+N` с Desktop) всегда ошибочно показывал Review вместо Advanced. Исправлено попутно при добавлении `ScreenVisualEffects` в тот же switch.

### 9.16 Start menu/taskbar — достижимая часть (`taskbar.go`, слайс 26)

Раздел сайта вне исходных 6 групп аудита. Реализована достижимая часть (8 из 9 под-пунктов); самый сложный под-пункт (кастомные закреплённые иконки таскбара через locked Start layout XML + scheduled-task-based unlock flow) сознательно не взят — см. бэклог. Механизмы сверены с `modifier/Optimizations.cs` эталона и `resource/ShowAllTrayIcons.*`.

- Простые (одна reg.exe-команда, `SystemTweaks`): `DisableWidgets` (specialize, машинная политика). `LeftTaskbar`, `HideTaskViewButton`, `DisableBingResults` — каждый через свой независимый цикл монтирования default-user hive (per-account значения `Explorer\Advanced`/`Explorer` policy, не общесистемная политика).
- `ShowAllTrayIcons` (`SystemTweaks`) — ветвление по билду ОС (тот же подход, что у эталона: Win10 и Win11 требуют разных механизмов): на Win10 — прямое значение реестра `EnableAutoTray=0` в default-user hive; на Win11 — scheduled task, запускающийся при каждом логоне и помечающий каждую иконку трея `IsPromoted=1` (XML задачи скопирован дословно из `resource/ShowAllTrayIcons.xml`).
- `Profile.TaskbarSearch` — тот же RunOnce→живой-HKCU паттерн, что Desktop Icons/Visual Effects: ставит `HKCU\...\Search\SearchboxTaskbarMode` (0=hide/1=icon/2=box/3=label) + перезапускает Explorer, при первом входе КАЖДОГО будущего аккаунта.
- `Profile.StartPins` (только Win11) — специализированная политика `HKLM\SOFTWARE\Microsoft\PolicyManager\current\device\Start\ConfigureStartPins` = сырой JSON (`{"pinnedList":[...]}`), проверка на билд ОС внутри самого PS-скрипта (`OSVersion.Version.Build -lt 20000 → return`). JSON встраивается в одинарные PowerShell-кавычки через `psSingleQuoteEscape` (удвоение `'`), не через `escapeForOuterCommand` (тот экранирует двойные кавычки для внешней обёртки, здесь нужен другой уровень экранирования).
- `Profile.StartTiles` (только Win10) — прямая запись файла `C:\Users\Default\AppData\Local\Microsoft\Windows\Shell\LayoutModification.xml`, БЕЗ монтирования куста реестра (это обычный файл, путь существует сразу после применения образа).

---

## 10. Сборка и проверки (заморожен)

`Makefile`:

```
fmt        gofmt -w .
fmt-check  test -z "$(gofmt -l .)"
vet        go vet ./...
lint       golangci-lint run   (только govet, staticcheck, unused, errcheck — .golangci.yml)
test       go test ./... -race
gate       fmt-check vet lint test
```

`make gate` обязателен перед завершением любого слайса. CI (GitHub Actions) гоняет то же самое на push в `main`/`dev` и на PR.

---

## 11. Git-процесс (заморожен)

- Рабочая ветка — `dev`, после готового слайса/фикса — в `main` (fast-forward где возможно, иначе честный merge).
- Версионные теги `vX.Y.Z`, по одному на слайс. На момент этой версии файла: `v0.0.0`…`v0.15.0`, CI зелёный на каждом.
- `go.sum` нельзя сгенерировать локально в песочнице (нет доступа к `proxy.golang.org`) — гейт-джоба в CI гоняет `go mod tidy` и коммитит `go.sum` обратно с `[skip ci]`. Из-за этого `git push` иногда отклоняется как «fetch first» просто потому, что этот бот-коммит успел прилететь на remote — это ожидаемо, не признак внешнего вмешательства: `git fetch` + rebase/merge + push ещё раз.
- Исключение из предыдущего пункта: если `tech.md` пропал из репозитория без объяснения — это был реальный внешний (человеческий) коммит, который нужно было сохранить, а не перезаписать бездумно.
- GitHub PAT пользователь присылает в чат каждый раз заново (токены не секретны, если засветились в переписке — стоит их ротировать после использования).

---

## 12. Дисциплина точности схемы (заморожен)

Ни один XML-элемент/атрибут, значимый для функциональности, не пишется по памяти. Порядок: (1) проверить по Microsoft Learn или по рабочему примеру, (2) реализовать, (3) проверить фактический сгенерированный вывод — собрать бинарник в CI и забрать файл через `raw.githubusercontent.com`. GitHub Checks-аннотации не годятся для этой проверки — они портят текст, похожий на XML-теги.

Что ещё не проверено реальным Windows-окружением (песочница его не даёт): фактическая установка сгенерированного `autounattend.xml` через USB/VM. Особо хрупкое место для первой такой проверки — квотинг в `UserOnce`/default-user-hive командах (слайс 10).

---

## 13. CONTRACT GAP

Если для задачи нужен тип/поле/API/XML-элемент, которого нет в этом файле — работа останавливается, в ответ идёт блок:

```
CONTRACT GAP
Нужно: <что именно>
Где не хватает: <раздел файла>
Предлагаемое решение: <опционально>
```

Код с придуманным контрактом не пишется. Обновляет контракт только владелец проекта — версия файла поднимается сверху.

---

## 14. Ревью после каждого слайса

Отдельным заходом, после того как слайс готов и `make gate` зелёный. Задача захода — искать проблемы, а не хвалить написанное.

1. `internal/xmlgen` не импортирует `cobra`/`bubbletea` и не читает диск.
2. CLI/TUI не собирают XML напрямую, только через `xmlgen.BuildAnswerFile`.
3. Имена JSON-полей совпадают с разделом 4. Ни одного придуманного `snake_case`-имени мимо `schema.go`.
4. Ни один новый XML-элемент не взят по памяти без проверки (раздел 12).
5. Тесты проверяют реальные критерии приёмки слайса, а не повторяют реализацию; на каждый отказ валидации есть тест на отказ.
6. TUI-экраны используют виджеты из `internal/tui/widgets`, самописных полей/таблиц в экранах нет.
7. Мёртвого кода нет: неиспользуемые экспорты, поля, экраны.
8. Файлов и абстракций сверх раздела 3 не появилось без функциональной причины.
9. Если tweak/скрипт/реестровый путь — проверь: это правда affecting default-user hive корректно (для полей, которые должны применяться на будущие аккаунты), а не только на текущего пользователя.

Находки правятся в том же заходе, потом гейт прогоняется заново.

---

## 15. История и бэклог

### Сделано (v0.0.0 → v0.15.0, слайсы 0–15)

Кратко, по темам (детали — в git-истории и тегах):

- **0–6**: каркас; язык/раскладка/издание; имя компьютера + до 5 локальных аккаунтов + автовход; CLI `generate` end-to-end; express settings + 3 базовых tweak'а (`DisableWindowsUpdate`, `DisableUAC`, `BypassWin11Requirements`); TUI (6 экранов) + 2 пресета (`minimal`, `single-user`); Wi-Fi через `netsh wlan add profile`.
- **7**: часовой пояс, OOBE hide-flags, `BypassOnlineAccountRequirement`/BypassNRO.
- **8**: `SystemTweaks` расширен до 17 полей; экран Tweaks переведён на табличный (не по-полю-switch) дизайн.
- **9**: удаление приложений (`RemoveApps`, 32 значения), новый экран Apps.
- **10**: пользовательские скрипты — 4 категории (System/DefaultUser/FirstLogon/UserOnce), новый экран Scripts, `LabeledTextArea`.
- **11**: удаление компонентов Windows через DISM (`RemoveFeatures`, 7 значений) — добавлено на существующий экран Apps, не отдельным экраном.
- **12**: истечение пароля и политика блокировки аккаунта (`net accounts`) — добавлено на экран Tweaks.
- **13**: тонкая настройка File Explorer через default-user hive — добавлено на экран Tweaks (признано перегруженным).
- **14**: персонализация цветов (тема, акцент, прозрачность) — получила отдельный новый экран Personalization.
- **15**: сплошной цвет обоев рабочего стола — добавлено в существующий `PersonalizationSettings`/экран Personalization, не новый тип.

- **16**: доработка System tweaks (tech.md backlog group B, 6 полей) — `DeleteHiddenJunctions`, `PreventAutomaticReboot`, `TurnOffSystemSounds`, `DisableAppSuggestions`, `DisablePointerPrecision`, `PreventDeviceApps`. Механизмы сверены с исходником github.com/cschneegans/unattend-generator (`modifier/Optimizations.cs`), не по памяти. Новый файл `internal/xmlgen/components/optimizations.go`. `SystemTweaks` 17→23 полей, экран Tweaks обновлён (`tweaksCount` 17→23, индексация выведена из константы). `Harden ACLs`, `Make Edge uninstallable`, `Delete empty C:\Windows.old` из Group B сознательно НЕ взяты в этот слайс (другие/более рискованные механизмы) — остаются в бэклоге.

- **17**: расширение Remove bloatware (tech.md backlog group A, 14 из ~18 позиций) + новый механизм удаления. 10 новых `RemovableApp` через существующий Appx-механизм (Bing Search, Dev Home, Game Assist, Microsoft Store, Notepad modern, Outlook for Windows, Paint, Wallet, Windows Media Player modern, Windows Terminal), 4 новых `RemovableFeature` через существующий DISM-capability-механизм (Windows Hello, Math Input Panel, OneSync, Steps Recorder), `AppOneDrive` через новый custom-механизм (файлы + default-user-hive run-key, не Appx). Новый файл `internal/xmlgen/components/optionalfeatures.go` + тип `RemovableOptionalFeature` — третий механизм удаления (`Disable-WindowsOptionalFeature`), закрывает Recall/MediaFeatures/RemoteDesktopClient. Побочно исправлена неточность в `FeatureSpeech` (второй capability-селектор). Экран Apps получил третью группу чекбоксов. Media Features и Recall изначально требовали нового механизма — реализован в этом же слайсе, не отложен. PowerShell 2.0 (тот же новый механизм) в бэклоге сознательно не тронут.

- **18**: остаток Group B (`Harden ACLs`, `Make Edge uninstallable`, `Delete Windows.old`) — все 3 механизма сверены с `modifier/Optimizations.cs` эталона. `HardenSystemDriveACL` — простая specialize-команда. `DeleteWindowsOld` — FirstLogonCommands (oobeSystem), не specialize, как остальные tweaks — важное отличие, учтено в сигнатуре `NewShellSetupOOBE`. `MakeEdgeUninstallable` — specialize, ps1-скрипт скопирован дословно из `resource/MakeEdgeUninstallable.ps1`. `SystemTweaks` 23→26. Group B бэклога закрыта полностью.

- **19**: Sticky Keys + Lock Keys (tech.md backlog group C, 2 из 9 пунктов). Новый файл `internal/xmlgen/components/accessibility.go`, новый экран `screens.ScreenAccessibility` (между Personalization и Scripts). `StickyKeysSettings` (Mode + Flags) и `*LockKeySettings` (nil = не трогать) — оба механизма пишут одновременно в `HKU\DefaultUser` (для будущих аккаунтов) и `HKU\.DEFAULT` (для текущей сессии/экрана блокировки, без load/unload — этот куст всегда смонтирован). Lock Keys' `Behavior=ignore` строит бинарный Scancode Map с нуля (Go-реализация формата, побайтово сверена с C#-источником, покрыта тестом на конкретные hex-байты).

- **20**: Desktop Icons + Folders on Start (tech.md backlog group C, 4 из 9 пунктов, суммарно с 19). Новый файл `internal/xmlgen/components/startmenu.go`, новый экран `screens.ScreenDesktop` (между Accessibility и Scripts). Оба механизма используют RunOnce (как `UserOnceScripts` в `scripts.go`), поскольку таргетят живой `HKCU` нового аккаунта, а не default-user hive. `Profile.DesktopIcons map[DesktopIcon]bool` (13 значений) и `Profile.StartFolders []StartFolder` (9 значений, GUID-байты из `resource/*.json` эталона, decode-once в hex-константы). Sознательное упрощение задокументировано: `StartFolders` не может выразить «закрепить ровно ноль папок».
- **21**: AppLocker + VM guest tools + `DisableCoreIsolation` (tech.md backlog group C, +3/9, 9/9 закрыта полностью). Новый файл `internal/xmlgen/components/vmapplocker.go`, новый экран `screens.ScreenAdvanced` (между Desktop и Scripts). `InstallVMGuestTools []VMGuestTool` (4 ps1-скрипта дословно из эталона, каждый — отдельная FirstLogonCommand), `AppLockerPolicyXML *string` (raw XML, only well-formedness validated, не XSD), `SystemTweaks.DisableCoreIsolation` (27-й твик). Group C бэклога закрыта полностью.
- **22**: 5 из 6 пунктов Group E — `KeepSensitiveFiles` (единственное поле-исключение из конвенции «zero value ничего не меняет», задокументировано и защищено в тестовой фикстуре), `UseNarrator` (windowsPE+specialize+UserOnce), `ComputerNameScript` (взаимоисключимо с `ComputerName`, фоновый процесс из `SetComputerName.ps1` дословно), `ObscurePasswords` (Base64/UTF-16LE, stdlib, новый helper `newPasswordElement` заменил все 4 места конструирования пароля), `WifiSettings.RawProfileXML`. Экран Advanced расширен (2 текстовых поля + 3 чекбокса), экран Wifi получил переключатель raw-режима. Новый файл `internal/xmlgen/components/misc.go`. 6-й пункт (глобальный Hide PowerShell windows) сознательно не взят — инвазивный рефакторинг сигнатур, отдельный пункт бэклога.
- **23**: `DeleteEdgeDesktopIcon` (28-й SystemTweak, побочная находка слайса 20) + глобальный `HidePowerShellWindows` (последний пункт группы E, ЗАКРЫТА ПОЛНОСТЬЮ 6/6). `invokeCommand` получил параметр `hidden bool`, протащен через 13 функций в 5 файлах + сигнатуры `NewDeployment`/`NewShellSetupOOBE`. Затрагивает встроенные твики и пользовательские скрипты одинаково.
- **24**: достижимая часть группы D — `ActivationKey`, `EditionModeFirmware`, `ProcessorArchitecture`. По ходу исследования выяснено: «без интернета» дублировал уже реализованный `BypassOnlineAccountRequirement` (не отдельная задача), а 7 других под-пунктов группы D (8.3-имена, Defender-в-PE, паузы, compact, skip integrity, выбор образа, `$OEM$`) — все части одного механизма замены PE-этапа своим `.cmd`-скриптом с diskpart/dism, конфликтующего с исключением диск-партиционирования (раздел 1) — закрыты как «неприменимо», не как «сделать позже». `UserData.ProductKey` стал указателем (`omitempty`) для режима firmware. `newStandardAttrs` принимает архитектуру, протащено через 6 конструкторов.
- **25**: Visual Effects (раздел сайта вне исходных 6 групп аудита) — новый файл `internal/xmlgen/components/visualeffects.go`, новый экран `screens.ScreenVisualEffects` (между Desktop и Advanced). 4 режима, 17 индивидуальных эффектов, тот же RunOnce→живой-HKCU паттерн, что Desktop Icons. Попутно найден и исправлен баг с слайса 21: `ScreenAdvanced` отсутствовал в `rebuildScreen`, переход туда показывал Review вместо Advanced.
- **26**: Start menu/taskbar, достижимая часть (8 из 9 под-пунктов) — новый файл `internal/xmlgen/components/taskbar.go`, новый экран `screens.ScreenTaskbar` (между VisualEffects и Advanced). 5 новых `SystemTweaks` (`DisableWidgets`, `LeftTaskbar`, `HideTaskViewButton`, `DisableBingResults`, `ShowAllTrayIcons` — последний с ветвлением по билду ОС и scheduled-task на Win11), `Profile.TaskbarSearch`, `Profile.StartPins` (Win11 JSON policy), `Profile.StartTiles` (Win10 XML-файл в Default-профиль). Кастомные закреплённые иконки таскбара (locked Start layout + scheduled-task-unlock) сознательно не взяты — остаются в бэклоге как единственный нереализованный под-пункт.

### Бэклог — полная сверка с schneegans.de (аудит 2026-09-03)

Источник: [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/), commit `88d81f0`. Сверка сделана постранично, раздел за разделом сайта, против фактического кода (не только против старого текста этого файла — там нашлись расхождения, см. ниже).

Каждый пункт при взятии в работу — отдельный слайс: сначала проверка реального механизма (раздел 12), затем реализация, затем поднятие версии этого файла с добавлением контракта в раздел 4/9 и пометкой пункта здесь как сделанного (переносится в раздел «Сделано» выше).

Группы упорядочены по приоритету (первая — предлагаемая следующая, но порядок не жёсткий, решает владелец).

**Группа A — расширение Remove bloatware — ЗАКРЫТО в слайсе 17, кроме Media Features/Recall (см. ниже).**
Реализовано 14 из ~18 недостающих позиций: 10 через существующий Appx-механизм (`apps.go`: Bing Search, Dev Home, Game Assist, Microsoft Store, Notepad (modern), Outlook for Windows, Paint, Wallet, Windows Media Player (modern), Windows Terminal), 4 через существующий DISM-capability-механизм (`features.go`: Facial recognition/Windows Hello — 3 capability, Math Input Panel, OneSync, Steps Recorder), 1 через новый custom-механизм (OneDrive — не Appx-пакет, а файлы + default-user-hive registry run-key, см. `RemoveOneDrive*Command` в `apps.go`). Точные селекторы взяты из `resource/Bloatware.json` исходника-эталона (github.com/cschneegans/unattend-generator), не по памяти. Попутно исправлена неточность в существующем `FeatureSpeech` — у сайта 2 capability-селектора (`Language.Speech` + `Language.TextToSpeech`), в проекте был только первый.

**Media Features и Recall — вынесены из группы A, требовали НОВОГО механизма.** У сайта они (и Remote Desktop Client — уже был неучтён как отдельный гэп) используют `Get-WindowsOptionalFeature`/`Disable-WindowsOptionalFeature` — третий механизм удаления, отдельный от Appx и DISM capability, который раньше в проекте не был реализован (упоминался в бэклоге только применительно к PowerShell 2.0). Слайс 17 реализовал и его: новый файл `internal/xmlgen/components/optionalfeatures.go`, новый тип `profile.RemovableOptionalFeature` + поле `Profile.RemoveOptionalFeatures`, все 3 позиции (Recall, MediaFeatures, RemoteDesktopClient) закрыты этим же слайсом.

**Группа B — оставшиеся System tweaks — ЗАКРЫТА полностью в слайсе 18.**

**Группа C — целые отсутствующие разделы сайта (новая функциональность, скорее всего отдельные экраны TUI). ЗАКРЫТА ПОЛНОСТЬЮ в слайсах 19–21 (Lock keys, Sticky keys, Desktop icons, Folders on Start, VM guest tools, VM host core isolation, AppLocker). Start menu/taskbar был заявлен отдельно от исходного 9-пунктового списка группы, остаётся открытым ниже; Visual effects (тоже был заявлен отдельно) закрыт в слайсе 25.**
- **Start menu and taskbar** — ЗАКРЫТ почти полностью в слайсе 26 (8 из 9 под-пунктов: режим поля поиска, отключение виджетов, left-align таскбара, скрытие Task View, «always show tray icons», отключение Bing-результатов, плитки Start Win10, pins Win11 — все сделаны). Остался один под-пункт:
  - **Конфигурация закреплённых иконок таскбара через XML** (`TaskbarIcons` у эталона) — сознательно не взят в слайс 26: требует locked Start layout XML + `LockedStartLayout` реестровых значений + scheduled-task-based unlock flow (`UnlockStartLayout`) + отдельный `.vbs`-скрипт — значительно сложнее остальных 8 под-пунктов вместе взятых, отдельный кандидат на будущий слайс.
- ~~Visual effects~~ — закрыто в слайсе 25 (`Profile.VisualEffects`, раздел 9.15).
- ~~Delete Edge desktop icon~~ — закрыто в слайсе 23 (`SystemTweaks.DeleteEdgeDesktopIcon`, см. раздел 9.3/9.13).

**Группа D — Windows PE / установка образа. ЗАКРЫТА почти полностью в слайсе 24 (3 достижимых пункта сделаны, 7 закрыты как «неприменимо», 1 был задвоением уже реализованного).**
- ~~Использовать ключ активации, сохранённый в BIOS/UEFI прошивке~~ — сделано в слайсе 24 (`EditionSettings.Mode=firmware`).
- ~~Отдельное поле продукт-ключа только для активации~~ — сделано в слайсе 24 (`Profile.ActivationKey`).
- ~~Выбор processor architecture~~ — сделано в слайсе 24 (`Profile.ProcessorArchitecture`), но только ОДНО значение за раз — множественная архитектура в одном XML (как у эталона) сознательно не реализована, см. раздел 4/9.14.
- ~~`Allow Windows 11 to be installed without internet connection`~~ — это оказалось задвоением уже реализованного `BypassOnlineAccountRequirement` (тот же `BypassNRO`), не отдельная задача. Исправлено в слайсе 24, раздел 9.14.
- **Неприменимо при текущем архитектурном решении** (все 7 пунктов — части одного механизма замены PE-этапа своим `.cmd`-скриптом с diskpart/dism/bcdboot, что конфликтует с исключением диск-партиционирования, раздел 1 — не «сделать позже», а закрыто как неприменимое, см. 9.14 для деталей):
  - Свой `.cmd`-скрипт для полного PE-этапа (партиционирование + применение образа вручную).
  - Отключение 8.3-имён файлов (`fsutil 8dot3name`).
  - Отключение Windows Defender на этапе PE.
  - Паузы перед разметкой диска / перед финальной перезагрузкой PE-этапа.
  - Compact-режим применения образа, пропуск `/CheckIntegrity /Verify`.
  - Выбор образа для установки по имени/индексу внутри .wim.
  - `$OEM$` distribution share / configuration set.

**Группа E — прочие setup-settings и мелкие механизмы. ЗАКРЫТА ПОЛНОСТЬЮ (6/6) в слайсах 22–23.**

**Группа F — большой архитектурный кусок, решить отдельно, стоит ли вообще брать.**
- **Сырой XML passthrough** для ~80 «сырых» компонентов sysprep (Microsoft-Windows-Audio-AudioCore, TCPIP, TerminalServices-* и т.д., полный список — на странице сайта в разделе «XML markup for more components»). Самый крупный по объёму нереализованный кусок сайта — фактически аварийный люк на все компоненты Windows unattend, которые генератор явно не поддерживает. У сайта это одно текстовое поле на компонент + pass. Для CLI/TUI-инструмента формат ввода такого объёма XML через TUI неочевиден (скорее подходит только CLI/JSON-профилю, не экрану) — решить архитектуру до реализации.

**Не в бэклоге (сознательно исключено или устарело):**
- **Разметка диска** — явно исключено из скоупа (раздел 1), сюда не возвращаемся.
- **Legacy optional features** (PowerShell 2.0) — третий механизм удаления (`Disable-WindowsOptionalFeature`), сознательно пропущен в слайсе 11.
- ~~Windows Fax and Scan~~ — упоминался в старой версии этого файла как второй такой пункт, но в текущей версии сайта (commit `88d81f0`) в списке Remove Bloatware уже не значится — похоже, сайт его убрал; больше не бэклог.
