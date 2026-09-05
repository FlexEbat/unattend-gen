# tech.md — unattend-gen

**Версия: v22** (2026-09-05)

Changelog:
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
      apps.go, personalization.go, accessibility.go, desktop.go, scripts.go, review.go
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
- `EditionSettings{Mode: generic_key|custom_key|interactive, Edition *WindowsEdition, ProductKey *string}`.
- `UserAccount{Name string (≤20), DisplayName *string, Password *string (nil=без пароля, "" запрещено), Group: Administrators|Users}`.
- `FirstLogon{Mode: first_created_account|builtin_administrator|none, BuiltinAdministratorPassword *string}`.
- `ExpressSettings{Mode: all_disabled|all_enabled|interactive}`.
- `SystemTweaks` — 26 булевых полей (17 из слайса 8 + 6 из слайса 16 + 3 из слайса 18, см. раздел 9.3), все опциональны, zero value = ничего не меняется.
- `WifiSettings{SSID (≤32), Authentication: Open|WPA2Personal|WPA3Personal, Password *string, ConnectHidden bool}`.
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

Порядок экранов (`internal/tui/app.go`, `screens.ID`): **Welcome → Language → Accounts → Tweaks → Wifi → Apps → Personalization → Accessibility → Desktop → Scripts → Review**.

- Каждый экран — отдельный `tea.Model` в `internal/tui/screens/*.go`, общий `*profile.Profile` передаётся через `rebuildScreen` при каждой навигации, так экран всегда синхронизирован с последним состоянием.
- `screens.ScreenApps` — совмещённый экран: чекбоксы удаляемых приложений (`RemoveApps`), удаляемых DISM-компонентов (`RemoveFeatures`) и удаляемых legacy optional features (`RemoveOptionalFeatures`, слайс 17) — три разных механизма, одна таблица фокуса (`internal/tui/screens/apps.go`, `checkboxAt`).
- `screens.ScreenAccessibility` (слайс 19) — Sticky Keys (select + до 6 чекбоксов, только когда `mode=custom`) и Lock Keys (чекбокс «настраивать» + 6 select'ов, видны только если включён — `nil` `LockKeys` иначе, как на сайте-эталоне). Между Personalization и Scripts.
- `screens.ScreenDesktop` (слайс 20) — видимость значков рабочего стола (мастер-чекбокс «настраивать»: выключен = `nil` `DesktopIcons`, включён = все 13 значков получают явный чекбокс) + закреплённые папки на Start (`StartFolders`, простой список без мастер-чекбокса — пустой список сам по себе уже значит «не трогать», доп. переключатель не нужен). Между Accessibility и Scripts.
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

Простые (одна reg.exe-команда): `DisableWindowsUpdate`, `DisableUAC`, `BypassWin11Requirements` (единственный tweak через `Microsoft-Windows-Setup/RunSynchronous` в windowsPE, а не Deployment/specialize — раньше в загрузке), `DisableSmartAppControl`, `DisableSmartScreen`, `DisableFastStartup`, `DisableSystemRestore`, `EnableLongPaths`, `EnableRemoteDesktop`, `AllowPowerShellScripts`, `DisableLastAccessTimestamp`, `PreventDeviceEncryption`, `DisableAutoSignOnLastUser`, `DisableWPBT`, `AuditProcessCreation`, `HideEdgeFirstRun`, `DisableEdgeStartupBoost`, `PreventDeviceApps` (слайс 16), `HardenSystemDriveACL` (слайс 18, `icacls.exe C:\ /remove:g "*S-1-5-11"`, снимает права Authenticated Users на корень системного диска).

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

### Бэклог — полная сверка с schneegans.de (аудит 2026-09-03)

Источник: [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/), commit `88d81f0`. Сверка сделана постранично, раздел за разделом сайта, против фактического кода (не только против старого текста этого файла — там нашлись расхождения, см. ниже).

Каждый пункт при взятии в работу — отдельный слайс: сначала проверка реального механизма (раздел 12), затем реализация, затем поднятие версии этого файла с добавлением контракта в раздел 4/9 и пометкой пункта здесь как сделанного (переносится в раздел «Сделано» выше).

Группы упорядочены по приоритету (первая — предлагаемая следующая, но порядок не жёсткий, решает владелец).

**Группа A — расширение Remove bloatware — ЗАКРЫТО в слайсе 17, кроме Media Features/Recall (см. ниже).**
Реализовано 14 из ~18 недостающих позиций: 10 через существующий Appx-механизм (`apps.go`: Bing Search, Dev Home, Game Assist, Microsoft Store, Notepad (modern), Outlook for Windows, Paint, Wallet, Windows Media Player (modern), Windows Terminal), 4 через существующий DISM-capability-механизм (`features.go`: Facial recognition/Windows Hello — 3 capability, Math Input Panel, OneSync, Steps Recorder), 1 через новый custom-механизм (OneDrive — не Appx-пакет, а файлы + default-user-hive registry run-key, см. `RemoveOneDrive*Command` в `apps.go`). Точные селекторы взяты из `resource/Bloatware.json` исходника-эталона (github.com/cschneegans/unattend-generator), не по памяти. Попутно исправлена неточность в существующем `FeatureSpeech` — у сайта 2 capability-селектора (`Language.Speech` + `Language.TextToSpeech`), в проекте был только первый.

**Media Features и Recall — вынесены из группы A, требовали НОВОГО механизма.** У сайта они (и Remote Desktop Client — уже был неучтён как отдельный гэп) используют `Get-WindowsOptionalFeature`/`Disable-WindowsOptionalFeature` — третий механизм удаления, отдельный от Appx и DISM capability, который раньше в проекте не был реализован (упоминался в бэклоге только применительно к PowerShell 2.0). Слайс 17 реализовал и его: новый файл `internal/xmlgen/components/optionalfeatures.go`, новый тип `profile.RemovableOptionalFeature` + поле `Profile.RemoveOptionalFeatures`, все 3 позиции (Recall, MediaFeatures, RemoteDesktopClient) закрыты этим же слайсом.

**Группа B — оставшиеся System tweaks — ЗАКРЫТА полностью в слайсе 18.**

**Группа C — целые отсутствующие разделы сайта (новая функциональность, скорее всего отдельные экраны TUI). Lock keys/Sticky keys/Desktop icons/Folders on Start ЗАКРЫТЫ в слайсах 19–20, остальное открыто.**
- **Start menu and taskbar** — самый крупный из отсутствующих разделов: режим отображения поля поиска в таскбаре, конфигурация закреплённых иконок таскбара через XML, отключение виджетов, left-align таскбара (Win11), скрытие кнопки Task View, «always show tray icons», отключение Bing-результатов в поиске, плитки Start (Win10) и pins (Win11) через XML/JSON — НЕ то же самое, что закрытые в слайсе 20 «Folders on Start» (папки у кнопки питания).
- **Visual effects** — полный набор чекбоксов производительности/анимации (уже был в старом бэклоге под общим названием «визуальные эффекты»).
- **Delete Edge desktop icon** (`DeleteEdgeDesktopIcon`) — замечен при реализации слайса 20 (соседний код в `Optimizations.cs`), но сознательно не взят в тот слайс: отдельный простой булев твик (2 `Remove-Item` — `C:\Users\Public\Desktop\Microsoft Edge.lnk` в specialize + `%USERPROFILE%\Desktop\Microsoft Edge.lnk` в UserOnce), не часть словаря видимости иконок. Дешёвый кандидат на следующий заход по System tweaks или Desktop icons.
- **VM hosts: Core isolation toggle** — включить/выключить virtualization-based security. Отдельно от VM guest tools (ниже), не то же самое.
- **VM guest tools** — установка VirtualBox Guest Additions / VMware Tools / VirtIO / Parallels Tools. Известный пункт.
- **AppLocker policy** — известный пункт, готовый XML-шаблон с сайта можно взять за основу.

**Группа D — Windows PE / установка образа (другой pass, другая часть жизненного цикла).**
Не было в старом бэклоге вообще, обнаружено при этом аудите:
- Использовать ключ активации, сохранённый в BIOS/UEFI прошивке (не вводить заново).
- Свой `.cmd`-скрипт для полного PE-этапа (партиционирование + применение образа вручную) — конфликтует с явным исключением диск-партиционирования из скоупа (раздел 1), нужно решить отдельно, скорее всего не брать.
- Отключение 8.3-имён файлов (`fsutil 8dot3name`).
- Отключение Windows Defender на этапе PE (через загрузку куста SYSTEM и правку `Start`-значений сервисов).
- Паузы перед разметкой диска / перед финальной перезагрузкой PE-этапа.
- Compact-режим применения образа, пропуск `/CheckIntegrity /Verify`.
- Выбор образа для установки по имени/индексу внутри .wim, а не только по edition.
- Отдельное поле продукт-ключа только для **активации** (независимо от ключа установки).
- Выбор нескольких **processor architectures** в одном XML (x86/x64/ARM64).
- `Allow Windows 11 to be installed without internet connection` — отдельный чекбокс PE-этапа, НЕ совпадает по механизму с существующим `BypassOnlineAccountRequirement` (тот — про OOBE-экраны после установки, этот — про сам Windows Setup).
- `$OEM$` distribution share / configuration set — копирование содержимого папки `$OEM$` на целевой диск.

**Группа E — прочие setup-settings и мелкие механизмы.**
- Глобальный переключатель «Hide PowerShell windows during setup» (`-WindowStyle Hidden` вместо `Normal` для всех PS-скриптов сразу, не только пользовательских).
- «Keep sensitive files» — по умолчанию сайт удаляет `C:\Windows\Panther\unattend.xml` (и `-original.xml`) и `C:\Windows\Setup\Scripts\Wifi.xml` после установки; в проекте это поведение вообще не описано и не управляется ни в какую сторону.
- Автозапуск Narrator во время установки и после логона.
- Динамическое имя компьютера через PowerShell-скрипт (сейчас — только статичная строка или «пусть Windows сгенерирует»).
- Base64-обфускация паролей аккаунтов в самом сгенерированном XML (не то же самое, что base64-обёртка контента скриптов, которая уже есть).
- Импорт готового WLAN-профиля как raw XML (`netsh wlan export profile key=clear`) в дополнение к текущему ручному вводу SSID/пароль.

**Группа F — большой архитектурный кусок, решить отдельно, стоит ли вообще брать.**
- **Сырой XML passthrough** для ~80 «сырых» компонентов sysprep (Microsoft-Windows-Audio-AudioCore, TCPIP, TerminalServices-* и т.д., полный список — на странице сайта в разделе «XML markup for more components»). Самый крупный по объёму нереализованный кусок сайта — фактически аварийный люк на все компоненты Windows unattend, которые генератор явно не поддерживает. У сайта это одно текстовое поле на компонент + pass. Для CLI/TUI-инструмента формат ввода такого объёма XML через TUI неочевиден (скорее подходит только CLI/JSON-профилю, не экрану) — решить архитектуру до реализации.

**Не в бэклоге (сознательно исключено или устарело):**
- **Разметка диска** — явно исключено из скоупа (раздел 1), сюда не возвращаемся.
- **Legacy optional features** (PowerShell 2.0) — третий механизм удаления (`Disable-WindowsOptionalFeature`), сознательно пропущен в слайсе 11.
- ~~Windows Fax and Scan~~ — упоминался в старой версии этого файла как второй такой пункт, но в текущей версии сайта (commit `88d81f0`) в списке Remove Bloatware уже не значится — похоже, сайт его убрал; больше не бэклог.
