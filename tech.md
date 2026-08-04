# tech.md — unattend-gen (TUI + CLI генератор autounattend.xml)

**Версия: v1** (2026-08-04)

Changelog:
- v1 — первая версия. Стек: Go/Bubble Tea (раздел 2), кросс-платформенный статический бинарник. Заморожены на эту версию: модель `Profile`, API `profile`/`xmlgen`, набор TUI-примитивов (разделы 4–6) — контракт для стадий 1–4 из раздела 15. Контракт для стадий 5+ (раздел 15) ещё не зафиксирован и добавляется отдельными правками файла по мере реализации каждой стадии, через ту же процедуру `CONTRACT GAP` (раздел 13).

---

## 0. Как читать этот файл

Это источник истины проекта. Читай его перед каждой задачей и подчиняйся дословно.

- Имена полей, типов, функций и XML-компонентов берутся отсюда. Свои варианты не придумывай.
- Нужного контракта здесь нет → СТОП, выдай блок `CONTRACT GAP` (раздел 13). Код с выдуманным полем или компонентом не пиши.
- Файл меняет только владелец проекта. Каждое изменение контракта поднимает версию сверху файла.
- Работаешь один слайс за заход. Стадии и слайсы — раздел 15, иди по нему сверху вниз.

---

## 1. Проект

CLI- и TUI-утилита на Go, которая собирает `autounattend.xml` — файл ответов для автоматической установки Windows 10/11 (см. [документацию Microsoft](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/update-windows-settings-and-scripts-create-your-own-answer-file-sxs)). Прототип — веб-сервис [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/), и цель проекта — полная функциональная копия этого сервиса в виде TUI/CLI: все его разделы (язык и раскладки, разметка диска и WinPE, edition и product key, архитектуры процессора, setup settings, имя компьютера, часовой пояс, учётные записи, политики паролей и блокировки, твики File Explorer/Start/taskbar, полный список системных твиков, визуальные эффекты, значки рабочего стола, VM host/guest, Wi-Fi, express settings, клавиши-модификаторы, персонализация, удаление bloatware, пользовательские скрипты, AppLocker, произвольная XML-разметка компонентов) должны в итоге появиться в этом инструменте.

Пользователь либо заполняет форму в терминале (TUI, Bubble Tea), либо готовит JSON-профиль и вызывает CLI-команду. Оба пути используют одну и ту же структуру `Profile` и один и тот же сборщик XML. Результат — файл `autounattend.xml`, который можно положить на установочный USB-накопитель или примонтировать в виртуальную машину.

Утилита собирается в один статический бинарник под Linux/macOS/Windows и запускается там, где сейчас находится пользователь — не обязательно на Windows: файл ответов чаще всего готовят заранее, с другой машины. Работает офлайн, без сервера, без сети, без аутентификации. Профили — обычные JSON-файлы на диске, не база данных.

Полный охват собирается не за один заход и не за одну версию контракта: раздел 15 разбивает весь функционал исходного сервиса на стадии и слайсы, раздел 4 фиксирует контракт по мере того, как стадия добирается до реализации. Раздел, которого пока нет ни в разделе 4, ни в разделе 15 (то есть стадия для него ещё не расписана) — это `CONTRACT GAP` (раздел 13), а не повод придумать поле на месте: сначала стадия добавляется в раздел 15 и её контракт — в раздел 4, отдельной правкой этого файла с поднятием версии, и только потом пишется код.

---

## 2. Стек

- Go 1.23+
- `spf13/cobra` — CLI-команды
- `charmbracelet/bubbletea` — TUI, архитектура Model/Update/View (Elm-стиль)
- `charmbracelet/bubbles` — готовые TUI-компоненты (`textinput`, `table`, `list`) как основа для примитивов раздела 6
- `charmbracelet/lipgloss` — стили и раскладка TUI без ручной работы с ANSI-кодами
- `encoding/xml`, `encoding/json` (стандартная библиотека) — сборка XML и чтение/запись профилей. Сторонних XML/JSON-библиотек не добавлять.
- `go-playground/validator/v10` — валидация полей `Profile` через struct-теги, дополняется ручными правилами там, где тегов недостаточно (раздел 8)
- Тесты: стандартный `testing` + `go test`
- Проверки: `gofmt`, `go vet`, `golangci-lint`

Без ORM, без базы данных, без веб-фреймворков, без сетевых вызовов. Профили — JSON-файлы, читаются и пишутся напрямую через `os`/`encoding/json`. Модуль собирается без CGO (`CGO_ENABLED=0`), чтобы кросс-компиляция под все три ОС работала без внешних тулчейнов.

---

## 3. Структура папок

```
cmd/
  unattend-gen/
    main.go                    точка входа, вызывает cli.Execute()
internal/
  cli/
    root.go                    корневая cobra-команда
    profile.go                 profile init / profile list
    validate.go                команда validate
    generate.go                команда generate
    tui.go                     команда tui
  profile/
    schema.go                  структура Profile и вложенные типы (раздел 4)
    validate.go                ValidateProfile(data []byte) ValidationResult — чистая логика, без диска и без XML
    store.go                   LoadProfile, SaveProfile, ListProfiles
    validate_test.go
    store_test.go
  xmlgen/
    builder.go                 BuildAnswerFile(profile *Profile) (string, error) — точка входа сборки
    components/
      international.go         Microsoft-Windows-International-Core
      setup.go                 Microsoft-Windows-Setup (product key, EULA, bypass Win11)
      shellsetup.go            Microsoft-Windows-Shell-Setup (имя компьютера, учётки, OOBE, твики)
      wlan.go                  Microsoft-Windows-WLANSVC (слайс 6)
    builder_language_test.go
    builder_accounts_test.go
    builder_tweaks_test.go
    builder_wifi_test.go
  tui/
    app.go                     корневая bubbletea.Model, переключение экранов
    screens/
      welcome.go                выбор/создание профиля
      language.go                язык, локаль, раскладка, edition/product key
      accounts.go                 имя компьютера, учётные записи, первый вход
      tweaks.go                   express settings, системные твики
      wifi.go                      Wi-Fi профиль
      review.go                    предпросмотр и сохранение .xml
    widgets/                    TUI-примитивы из раздела 6, обёртки над bubbles
presets/
  minimal.json                 минимальный набор полей, остальное — значения по умолчанию
  single-user.json             один локальный администратор, телеметрия выключена
go.mod
go.sum
.golangci.yml
```

Правила размещения:

- `internal/xmlgen` не читает диск, не импортирует `cobra`/`bubbletea` и не знает о CLI/TUI. Вход — `*profile.Profile`, выход — строка XML или элемент дерева.
- CLI и TUI не строят XML вручную и не работают с `encoding/xml` напрямую. Любая сборка идёт через `xmlgen.BuildAnswerFile`.
- `internal/profile/validate.go` не импортирует `internal/xmlgen` и не пишет на диск, чтобы тестироваться без сборки XML.
- `internal/profile/store.go` не валидирует содержимое — только читает/пишет JSON и отдаёт/принимает уже готовый `*Profile`. Валидация сырых данных до создания `Profile` — на вызывающей стороне (CLI/TUI), через `ValidateProfile`.
- Пакеты лежат под `internal/`, чтобы наружу модуля ничего не экспортировалось случайно — это не библиотека, а приложение.

---

## 4. Контракт данных (заморожен)

Структуры `Profile` (`internal/profile/schema.go`):

```go
package profile

type LanguageSettings struct {
	UILanguage      string `json:"ui_language"`      // BCP-47, напр. "en-US" — язык интерфейса Windows
	Locale          string `json:"locale"`            // BCP-47 — региональный формат чисел/дат
	KeyboardLayout  string `json:"keyboard_layout"`   // напр. "en-US", "de-DE"
}

type EditionMode string

const (
	EditionModeGenericKey  EditionMode = "generic_key"
	EditionModeCustomKey   EditionMode = "custom_key"
	EditionModeInteractive EditionMode = "interactive"
)

type WindowsEdition string

const (
	EditionHome       WindowsEdition = "Home"
	EditionPro        WindowsEdition = "Pro"
	EditionEducation  WindowsEdition = "Education"
	EditionEnterprise WindowsEdition = "Enterprise"
)

type EditionSettings struct {
	Mode       EditionMode     `json:"mode"`
	Edition    *WindowsEdition `json:"edition"`     // обязателен, когда Mode == EditionModeGenericKey
	ProductKey *string         `json:"product_key"` // обязателен, когда Mode == EditionModeCustomKey
}

type AccountGroup string

const (
	GroupAdministrators AccountGroup = "Administrators"
	GroupUsers          AccountGroup = "Users"
)

type UserAccount struct {
	Name        string        `json:"name"`
	DisplayName *string       `json:"display_name"`
	Password    *string       `json:"password"` // пустая строка недопустима, nil = без пароля
	Group       AccountGroup  `json:"group"`
}

type FirstLogonMode string

const (
	FirstLogonFirstCreatedAccount FirstLogonMode = "first_created_account"
	FirstLogonBuiltinAdmin        FirstLogonMode = "builtin_administrator"
	FirstLogonNone                FirstLogonMode = "none"
)

type FirstLogon struct {
	Mode                         FirstLogonMode `json:"mode"`
	BuiltinAdministratorPassword *string        `json:"builtin_administrator_password"` // обязателен при Mode == FirstLogonBuiltinAdmin
}

type ExpressSettingsMode string

const (
	ExpressAllDisabled ExpressSettingsMode = "all_disabled"
	ExpressAllEnabled  ExpressSettingsMode = "all_enabled"
	ExpressInteractive ExpressSettingsMode = "interactive"
)

type ExpressSettings struct {
	Mode ExpressSettingsMode `json:"mode"`
}

type SystemTweaks struct {
	DisableWindowsUpdate      bool `json:"disable_windows_update"`
	DisableUAC                bool `json:"disable_uac"`
	BypassWin11Requirements   bool `json:"bypass_win11_requirements"`
}

type WifiAuthentication string

const (
	WifiOpen           WifiAuthentication = "Open"
	WifiWPA2Personal    WifiAuthentication = "WPA2Personal"
	WifiWPA3Personal    WifiAuthentication = "WPA3Personal"
)

type WifiSettings struct {
	SSID           string             `json:"ssid"`
	Authentication WifiAuthentication `json:"authentication"`
	Password       *string            `json:"password"` // обязателен, когда Authentication != WifiOpen
	ConnectHidden  bool               `json:"connect_hidden"`
}

type Profile struct {
	SchemaVersion   int             `json:"schema_version"` // всегда 1 в этой версии контракта
	Name            string          `json:"name"`            // имя профиля, совпадает с именем файла без .json
	Language        LanguageSettings `json:"language"`
	Edition         EditionSettings  `json:"edition"`
	ComputerName    *string          `json:"computer_name"` // nil = Windows сгенерирует случайное имя
	Accounts        []UserAccount    `json:"accounts"`       // 0..5
	FirstLogon      FirstLogon       `json:"first_logon"`
	ExpressSettings ExpressSettings  `json:"express_settings"`
	SystemTweaks    SystemTweaks     `json:"system_tweaks"`
	Wifi            *WifiSettings    `json:"wifi"` // nil = настройка Wi-Fi пропускается
}
```

Правила:

- `SchemaVersion` всегда `1` в этой версии контракта. Поле существует, чтобы будущие версии профиля могли определить формат при загрузке.
- Полей `MultiLanguage`, `Partitioning`, `BloatwareRemoval`, `CustomScripts`, `AppLocker`, `CustomXML` в v1 нет. Понадобилось — `CONTRACT GAP`.
- `ComputerName`, если задан, должен пройти правила из раздела 8 (валидация NetBIOS-имени) — это делает `ValidateProfile`, не сама структура и не `xmlgen`.
- `Accounts` — от 0 до 5 записей, как в исходном сервисе. Больше пяти — `CONTRACT GAP`.
- Опциональные поля — указатели (`*string`, `*WifiSettings` и т.д.), а не пустая строка/нулевое значение. Отсутствие значения и пустое значение — разные состояния, путать их нельзя.

---

## 5. API (заморожен)

`internal/profile/validate.go`:

```go
type ValidationResult struct {
	Profile *Profile // nil, если Errors не пуст
	Errors  []string // тексты на русском, раздел 8
}

func ValidateProfile(data []byte) ValidationResult
```

`internal/profile/store.go`:

```go
func LoadProfile(path string) (*Profile, error)
func SaveProfile(profile *Profile, path string) error
func ListProfiles(dir string) ([]string, error)
```

`internal/xmlgen/builder.go`:

```go
func BuildAnswerFile(profile *Profile) (string, error)
// возвращает готовый XML-документ (с XML-декларацией, UTF-8, с отступами),
// готовый к записи как autounattend.xml
```

Правила:

- `BuildAnswerFile` принимает только уже валидный `*Profile`. Валидацию входных данных делает вызывающий слой (CLI/TUI) через `ValidateProfile` до вызова.
- `BuildAnswerFile` возвращает `error` только на непредвиденную внутреннюю ошибку сериализации (по сути недостижимо на валидном профиле). Ожидаемые развилки профиля (`Wifi == nil`, `Accounts` пустой и т.д.) — штатные состояния, не ошибки, и `error` на них не возвращается.
- `LoadProfile` возвращает `*Profile` или оборачивает `os.ErrNotExist` / ошибку `encoding/json` через `fmt.Errorf("%w", ...)` — их перехватывает CLI/TUI и печатает читаемое сообщение. `store.go` сам ошибки не форматирует для пользователя.
- `ListProfiles` возвращает пути к `*.json` в каталоге, отсортированные по имени. Не читает содержимое файлов и не валидирует их.

---

## 6. TUI-примитивы (собираются в каркасе)

Лежат в `internal/tui/widgets/`, используются во всех экранах. Свой инпут или свою таблицу на отдельном экране не писать.

| компонент             | база                          | назначение                                                          |
| ---------------------- | ------------------------------ | ---------------------------------------------------------------------- |
| `LabeledInput`          | `bubbles/textinput`             | текстовое поле с подписью и слотом под текст ошибки                    |
| `LabeledSelect`         | `bubbles/list`                  | список выбора с подписью, значения из констант enum-типа               |
| `PasswordInput`         | `bubbles/textinput` (`EchoMode: EchoPassword`) | текстовое поле с маскировкой ввода                            |
| `AccountsTable`         | `bubbles/table`                  | редактируемая таблица учётных записей (до 5 строк), колонки из `UserAccount` |
| `ConfirmBar`            | `lipgloss` (свёрстанная строка) | нижняя панель: «Назад», «Далее», «Сохранить», подсказки по горячим клавишам |

Отдельного экрана-витрины примитивов нет: примитивы вводятся в каркасе и сразу используются первым экраном. Виджет, который не используется ни одним экраном, не пишется.

---

## 7. CLI-команды и TUI-экраны

### CLI

| команда                                   | что делает                                                        |
| ------------------------------------------ | -------------------------------------------------------------------- |
| `unattend-gen profile init <name> [--preset NAME]` | создаёт `<name>.json` в текущем каталоге из пресета или из значений по умолчанию |
| `unattend-gen profile list`                 | печатает список профилей из `./profiles`                            |
| `unattend-gen validate <profile.json>`      | проверяет профиль, код выхода 0/1                                     |
| `unattend-gen generate <profile.json> [-o PATH]` | валидирует и пишет `autounattend.xml` (по умолчанию рядом с профилем) |
| `unattend-gen tui [profile.json]`           | открывает TUI, опционально сразу с загруженным профилем              |

Правила:

- `generate` с невалидным профилем ничего не пишет на диск, печатает список ошибок в stderr и завершается с кодом 1.
- `generate` при успехе печатает путь к записанному файлу в stdout и завершается с кодом 0.
- `profile init` без `--preset` создаёт профиль с пустыми учётками и `express_settings.mode == "interactive"`; с `--preset minimal|single-user` — из соответствующего файла в `presets/`.

### TUI-экраны

| экран        | что делает                                                              |
| ------------- | -------------------------------------------------------------------------- |
| `welcome`     | выбор существующего профиля из `./profiles` или создание нового            |
| `language`    | язык/локаль/раскладка, edition и product key                                |
| `accounts`    | имя компьютера, таблица учётных записей, выбор first logon                  |
| `tweaks`      | express settings, системные твики (раздел 4: `SystemTweaks`)                |
| `wifi`        | включить/выключить настройку Wi-Fi и её поля                                |
| `review`      | предпросмотр итогового XML, сохранение профиля и/или экспорт `autounattend.xml` |

Навигация линейная («Далее»/«Назад»), но переход на `review` возможен с любого экрана — профиль хранится целиком в корневой `Model` приложения и достраивается по мере заполнения экранов, а не пересобирается заново. Поле, не заполненное пользователем, держит значение по умолчанию (раздел 8), а не пустую строку. Каждый экран — своя `bubbletea.Model` с `Update`/`View`, переключение экранов — через `tea.Msg`, обрабатываемый корневой моделью.

---

## 8. Валидация

`internal/profile/validate.go`, чистая функция без импорта `xmlgen` и без обращения к диску:

```go
func ValidateProfile(data []byte) ValidationResult
```

Правила:

- `Language.UILanguage`, `Locale`, `KeyboardLayout` — непустые строки, вид BCP-47 (`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`). Не проверяется по справочнику языков Windows — только формат.
- `Edition.Mode == EditionModeGenericKey` требует `Edition.Edition` заданным, `ProductKey` игнорируется.
- `Edition.Mode == EditionModeCustomKey` требует `ProductKey` вида `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX` (латиница и цифры, 5 групп по 5 символов).
- `ComputerName`, если задан: 1–15 символов, латинские буквы/цифры/дефис, не состоит только из цифр, не начинается и не заканчивается дефисом (правила NetBIOS-имени).
- `Accounts`: 0–5 записей. `Name` непустой, до 20 символов, без символов `" / \ [ ] : ; | = , + * ? < >`. Пароль, если задан, непустой (пустая строка — ошибка, используйте `nil`).
- `FirstLogon.Mode == FirstLogonFirstCreatedAccount` требует, чтобы `Accounts` был не пуст.
- `FirstLogon.Mode == FirstLogonBuiltinAdmin` требует непустой `BuiltinAdministratorPassword`.
- `Wifi`, если задан: `SSID` 1–32 символа; `Authentication != WifiOpen` требует непустой `Password` (для WPA2/WPA3 — минимум 8 символов).
- Базовые проверки (обязательность, длина, `oneof` для enum-строк) — через struct-теги `validator` (`validate:"required,max=15"` и т.п.). Правила, которые тегами не выражаются (формат BCP-47, кросс-полевые условия вроде «пароль обязателен, если аутентификация не Open»), — руками в `ValidateProfile` после `validator.Struct`.
- Тексты ошибок на русском, по одной строке на ошибку: `Имя компьютера длиннее 15 символов`, `Пароль обязателен для встроенной учётной записи Administrator`, `SSID не может быть пустым`.

---

## 9. Тесты

Тесты выводятся из критериев приёмки слайса (раздел 15), а не из готового кода. Тест кодирует критерий, а не повторяет реализацию.

Запрещённый вид теста: «собрали XML, проверили, что он не пустой». Такой тест зелёный и на сломанной сборке.

На каждый слайс минимум:

- тест на каждый критерий приёмки, где есть проверяемое условие;
- тест на путь ошибки (невалидный профиль, отсутствующий файл);
- тесты чистой логики (`validate.go`) без сборки XML и без диска.

Как проверять результат `BuildAnswerFile`: парсить возвращённую строку через `encoding/xml.Unmarshal` во вспомогательную struct теста и проверять конкретные поля/атрибуты, а не делать `strings.Contains(xml, "Pro")`. Сравнение с образцовым XML-файлом целиком не используется — билдер меняется по слайсам, такой тест ломался бы на каждом слайсе.

Тесты `store.go` используют `t.TempDir()`, реальный каталог профилей не трогают.

CLI-команды тестируются вызовом `cobra.Command.Execute()` с подменённым `SetArgs`/`SetOut`/`SetErr`, без реального терминала. TUI-экраны тестируются через `teatest` (`charmbracelet/x/exp/teatest`) — программный ввод клавиш, снятие состояния корневой `Model`, не пиксельный снапшот экрана.

Файлы тестов лежат рядом с кодом (`schema.go` → `schema_test.go`), как принято в Go, отдельного каталога `tests/` нет.

---

## 10. Проверки и гейт

Makefile (или `go run` эквиваленты, если Makefile не заводится):

```
fmt:        gofmt -w .
fmt-check:  test -z "$$(gofmt -l .)"
vet:        go vet ./...
lint:       golangci-lint run
test:       go test ./... -race
gate:       make fmt-check && make vet && make lint && make test
```

`gate` красный → слайс не закрыт. Правки по гейту делаются в том же заходе.

Что ловит каждая команда:

- `gofmt -l` — расхождения форматирования (список файлов непустой = есть расхождения);
- `go vet` — подозрительные конструкции (неверные `Printf`-форматы, недостижимый код и т.п.);
- `golangci-lint` — мёртвый код, неиспользуемые импорты/переменные, стиль (набор линтеров фиксируется в `.golangci.yml`, минимум `govet`, `staticcheck`, `unused`, `errcheck`);
- `go test -race` — критерии приёмки, гонки данных в TUI-моделях.

---

## 11. Конвенция текста

Код, комментарии, коммиты, сообщения об ошибках в консоль (`error`-значения, логи) — по-английски. Тексты TUI и валидации — по-русски.

Коммиты, Conventional Commits, фиксированный формат:

```
type(scope): summary
```

- `type` из набора `feat|fix|test|refactor|chore|docs`;
- `summary` в императиве, со строчной буквы, без точки, до 50 символов;
- тело только чтобы объяснить *почему*, не *что*;
- коммитить по ходу работы маленькими шагами, не одним коммитом в конце. Каждый коммит по возможности проходит сборку (`go build ./...`).

Примеры: `feat(xmlgen): add international-core component`, `test(validate): cover custom product key format`.

Правила письма для всей прозы проекта (коммиты, комментарии, README): активный залог, конкретика вместо общих фраз, без вводных оборотов, без длинного тире, без наречий-усилителей. Комментарий объясняет причину решения, а не пересказывает соседнюю строку. Закомментированный код не оставлять.
Перед написанием README посмотри [https://docs.github.com/en](https://docs.github.com/en), [https://github.com/matiassingers/awesome-readme](https://github.com/matiassingers/awesome-readme), [https://www.makeareadme.com/](https://www.makeareadme.com/).

---

## 12. Definition of Done одного слайса

1. Критерии приёмки слайса из раздела 15 выполнены, проверены руками (CLI-вызов или прогон TUI из собранного бинарника).
2. Тесты написаны из критериев, `gate` зелёный.
3. Новых файлов и абстракций сверх описанных в этом файле нет.
4. Мёртвого кода нет: неиспользуемых экспортов, виджетов, компонентов XML.
5. Коммиты по конвенции раздела 11.
6. `tech.md` не изменён (изменение контракта идёт отдельно, через раздел 13).

---

## 13. CONTRACT GAP

Не хватает поля, компонента XML, команды или экрана — работа останавливается. Выдай блок и жди ответа:

```
CONTRACT GAP
Что нужно: <поле/компонент/команда>
Зачем: <какой критерий приёмки без него не выполняется>
Предлагаемая форма: <точная сигнатура, поле структуры или XML-компонент с pass>
Что делаю пока: <заглушка локально в своём слайсе / жду>
```

Код с выдуманным полем структуры или выдуманным XML-компонентом не пиши. Структуру `Profile`, API из раздела 5, список TUI-примитивов сам не расширяй.

---

## 14. Правила поведения в сессии

- Думай до кода: назови допущения, спроси при неоднозначности, покажи варианты вместо молчаливого выбора.
- Простота: никаких фич сверх запрошенного, никаких абстракций под одноразовый код, никакой обработки ошибок, которых не бывает.
- Хирургические правки: соседний рабочий код не улучшать и не рефакторить. Каждая изменённая строка следует из текущей задачи.
- Один слайс за заход. Не выкатывай весь генератор сразу.
- Ревью идёт вторым заходом, после того как слайс готов, а не в том же сообщении, где написан код.

---

## 15. Стадии и слайсы

Порядок жёсткий, сверху вниз. Один слайс за один заход сессии. Слайс закрыт, когда выполнен Definition of Done из раздела 12.

Слайс вертикальный: от поля в `Profile` до элемента в сгенерированном XML или до экрана TUI. Половина слайса не закрывается.

- **Стадия 1 — каркас.** Слайс 0. Фичи не начинаются, пока критерии каркаса не зелёные целиком.
- **Стадия 2 — MVP.** Слайсы 1–3: язык и edition, имя компьютера и учётные записи, CLI end-to-end.
- **Стадия 3 — расширение.** Слайс 4: express settings и системные твики.
- **Стадия 4 — удобство.** Слайсы 5–6: TUI и пресеты, Wi-Fi.
- **Стадии 5+ — полный охват исходного сервиса.** Список ниже, в порядке реализации. Контракт (структуры, поля, XML-компоненты) для каждой стадии не расписан заранее — фиксируется в разделе 4 отдельной правкой файла непосредственно перед тем, как стадия берётся в работу, тем же порядком, что уже сделан для стадий 1–4. До этой правки соответствующий функционал — `CONTRACT GAP`.

Роадмап стадий 5+ (каждая — несколько слайсов по образцу разделов 15.0–15.6, состав слайсов внутри стадии определяется при фиксации её контракта):

| стадия | охватывает (по разделам исходного сервиса) |
| --- | --- |
| 5. Разметка диска и WinPE | интерактивный/сгенерированный WinPE-скрипт, выбор образа (имя/индекс/интерактивно), партиционирование (GPT/MBR/custom diskpart), Windows RE, disable 8.3, ассерты про диск |
| 6. Windows edition, расширенно | N-редакции, ключ из BIOS/UEFI, выбор образа по product key, несколько архитектур процессора в одном файле |
| 7. Setup settings и время | bypass Windows 11 (полный набор), офлайн-установка без интернета, скрытие PowerShell-окон, keep sensitive files, Narrator, часовой пояс, computer name через PowerShell-скрипт, политика истечения паролей, Account Lockout policy |
| 8. File Explorer, Start, taskbar | видимость скрытых файлов, классическое контекстное меню, панель поиска в taskbar, закреплённые значки (XML), Start-плитки/pins (Windows 10/11), tray-иконки |
| 9. Системные твики, полный список | весь чек-лист System tweaks исходного сервиса сверх уже заданного в разделе 4 (SmartScreen, Fast Startup, System Restore, long paths, RDP, ACL, junction points, execution policy и т.д.) |
| 10. Внешний вид | визуальные эффекты (custom/performance/appearance), значки рабочего стола, папки на Start |
| 11. VM host/guest | core isolation, VirtualBox/VMware/VirtIO/Parallels guest additions |
| 12. Персонализация | цветовая тема, акцентный цвет, обои (включая PowerShell-скрипт загрузки), экран блокировки |
| 13. Remove bloatware | полный чек-лист встроенных приложений на удаление |
| 14. Пользовательские скрипты | `System`/`DefaultUser`/`FirstLogon`/`UserOnce`, форматы `.cmd`/`.ps1`/`.reg`/`.vbs`, restart Explorer |
| 15. AppLocker | применение готовой AppLocker-политики |
| 16. Произвольная XML-разметка | вставка XML-маркапа для компонентов, не покрытых генератором напрямую |

Порядок стадий 5–16 ориентировочный и может быть изменён владельцем проекта до фиксации контракта следующей стадии — но раздел 15 меняется отдельной правкой, не по инициативе исполнителя слайса.

---

### Слайс 0 — каркас

**Статус: выполнено.**

**Что собрать:**

- Модуль Go (`go.mod`), зависимости: `spf13/cobra`, `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`, `go-playground/validator/v10`.
- `Makefile` с целями `fmt`, `fmt-check`, `vet`, `lint`, `test`, `gate` из раздела 10, `.golangci.yml` с минимальным набором линтеров.
- `internal/profile/schema.go` — структура `Profile` и вложенные типы из раздела 4 целиком (даже те поля, которые заполнятся только в следующих слайсах — значения по умолчанию должны давать валидный профиль).
- `internal/xmlgen/builder.go` — `BuildAnswerFile`, возвращает минимальный валидный skeleton: XML-декларация, корневой `<unattend xmlns="urn:schemas-microsoft-com:unattend">`, без компонентов внутри.
- `internal/profile/store.go` — `LoadProfile`, `SaveProfile`, `ListProfiles`.
- `internal/profile/validate.go` — `ValidateProfile` с правилами, которые уже применимы к полям слайса 0 (структура, `SchemaVersion`).
- `internal/cli/root.go`, `internal/cli/profile.go`, `internal/cli/validate.go` — команды `profile init` (без пресетов) и `validate`.
- `cmd/unattend-gen/main.go` — точка входа.
- `.gitignore`: `/unattend-gen`, `*.exe`, `profiles/*.json` кроме `presets/`.

**Критерии приёмки:**

1. `make gate` зелёный на пустом проекте.
2. `go build -o unattend-gen ./cmd/unattend-gen` собирается без ошибок под `GOOS=linux`, `GOOS=darwin`, `GOOS=windows` (проверить кросс-компиляцией, `CGO_ENABLED=0`).
3. `unattend-gen profile init demo` создаёт `demo.json`, который проходит `unattend-gen validate demo.json`.
4. `LoadProfile(path)` после `SaveProfile(p, path)` возвращает структуру, равную исходной `p` (`reflect.DeepEqual` или сравнение через `cmp.Diff`).
5. `BuildAnswerFile` на профиле по умолчанию отдаёт строку, которая парсится `encoding/xml.Unmarshal` без ошибок.

**Тесты:** `internal/profile/store_test.go` — сохранение/загрузка профиля по умолчанию; `internal/profile/validate_test.go` — валидный профиль по умолчанию проходит, профиль с `schema_version: 2` не проходит.

**Не делать:** TUI-экраны, XML-компоненты кроме корня, пресеты, CLI-команду `generate`.

---

### Слайс 1 — язык и edition (эталон)

**Статус: выполнено.**

Это эталонная вертикаль. Слайсы 2–4 повторяют её структуру: поле в `Profile` → правило в `validate.go` → компонент в `xmlgen/components/` → проверка через CLI.

**Файлы:** `internal/xmlgen/components/international.go`, `internal/xmlgen/components/setup.go`, правила в `internal/profile/validate.go` для `Language` и `Edition`.

**Что делает:** `BuildAnswerFile` добавляет в вывод компонент `Microsoft-Windows-International-Core` (проходы `windowsPE` и `specialize`) с `UILanguage`, `SystemLocale`, `UserLocale`, `InputLocale` из `profile.Language`, и компонент `Microsoft-Windows-Setup` (проход `windowsPE`) с `UserData/ProductKey/Key`, заполненным по правилам `Edition.Mode`.

**Критерии приёмки:**

1. `Edition.Mode == EditionModeGenericKey` — в XML попадает встроенный generic-ключ для выбранного `Edition.Edition` (таблица ключей — константа в `setup.go`, не пользовательский ввод).
2. `Edition.Mode == EditionModeCustomKey` — в XML попадает ровно `ProductKey` из профиля.
3. `Edition.Mode == EditionModeInteractive` — элемент `ProductKey` в XML отсутствует.
4. `Language.UILanguage`, `Locale`, `KeyboardLayout` из профиля дословно попадают в соответствующие атрибуты `International-Core`.
5. Невалидный BCP-47 в `Language.UILanguage` — `ValidateProfile` возвращает ошибку, `BuildAnswerFile` не вызывается.

**Тесты (`internal/xmlgen/builder_language_test.go`, дополнения в `validate_test.go`):** по критериям 1–5, через `encoding/xml.Unmarshal` на результате `BuildAnswerFile`.

**Не делать:** множественные языки (второй/третий), таблицу всех Windows-локалей, N-редакции (Pro N и т.д.).

---

### Слайс 2 — имя компьютера и учётные записи

**Статус: выполнено.**

**Файлы:** `internal/xmlgen/components/shellsetup.go`, правила в `internal/profile/validate.go` для `ComputerName`, `Accounts`, `FirstLogon`.

**Что делает:** `BuildAnswerFile` добавляет `Microsoft-Windows-Shell-Setup` (проход `specialize`) с `ComputerName`, и (проход `oobeSystem`) с `UserAccounts/LocalAccounts` из `profile.Accounts` и `AutoLogon`, заполненным по `profile.FirstLogon`.

**Критерии приёмки:**

1. `ComputerName` задан — попадает в XML дословно. `ComputerName == nil` — элемент `ComputerName` отсутствует (Windows сгенерирует случайное имя сама).
2. Каждый аккаунт из `Accounts` даёт один элемент `LocalAccount` с верными `Name`, `DisplayName` (или без него, если `nil`), `Group`.
3. `FirstLogon.Mode == FirstLogonBuiltinAdmin` — в XML активируется встроенная учётная запись Administrator с заданным паролем, автологон настроен на неё.
4. `FirstLogon.Mode == FirstLogonFirstCreatedAccount` — автологон настроен на первый элемент `Accounts`.
5. `FirstLogon.Mode == FirstLogonNone` — элемент `AutoLogon` отсутствует.
6. Компьютерное имя длиннее 15 символов — `ValidateProfile` отклоняет профиль с сообщением `Имя компьютера длиннее 15 символов`.

**Тесты (`internal/xmlgen/builder_accounts_test.go`, дополнения в `validate_test.go`):** по критериям 1–6.

**Не делать:** обфускацию паролей Base64, PowerShell-скрипт для динамического имени компьютера, домены Active Directory.

---

### Слайс 3 — CLI end-to-end

**Файлы:** `internal/cli/generate.go`, доработка `validate.go` и `profile.go` (пресетов пока нет — только значения по умолчанию).

**Что делает:** `unattend-gen generate <profile.json> [-o PATH]` читает профиль, валидирует, при успехе пишет `autounattend.xml` и печатает путь; при ошибке ничего не пишет и печатает список ошибок.

**Критерии приёмки:**

1. `generate` на валидном профиле создаёт файл, парсящийся как XML, с корректным кодом выхода 0.
2. `generate` на невалидном профиле (например, `custom_key` без `product_key`) не создаёт файл, код выхода 1, ошибки в stderr.
3. `generate` без `-o` пишет `autounattend.xml` рядом с файлом профиля.
4. `generate` с несуществующим путём к профилю — понятное сообщение об ошибке, код выхода 1, не Go-паника и не стектрейс.

**Тесты (`internal/cli/generate_test.go`):** по критериям 1–4, через вызов `Execute()` с подменённым выводом и `t.TempDir()`.

**Не делать:** TUI, прогресс-бар, интерактивные подтверждения перезаписи файла.

---

### Слайс 4 — express settings и системные твики

**Файлы:** `internal/xmlgen/components/shellsetup.go` (расширение oobeSystem), `internal/xmlgen/components/setup.go` (bypass-требования Windows 11 добавляются в проход `windowsPE`), правила валидации не меняются — поля уже bool/enum, формальной валидации сверх типов не требуют.

**Что делает:** `ExpressSettings.Mode` управляет блоком телеметрии OOBE; `SystemTweaks.DisableWindowsUpdate`, `DisableUAC`, `BypassWin11Requirements` добавляют соответствующие элементы (для Windows Update — синхронная команда реестра; для UAC и bypass — реестровые правки через `RunSynchronousCommand` в `specialize`).

**Критерии приёмки:**

1. `ExpressSettings.Mode == ExpressAllDisabled` отключает все категории телеметрии в OOBE.
2. `ExpressSettings.Mode == ExpressAllEnabled` включает все категории.
3. `ExpressSettings.Mode == ExpressInteractive` — блок OOBE express settings в XML отсутствует, Windows Setup спросит сама.
4. Каждый включённый флаг `SystemTweaks` добавляет ровно одну команду `RunSynchronousCommand` с ожидаемым содержимым; выключенный флаг не добавляет ничего.
5. Все три флага выключены — блока системных твиков в XML нет вообще, а не пустой контейнер.

**Тесты (`internal/xmlgen/builder_tweaks_test.go`):** по критериям 1–5.

**Не делать:** остальные системные твики из исходного сервиса (SmartScreen, Fast Startup, ACL и т.д.) — добавляются отдельным контрактом при необходимости, не сверх раздела 4.

---

### Слайс 5 — TUI и пресеты

Слайс не меняет `Profile`, `validate.go` и `xmlgen/*`. Из API используется только то, что уже описано в разделе 5.

**Файлы:** `internal/tui/app.go`, все файлы `internal/tui/screens/*`, все файлы `internal/tui/widgets/*` из раздела 6, `presets/minimal.json`, `presets/single-user.json`, доработка `internal/cli/profile.go` (`--preset`), `internal/cli/tui.go`.

**Что делает:** экраны из раздела 7 в указанном порядке, состояние — единый `*profile.Profile` в корневой `Model`, экран `review` показывает XML через `xmlgen.BuildAnswerFile` и позволяет сохранить `.json` и/или `.xml`.

**Критерии приёмки:**

1. Заполнение всех экранов и сохранение на `review` даёт тот же XML, что `unattend-gen generate` на эквивалентном JSON-профиле.
2. Переход «Назад» не теряет уже введённые на предыдущих экранах данные.
3. `unattend-gen profile init demo --preset minimal` создаёт профиль, идентичный `presets/minimal.json` с заменённым `name`.
4. Невалидное значение на экране (например, слишком длинное имя компьютера) не даёт перейти на `review`, ошибка показана рядом с полем.

**Тесты:** для TUI — `internal/tui/app_test.go` через `teatest` на критерии 1 и 4 (программный ввод в поля через `tea.KeyMsg`, снятие состояния корневой `Model`, не пиксельный снапшот экрана). Для пресетов — `internal/profile/store_test.go`, критерий 3.

**Не делать:** тему оформления, мышь (только клавиатура), сохранение истории между запусками TUI сверх текущего профиля.

---

### Слайс 6 — Wi-Fi

Слайс не трогает `Accounts`, `Language`, `Edition` и их компоненты.

**Файлы:** `internal/xmlgen/components/wlan.go`, правила в `internal/profile/validate.go` для `Wifi`, экран `internal/tui/screens/wifi.go` (уже создан в слайсе 5 как заглушка — здесь получает логику).

**Что делает:** `Wifi != nil` добавляет компонент `Microsoft-Windows-WLANSVC` (проход `oobeSystem`) с профилем сети: SSID, тип аутентификации, пароль (в открытом виде — обфускация паролей вне v1), видимость сети.

**Критерии приёмки:**

1. `Wifi == nil` — компонент WLAN в XML отсутствует, оборудование настраивается вручную или через сервисный `$WinPEDriver$`.
2. `Authentication == WifiOpen` — элемент пароля в XML отсутствует.
3. `Authentication` в `{WifiWPA2Personal, WifiWPA3Personal}` без пароля короче 8 символов — `ValidateProfile` отклоняет профиль.
4. `ConnectHidden == true` — сеть помечена как скрытая (`nonBroadcast`), иначе — нет.

**Тесты (`internal/xmlgen/builder_wifi_test.go`):** по критериям 1–4.

**Не делать:** импорт XML-профиля Wi-Fi из `netsh`, множественные сети, корпоративную аутентификацию (WPA2/WPA3-Enterprise, 802.1X).

---

## 16. Ревью после каждого слайса

Отдельным заходом, после того как слайс готов и гейт зелёный. Задача захода — искать проблемы, а не хвалить написанное.

Чек-лист:

1. Построение XML (`encoding/xml`) только в `internal/xmlgen/`. В `internal/cli/` и `internal/tui/` его нет.
2. Имена полей совпадают с разделом 4. Ни одного придуманного поля вроде `computerName` вместо `ComputerName` или `wifiSsid` вместо `Wifi.SSID`.
3. Обхода `Accounts`/других срезов с побочными эффектами на диск в цикле нет (например, `SaveProfile` внутри цикла по аккаунтам).
4. Обработки ошибок, которых не бывает, нет: `if err != nil` вокруг кода, где `err` всегда `nil` по конструкции; лишние проверки `nil` там, где указатель уже гарантированно не `nil` по инварианту вызова.
5. Тесты проверяют критерии приёмки, а не повторяют реализацию. На каждый критерий с отказом есть тест на отказ.
6. Использованы примитивы из `internal/tui/widgets`, самописных полей на экранах нет.
7. Мёртвого кода нет: неиспользуемые экспорты, экраны, XML-компоненты без вызова (`golangci-lint` с `unused` должен был это поймать в гейте — ревью проверяет, что линтер не отключён точечным `//nolint`).
8. Файлов и абстракций сверх раздела 3 не появилось.

Находки правятся в том же заходе, потом гейт прогоняется заново.
