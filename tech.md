# tech.md — unattend-gen

**Версия: v16** (2026-09-02)

Changelog:
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
      apps.go, personalization.go, scripts.go, review.go
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
      scripts.go                     System/DefaultUser/FirstLogon/UserOnce скрипты, wrapCommand/escapeForOuterCommand
      accountpolicy.go               `net accounts` (истечение пароля, блокировка)
      fileexplorer.go                реестр в default-user hive (видимость файлов, контекстное меню и т.д.)
      personalization.go             реестр в default-user hive (тема, акцентный цвет, обои)
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
- `SystemTweaks` — 17 булевых полей (см. список в разделе 9.3), все опциональны, zero value = ничего не меняется.
- `WifiSettings{SSID (≤32), Authentication: Open|WPA2Personal|WPA3Personal, Password *string, ConnectHidden bool}`.
- `RemovableApp` — строковый enum, 32 значения (список — `profile.RemovableApps`, порядок = порядок в TUI).
- `RemovableFeature` — строковый enum, 7 значений (`profile.RemovableFeatures`): InternetExplorer, WordPad, PowerShellISE, OpenSSHClient, MediaPlayer, Speech, Handwriting.
- `CustomScript{Format: cmd|ps1|reg|vbs, Content string}`.
- `PasswordExpirationSettings{Mode: default|never|custom (""=default), Days *int}`.
- `AccountLockoutSettings{Mode: default|disabled|custom (""=default), Threshold/WindowMinutes/DurationMinutes *int}`.
- `FileExplorerSettings{HiddenFiles: default|show_hidden|show_all, ShowFileExtensions, ClassicContextMenu, HideFolderTooltips, OpenToThisPC, ShowEndTaskInTaskbar bool}` — zero value ничего не меняет.
- `PersonalizationSettings{SystemTheme/AppsTheme: light|dark, AccentColor *string (RRGGBB), ShowAccentOnStartTaskbar/ShowAccentOnTitleBars/DisableTransparency bool, SolidColorWallpaper *string (RRGGBB)}` — только цвета; файл обоев и экран блокировки не входят (см. раздел 15, бэклог).

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

Порядок экранов (`internal/tui/app.go`, `screens.ID`): **Welcome → Language → Accounts → Tweaks → Wifi → Apps → Personalization → Scripts → Review**.

- Каждый экран — отдельный `tea.Model` в `internal/tui/screens/*.go`, общий `*profile.Profile` передаётся через `rebuildScreen` при каждой навигации, так экран всегда синхронизирован с последним состоянием.
- `screens.ScreenApps` — совмещённый экран: чекбоксы удаляемых приложений (`RemoveApps`) и удаляемых компонентов (`RemoveFeatures`) в одной таблице фокуса.
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

Каждый флаг `SystemTweaks` — одна команда в общем списке `RunSynchronousCommand` этого компонента (тот же компонент несёт также команды из `PasswordExpiration`/`AccountLockout`/скриптов System и DefaultUser):

`DisableWindowsUpdate`, `DisableUAC`, `BypassWin11Requirements` (это единственный tweak, реализованный через `Microsoft-Windows-Setup/RunSynchronous` в windowsPE, а не через Deployment/specialize — раньше в загрузке), `DisableSmartAppControl`, `DisableSmartScreen`, `DisableFastStartup`, `DisableSystemRestore`, `EnableLongPaths`, `EnableRemoteDesktop`, `AllowPowerShellScripts`, `DisableLastAccessTimestamp`, `PreventDeviceEncryption`, `DisableAutoSignOnLastUser`, `DisableWPBT`, `AuditProcessCreation`, `HideEdgeFirstRun`, `DisableEdgeStartupBoost`.

Точные реестровые пути/утилиты — в `internal/xmlgen/components/setup.go` и комментариях к каждому tweak; при добавлении нового tweak путь проверяется заново, не копируется по аналогии вслепую.

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

- `apps.go`: сопоставление `RemovableApp → подстроки DisplayName` (не `PackageFamilyName` — те меняются между версиями Windows), команда `Get-AppxProvisionedPackage | Remove-AppxProvisionedPackage`.
- `features.go`: сопоставление `RemovableFeature → префикс DISM capability Name` (обратная ситуация: `Name` стабилен, суффикс `~~~lang~version` — нет), `Get-WindowsCapability -Online | Where Name -like "Prefix*" | Remove-WindowsCapability -Online`.

### 9.7 Политика паролей/блокировки (`accountpolicy.go`)

Единственный механизм без PowerShell/base64-обёртки — статичная команда `net accounts`: `/maxpwage:UNLIMITED|<days>`, `/lockoutthreshold:0` (disabled) или `/lockoutthreshold:N /lockoutwindow:M /lockoutduration:D`. Дефолты Windows (не переопределяются при `Mode=default`): 42 дня истечения пароля, 10 попыток / 10 мин окно / 10 мин блокировка.

### 9.8 Персонализация (`personalization.go`)

`AccentColor` (вход — `RRGGBB`) упаковывается в DWORD `AABBGGRR` (альфа `FF`, байты RGB в обратном порядке) для `DWM\AccentColor`/`ColorizationColor` — единственная содержательная конвертация в этом компоненте, покрыта тестом с известной парой вход/выход. `SolidColorWallpaper` пишет `Control Panel\Colors\Background` как `"R G B"` decimal (тоже конвертация из hex) и очищает `Control Panel\Desktop\Wallpaper`, чтобы показывался цвет, а не картинка.

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

### Бэклог (не начато, без приоритета/очерёдности)

Эти пункты — кандидаты на будущие слайсы. Пока нет отдельного слайса под конкретный пункт, поле/API для него не существует — это `CONTRACT GAP`, если кто-то попробует его использовать.

- **Обои файлом-картинкой.** Нужен новый механизм встраивания файла (base64), сам подход понятен (аналогичен скриптам), не реализован.
- **Экран блокировки картинкой.** Реальный механизм — `PersonalizationCSP`, подтверждённо Enterprise/Education/Pro-SharedPC-only по документации Microsoft — низкая надёжность для основной аудитории инструмента (в основном Home/Pro). Решить, стоит ли реализовывать, до начала работы.
- **Мультиязычность.** Второй/третий язык интерфейса и раскладка, GeoID домашнего региона.
- **Закрепление на Start/панели задач.**
- **Визуальные эффекты, значки рабочего стола.**
- **Гостевые дополнения ВМ** (VirtualBox/VMware/VirtIO/Parallels guest tools).
- **Политика AppLocker.**
- **Сырой XML passthrough** как аварийный люк для нестандартных случаев.
- **HardenACLs, MakeEdgeUninstallable** — оба сознательно пропущены в слайсе 8 как более сложные/рискованные (`MakeEdgeUninstallable` правит JSON-файл, не реестр — другой механизм).
- **Legacy optional features** (PowerShell 2.0, Windows Fax and Scan) — третий механизм удаления (`Disable-WindowsOptionalFeature`), сознательно пропущен в слайсе 11.
- **Разметка диска** — НЕ бэклог, явно исключено из скоупа (раздел 1), сюда не возвращаемся.

Каждый пункт бэклога при взятии в работу оформляется как отдельный слайс: сначала проверка реального механизма (раздел 12), затем реализация, затем — поднятие версии этого файла с добавлением контракта в раздел 4/9.
