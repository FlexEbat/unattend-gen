# unattend-gen — 使用指南

[English](USAGE.md) | [Русский](USAGE.ru.md) | 简体中文 | [Español](USAGE.es.md) | [हिन्दी](USAGE.hi.md)

本文是 `unattend-gen` 的完整使用手册。它是一个用 Go 编写的 CLI/TUI 工具，用于生成
Windows 10/11 的 `autounattend.xml` 应答文件。你只需创建一次**配置文件**（profile，
一个描述 Windows 安装与配置方式的 JSON 文件），之后每次重装系统都可以据此生成应答文件。
无需网页表单，无需联网，工具本身是单个静态二进制文件。

功能概览请见 [README.zh-CN.md](../README.zh-CN.md)。本指南涵盖每条命令、每个 TUI 界面，
以及配置文件中可以出现的每个字段。

如果本文与生成的 XML 不一致，以代码为准。本文由人工依据工具自身使用的同一事实来源
（`internal/profile/schema.go`）编写，难免有疏漏。发现问题请提交 issue。

## 目录

- [本工具做什么，以及刻意不做什么](#what-it-does)
- [安装与构建](#installing)
- [快速开始](#quick-start)
- [CLI 参考](#cli-reference)
- [逐屏介绍 TUI](#tui-screens)
- [配置文件 JSON 参考](#profile-json-reference)
- [内置预设](#built-in-presets)
- [配置文件示例](#example-profiles)
- [验证生成的应答文件](#verifying)
- [故障排查与已知限制](#limitations)

<a id="what-it-does"></a>
## 本工具做什么，以及刻意不做什么

`unattend-gen` 生成单个 `autounattend.xml` 文件。把它放在 U 盘（或安装介质根目录）中，
与 Windows 安装程序放在一起，Windows 安装程序会自动运行它，并跳过你已经预先配置好的所有提示。

**它不会对磁盘分区或格式化。** Windows 安装程序始终会以交互方式询问安装到哪个磁盘，
这是有意为之，并非功能缺失。本工具负责的是你选定磁盘或分区*之后*的一切：语言、账户、
系统调整、应用移除、个性化等等。

**运行时不需要网络。** 没有网页表单，没有服务器：它只是一个本地二进制文件，读取 JSON
配置文件并写出 XML 文件。

<a id="installing"></a>
## 安装与构建

从源码构建需要 Go 1.23+（本指南不提供预编译二进制；请查看仓库的 Releases 页面是否有适合你平台的版本）。

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

这会生成单个静态二进制文件 `unattend-gen`（Windows 上为 `unattend-gen.exe`），没有任何运行时依赖。

<a id="quick-start"></a>
## 快速开始

最快的方式是使用交互式 TUI：

```sh
./unattend-gen tui
```

用 **Ctrl+N**/**Esc** 在各界面间切换，按 **Ctrl+R** 可直接跳到摘要界面，然后保存。
这会写出一个 JSON 配置文件，之后可以据此重新生成，也可以手工编辑。

如果你更愿意直接操作 JSON：

```sh
# 用默认设置（大部分为交互式）创建新的配置文件
./unattend-gen profile init my-pc

# ...手工编辑 my-pc.json，或继续在 TUI 中编辑：
./unattend-gen tui my-pc.json

# 检查是否有效
./unattend-gen validate my-pc.json

# 生成应答文件（默认把 autounattend.xml 写在配置文件
# 旁边）
./unattend-gen generate my-pc.json
```

把得到的 `autounattend.xml` 复制到 Windows 安装 U 盘的根目录（与 `setup.exe` 放在一起；
如果制作自定义介质，则放在 ISO 根目录），然后从它启动：Windows 安装程序会自动找到并应用它。

<a id="cli-reference"></a>
## CLI 参考

每条命令都会先校验输入，并在出错时明确失败（非零退出码，错误信息写入 stderr），
而不会生成不完整或凭猜测得到的文件。

### `profile init <name> [--preset minimal|single-user]`

在当前目录创建 `<name>.json`。不带 `--preset` 时，配置文件由内置默认值生成
（语言/区域/键盘为 `en-US`，其余项基本都是“安装时再问我”；见下文[默认值](#defaults)）。
带 `--preset` 时，从两个内置预设之一开始（见[内置预设](#built-in-presets)），
且 `name` 始终是你在命令行中传入的值，与预设文件里写的无关。

```sh
./unattend-gen profile init laptop --preset single-user
# -> laptop.json
```

### `profile list`

列出 `./profiles` 中的每个 `*.json` 文件（每行一个路径，别无其他内容，输出便于脚本处理）。
该目录必须存在并包含配置文件；配置文件的存放位置没有其他设置。

```sh
./unattend-gen profile list
```

### `validate <profile.json>`

执行与 `generate` 相同的校验，但不生成应答文件。每个错误一行输出到 stderr，
有错误时以非零退出码结束；配置文件有效时输出 `профиль корректен`（校验消息和这行成功提示均为俄语，
见下文[语言约定](#text-conventions)），退出码为 0。

```sh
./unattend-gen validate laptop.json
```

### `generate <profile.json> [-o path]`

校验配置文件，然后构建并写出应答文件。不带 `-o`/`--output` 时，文件以 `autounattend.xml`
为名写入配置文件所在目录。成功时输出所写入的路径。

```sh
./unattend-gen generate laptop.json
./unattend-gen generate laptop.json -o /media/usb/autounattend.xml
```

### `tui [profile.json]`

打开交互式终端界面。不带参数时从内置默认值开始。给出配置文件路径时，会先加载它
（与 `validate`/`generate` 使用同一个加载器），因此可以来回迭代：在 TUI 中编辑、保存、
手工微调 JSON、再在 TUI 中打开，如此往复。

```sh
./unattend-gen tui
./unattend-gen tui laptop.json
```

<a id="text-conventions"></a>
**关于语言**：代码、注释、提交信息以及普通控制台文本（命令帮助、通过管道输出时看到的成败消息）
使用英语。你填写配置文件时 TUI 显示的文字，以及所有校验错误消息，使用俄语。
这是项目约定，并非缺陷。

<a id="tui-screens"></a>
## 逐屏介绍 TUI

界面按以下顺序出现；**Ctrl+N** 前往下一个，**Esc** 返回，**Ctrl+R** 从任意位置直接跳到 Review，
**Tab** / **Shift+Tab** 在界面内切换字段，**空格**切换复选框。每个界面都会与同一份内存中的配置文件同步，
因此来回切换不会丢失任何内容。

1. **Welcome** — 欢迎界面，无需配置。
2. **Language** — 界面语言 / 区域 / 键盘布局（BCP-47 代码）、Windows 版本模式（交互式 / 通用密钥 /
   自定义密钥 / 存储在 BIOS-UEFI 固件中的密钥）、单独的仅用于激活的产品密钥，以及目标处理器架构。
3. **Accounts** — 计算机名（留空则由 Windows 生成）、时区、最多 5 个本地账户
   （名称/显示名/密码/组，在可编辑表格中），首次登录行为，以及尽力而为的“跳过要求 Microsoft 账户”复选框。
4. **Tweaks** — 快速设置（遥测/诊断），以及全部 33 项系统调整（见下文[完整列表](#system-tweaks)）、
   密码过期与账户锁定策略、文件资源管理器调整。它们都是“翻转某个默认值”类设置，所以放在同一界面。
5. **Wifi** — 配置首次启动时自动连接的 Wi-Fi 网络：可以填写 SSID/安全类型/密码/是否隐藏，
   也可以粘贴导出的原始 WLAN 配置文件 XML。
6. **Apps** — 三组复选框：要移除的应用（Appx 包）、要移除的 Windows 功能（DISM 功能包），
   以及要移除的旧版可选功能（第三种独立的移除机制）；见下文[完整列表](#remove-apps)。
7. **Personalization** — 浅色/深色主题（系统与应用分别设置）、强调色、强调色显示位置
   （开始菜单/任务栏、标题栏）、透明效果，以及纯色桌面背景。
8. **Accessibility** — 粘滞键（关闭/禁用/自定义标志组合），以及 Caps/Num/Scroll Lock
   的初始状态和按下它们是否有作用。
9. **Desktop** — 显示哪些桌面图标（此电脑、回收站等，共 13 个），以及在开始菜单电源按钮旁固定哪些特殊文件夹（Windows 11）。
10. **VisualEffects** — Windows 的“性能选项”：一个预设（最佳外观 / 最佳性能）或 17 个单独的动画/外观开关。
11. **Taskbar** — 开始菜单与任务栏调整：禁用小组件、任务栏左对齐（Windows 11）、隐藏任务视图按钮、
    禁用搜索中的 Bing 结果、始终显示所有托盘图标、任务栏搜索框的显示模式、开始菜单固定项
    （Windows 11 JSON）/ 磁贴（Windows 10 XML），以及固定的任务栏图标（空白或自定义布局 XML）。
12. **Advanced** — 安装虚拟机来宾工具（VirtualBox/VMware/VirtIO/Parallels）、原始 AppLocker 策略 XML、
    用于动态计算计算机名的 PowerShell 脚本，以及三个小复选框：安装后保留应答文件而不是删除、
    自动启动讲述人、在生成的 XML 中混淆账户密码。
13. **Scripts** — 每个类别一个自定义脚本（System/DefaultUser/FirstLogon/UserOnce，见下文[自定义脚本](#custom-scripts)），
    在多行文本框中编辑，另有一个复选框控制脚本运行后是否重启资源管理器。
14. **Review** — 实时更新的整个配置文件摘要，如果校验通过，还会显示生成的 XML 预览。

并非每个界面上的每个字段都与顶层 JSON 键一一对应：有些界面编辑的是嵌套对象
（例如 `system_tweaks`、`personalization`）。下面的[配置文件 JSON 参考](#profile-json-reference)
按 JSON 结构组织，并注明各部分由哪个界面编辑，所以你可以从任一方向找到字段。

<a id="profile-json-reference"></a>
## 配置文件 JSON 参考

配置文件是一个 JSON 对象，`schema_version: 1`。下面每个字段除非标注为**必填**，否则都是可选的；
省略可选字段（或设为 `null`、将布尔值保持为 `false`）表示“保留 Windows 自身的默认行为”，
只有一个刻意的例外，会在出现处说明。

### 顶层

| 字段 | 类型 | 说明 |
|---|---|---|
| `schema_version` | int | **必填**，必须为 `1`。 |
| `name` | string | **必填**。自由文本：不会写入 XML，只是配置文件本身的标签。 |

### 语言与版本 — *(Language 界面)*

```json
"language": {
  "ui_language": "en-US",
  "locale": "en-US",
  "keyboard_layout": "en-US"
}
```

三个字段都**必填**，均为 BCP-47 代码（如 `en-US`、`de-DE`、`ru-RU`）。`keyboard_layout` 对应输入区域设置；
`locale` 同时设置系统区域和用户区域。

```json
"edition": {
  "mode": "interactive",
  "edition": null,
  "product_key": null
}
```

- `mode`，取值之一：
  - `"interactive"` — Windows 安装程序在安装时询问版本与密钥。
  - `"generic_key"` — 要求 `edition` 为 `"Home"`、`"Pro"`、`"Education"`、`"Enterprise"` 之一；
    安装程序使用 Microsoft 公开的该版本通用密钥（KMS 客户端安装密钥），之后仍需要你自行激活。
  - `"custom_key"` — 要求提供 `product_key`，即真实的 25 位密钥，格式为 `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX`。
    除非你显式设置了下面的 `activation_key`，否则该密钥之后也会用于激活。
  - `"firmware"` — 使用设备 BIOS/UEFI 固件中已嵌入的产品密钥（OEM 预装 Windows 的常见情况）；
    不会请求或写入任何密钥。

```json
"activation_key": null,
"processor_architecture": ""
```

- `activation_key`：单独的、**仅**用于激活的密钥（`Microsoft-Windows-Shell-Setup/ProductKey`），
  与 `edition` 用来选择安装版本的密钥（如有）无关。保持 `null` 时，会回退到 `edition.product_key`
  （仅当 `edition.mode` 为 `"custom_key"` 时），否则不使用任何密钥激活。
- `processor_architecture`：`"amd64"`（为空或缺省时的默认值）、`"x86"`、`"arm64"` 之一。
  每个配置文件只支持一种架构，与某些其他 unattend 生成器不同，本工具不会构建一个适用于多种架构的应答文件。

### 计算机名与时区 — *(Accounts 界面)*

```json
"computer_name": null,
"computer_name_script": null,
"timezone": null
```

- `computer_name`：静态主机名（1–15 个字符，字母/数字/连字符，不能以连字符开头或结尾，也不能全为数字）。
  `null` 表示由 Windows 随机生成。
- `computer_name_script`：安装期间运行的 PowerShell 脚本，其输出将成为计算机名，适合动态生成名称
  （例如根据序列号或命名规则）。与 `computer_name` **互斥**：两者同时设置会产生校验错误。
  重命名由一个后台进程完成，它会在安装结束后的一小段时间内反复应用该名称，
  以绕过 Windows 在 specialize 阶段自行改写名称的问题。
- `timezone`：Windows 时区 ID 字符串，例如 `"Russian Standard Time"`、`"Pacific Standard Time"`、`"UTC"`。
  `null` 表示由 Windows 自动检测。工具不会校验字符串是否对应真实时区（列表很长且随版本变化），
  只检查：只要设置了它，就不能是空字符串。

### 账户与首次登录 — *(Accounts 界面)*

```json
"accounts": [
  {
    "name": "alice",
    "display_name": null,
    "password": "Sup3rSecret!",
    "group": "Administrators"
  }
],
"first_logon": { "mode": "first_created_account" }
```

- `accounts`：最多 5 项。
  - `name` — **必填**，≤20 个字符，不得包含 `" / \ [ ] : ; | = , + * ? < >`。
  - `display_name` — 可选的友好名称。
  - `password` — `null` 表示无密码；空字符串无效（请改用 `null`）。
  - `group` — **必填**，`"Administrators"` 或 `"Users"`。
- `first_logon.mode`，取值之一：
  - `"none"` — 不自动登录；首次真正启动时显示常规登录界面。
  - `"first_created_account"` — 自动登录到 `accounts` 中的第一个账户（该列表必须非空）。
  - `"builtin_administrator"` — 自动登录到隐藏的内置 Administrator 账户；要求设置
    `first_logon.builtin_administrator_password`。

### 快速设置与绕过选项

```json
"express_settings": { "mode": "interactive" },
"bypass_online_account_requirement": false
```

- `express_settings.mode`：`"interactive"`（安装程序询问遥测等问题）、`"all_enabled"` 或 `"all_disabled"`。
  任何非 `"interactive"` 的值还会自动隐藏若干 OOBE 界面（EULA、OEM 注册、网络设置）。
- `bypass_online_account_requirement`：尽力而为地让安装程序在使用本地账户的情况下完成，
  而不要求登录 Microsoft 账户（写入众所周知的 `BypassNRO` 注册表值）。**不保证有效**：
  Microsoft 在 2025–2026 年间不止一次堵上了这个口子，在未于当前 Windows 版本上测试之前，
  不要在批量部署中依赖它。只要 `accounts` 中至少有一项，本地账户无论是否设置此标志，都能可靠地跳过该提示。

<a id="system-tweaks"></a>
### 系统调整 — *(Tweaks 界面)*

`"system_tweaks": { ... }` 内共有 33 个字段，每个都是普通布尔值，默认为 `false`（不更改），
**唯一例外**是下面说明的 `keep_sensitive_files`：

| 字段 | 作用 |
|---|---|
| `disable_windows_update` | 暂停/禁用 Windows 更新。 |
| `disable_uac` | 禁用用户账户控制（UAC）提示。 |
| `bypass_win11_requirements` | 跳过 TPM/安全启动/内存硬件检查（在 windowsPE 阶段、安装程序评估它们之前）。 |
| `disable_smart_app_control` | 禁用智能应用控制。 |
| `disable_smart_screen` | 禁用 SmartScreen（系统与 Edge）。 |
| `disable_fast_startup` | 禁用快速启动（混合启动）。 |
| `disable_system_restore` | 禁用系统还原。 |
| `enable_long_paths` | 启用超过 260 个字符的 NTFS 长路径支持。 |
| `enable_remote_desktop` | 启用远程桌面及防火墙规则。 |
| `allow_powershell_scripts` | 将 PowerShell 执行策略设为 `RemoteSigned`。 |
| `disable_last_access_timestamp` | `fsutil behavior set disablelastaccess 1`。 |
| `prevent_device_encryption` | 阻止自动 BitLocker 设备加密。 |
| `disable_auto_sign_on_last_user` | 禁用重启后自动登录上一个交互用户。 |
| `disable_wpbt` | 禁用 Windows Platform Binary Table 的执行。 |
| `audit_process_creation` | 启用进程创建审核，包括命令行。 |
| `hide_edge_first_run` | 跳过 Edge 的首次运行体验。 |
| `disable_edge_startup_boost` | 禁用 Edge 启动加速/后台模式。 |
| `delete_hidden_junctions` | 删除旧版 NTFS 联接点（例如 `C:\Documents and Settings`），对安装账户和所有未来账户生效。 |
| `prevent_automatic_reboot` | 阻止 Windows 更新重启正在使用的计算机（注册一个计划任务，不断把“活动时间”调整到当前时间）。 |
| `turn_off_system_sounds` | 将声音方案设为“无声音”，对安装账户和所有未来账户生效。 |
| `disable_app_suggestions` | 禁用 Content Delivery Manager 悄悄安装的推荐应用。 |
| `disable_pointer_precision` | 禁用“提高指针精确度”（鼠标加速）。 |
| `prevent_device_apps` | 阻止 Windows 下载/安装与特定硬件设备关联的应用。 |
| `harden_system_drive_acl` | 移除“Authenticated Users”组对 `C:\` 的写入权限。 |
| `make_edge_uninstallable` | 翻转内部策略标志，使 Edge 显示“卸载”选项。 |
| `delete_windows_old` | 删除 `C:\Windows.old`（仅与原地升级相关；全新安装时是无害的空操作）。 |
| `disable_core_isolation` | 禁用内存完整性/基于虚拟化的安全（在某些虚拟机来宾中或使用旧驱动时有用）。 |
| `delete_edge_desktop_icon` | 删除 Microsoft Edge 桌面快捷方式，对安装账户和所有未来账户生效。 |
| `disable_widgets` | 禁用小组件面板。 |
| `left_taskbar` | 任务栏左对齐（Windows 11；默认居中）。 |
| `hide_task_view_button` | 隐藏任务栏上的任务视图按钮。 |
| `disable_bing_results` | 禁用任务栏搜索中出现的 Bing 网页结果。 |
| `show_all_tray_icons` | 始终显示通知区域的所有图标，而不是折叠不活动的图标（Windows 10 和 11 的机制不同，已自动处理）。 |

`keep_sensitive_files` 位于配置文件的顶层，而不在 `system_tweaks` 内，请看紧接着的说明，
因为它的默认行为是整个模式中“false/缺省表示不更改”这一规则的唯一刻意例外。

```json
"keep_sensitive_files": false
```

默认情况下（`false`，即 **JSON 中缺省** 时），工具会在安装完成后删除
`C:\Windows\Panther\unattend.xml` / `unattend-original.xml`（Windows 安装程序保留的应答文件副本，
除非你同时设置了 `obscure_passwords`，否则其中包含明文密码）以及它自己的 Wi-Fi 配置文件临时文件。
设为 `true` 则保留这些文件。这是整个模式中唯一一个默认值*会执行某些操作*而不是*不更改任何内容*的字段，
权衡的结果是：安装后明文密码留在磁盘上，比打破惯例更糟。

### 密码过期与账户锁定 — *(Tweaks 界面)*

```json
"password_expiration": { "mode": "default", "days": null },
"account_lockout": {
  "mode": "default",
  "threshold": null,
  "window_minutes": null,
  "duration_minutes": null
}
```

- `password_expiration.mode`：`"default"`（Windows 默认值，42 天，不发出任何命令）、`"never"`（密码永不过期）
  或 `"custom"`（要求 `days` ≥ 1）。
- `account_lockout.mode`：`"default"`（Windows 默认值：10 次失败尝试 / 10 分钟窗口 / 锁定 10 分钟）、
  `"disabled"`（关闭锁定）或 `"custom"`（要求三个数值字段全部提供，且各自 ≥ 1）。

### 文件资源管理器调整 — *(Tweaks 界面)*

```json
"file_explorer": {
  "hidden_files": "default",
  "show_file_extensions": false,
  "classic_context_menu": false,
  "hide_folder_tooltips": false,
  "open_to_this_pc": false,
  "show_end_task_in_taskbar": false
}
```

- `hidden_files`：`"default"`、`"show_hidden"`（显示隐藏文件）或 `"show_all"`（显示隐藏文件和受保护的操作系统文件）。
- 其余布尔值：显示文件扩展名；在 Windows 11 上恢复经典（Windows 10 风格）右键菜单；隐藏文件夹/桌面图标的提示；
  让文件资源管理器打开到“此电脑”而不是“快速访问”/“主页”；在任务栏右键菜单中直接显示“结束任务”。

所有这些设置既适用于安装期间创建的账户，也适用于机器上的所有未来账户。

### Wi-Fi — *(Wifi 界面)*

```json
"wifi": {
  "ssid": "MyNetwork",
  "authentication": "WPA2Personal",
  "password": "hunter2000",
  "connect_hidden": false,
  "raw_profile_xml": null
}
```

`wifi` 整体是可选的：省略它（或设为 `null`）则完全不配置 Wi-Fi。

- `authentication`：`"Open"`、`"WPA2Personal"` 或 `"WPA3Personal"`。除非 `authentication` 为 `"Open"`，
  否则必须提供 `password`（≥8 个字符）。
- `raw_profile_xml`：设置后，将原样使用这份原始 WLAN 配置文件 XML（通过 `netsh wlan export profile key=clear` 导出），
  而不是用 `ssid`/`authentication`/`password`/`connect_hidden` 构建；这种情况下这四个字段变为可选。

<a id="remove-apps"></a>
### 移除应用、功能与可选功能 — *(Apps 界面)*

三个独立的列表，使用三种不同的底层移除机制，之所以分开，是因为一个列表里的名称在另一个列表里无效。

```json
"remove_apps": ["OneDrive", "Terminal", "Store"],
"remove_features": ["InternetExplorer"],
"remove_optional_features": ["Recall"]
```

**`remove_apps`**（Appx 包，通过 `Remove-AppxProvisionedPackage` 移除），可取：

`3DViewer`, `BingSearch`, `Calculator`, `Camera`, `Clipchamp`, `Clock`,
`Copilot`, `Cortana`, `DevHome`, `Family`, `FeedbackHub`, `GameAssist`,
`GetHelp`, `MailAndCalendar`, `Maps`, `MediaPlayerModern`, `MixedReality`,
`MoviesAndTV`, `News`, `Notepad`, `Office`, `OneDrive`, `OneNote`,
`Outlook`, `Paint`, `Paint3D`, `People`, `PhoneLink`, `PowerAutomate`,
`QuickAssist`, `Skype`, `SnippingTool`, `SolitaireCollection`,
`StickyNotes`, `Store`, `Teams`, `Terminal`, `Tips`, `ToDo`,
`VoiceRecorder`, `Wallet`, `Weather`, `XboxApps`.

`OneDrive` 在底层的处理方式与其余项不同（它不是以 Appx 包分发的：工具会删除其残留的快捷方式和安装可执行文件，
并移除其自启动项），但你的用法相同：在这同一个列表里按名称指定即可。

**`remove_features`**（Windows 可选*功能包*，通过 `Remove-WindowsCapability` 移除），可取：
`InternetExplorer`、`WordPad`、`PowerShellISE`、`OpenSSHClient`、`MediaPlayer`、`Speech`、`Handwriting`、
`WindowsHello`、`MathInputPanel`、`OneSync`、`StepsRecorder`。

**`remove_optional_features`**（旧版 Windows 可选功能，通过 `Disable-WindowsOptionalFeature` 移除，
是第三种独立机制），可取：`Recall`、`MediaFeatures`、`RemoteDesktopClient`。

### 个性化 — *(Personalization 界面)*

```json
"personalization": {
  "system_theme": "",
  "apps_theme": "",
  "accent_color": null,
  "show_accent_on_start_taskbar": false,
  "show_accent_on_title_bars": false,
  "disable_transparency": false,
  "solid_color_wallpaper": null
}
```

- `system_theme` / `apps_theme`：`"light"` 或 `"dark"`；为空/缺省则保留 Windows 默认值。
- `accent_color` / `solid_color_wallpaper`：6 位十六进制颜色字符串（如 `"FF8800"`），不带 `#` 前缀。
  `solid_color_wallpaper` 用纯色替换桌面背景；不支持使用图片文件作为壁纸。
- 这些设置对所有未来账户生效，而不只是安装期间创建的账户。

### 辅助功能 — *(Accessibility 界面)*

```json
"sticky_keys": { "mode": "default", "flags": [] },
"lock_keys": null
```

- `sticky_keys.mode`：`"default"`（Windows 默认值：开启，按 5 次 Shift 激活）、`"disabled"`
  （关闭 5×Shift 快捷方式本身，而不只是视觉附加效果）或 `"custom"`（使用 `flags`，可为
  `HotKeyActive`、`Indicator`、`TriState`、`TwoKeysOff`、`AudibleFeedback`、`HotKeySound` 的任意子集）。
- `lock_keys`：`null` 不改动 Caps/Num/Scroll Lock 的行为。设置对象则一次性配置这三个键：

  ```json
  "lock_keys": {
    "caps_lock":   { "initial": "off", "behavior": "toggle" },
    "num_lock":    { "initial": "on",  "behavior": "toggle" },
    "scroll_lock": { "initial": "off", "behavior": "ignore" }
  }
  ```

  `initial` 为 `"off"` 或 `"on"`（Windows 启动后立即的状态）；`behavior` 为 `"toggle"`（正常）或
  `"ignore"`（按下该键完全不起作用，重启后生效）。

### 桌面图标与“开始”文件夹 — *(Desktop 界面)*

```json
"desktop_icons": { "ThisPC": true, "RecycleBin": false },
"start_folders": ["Settings", "Documents", "Downloads"]
```

- `desktop_icons`：一个映射；你没有提到的键保持 Windows 默认值，提到的键被显式设为 `true`（显示）或 `false`（隐藏）。
  有效键：`ThisPC`、`UserFiles`、`Network`、`RecycleBin`、`ControlPanel`、`Desktop`、`Documents`、`Downloads`、
  `Music`、`Pictures`、`Videos`、`Gallery`、`Home`。
- `start_folders`：固定在 Windows 11 开始菜单电源按钮旁的特殊文件夹的有序列表，列表顺序即固定顺序。
  有效值：`Settings`、`FileExplorer`、`Documents`、`Downloads`、`Music`、`Pictures`、`Videos`、`Network`、`PersonalFolder`。
  空列表或缺省表示“保留 Windows 自带的默认集合”，用这个字段无法表达“一个文件夹都不固定”。

两者都对所有未来账户生效，而不只是安装期间创建的账户。

### 视觉效果 — *(VisualEffects 界面)*

```json
"visual_effects": { "mode": "default", "custom": {} }
```

- `mode`：`"default"`（不更改）、`"best_appearance"`（所有效果开启）、`"best_performance"`（所有效果关闭）
  或 `"custom"`（使用 `custom`）。
- `custom`：效果名称到 `true`/`false` 的映射；未提到的效果保持 Windows 默认值。有效键：`ControlAnimations`、
  `AnimateMinMax`、`TaskbarAnimations`、`DWMAeroPeekEnabled`、`MenuAnimation`、`TooltipAnimation`、`SelectionFade`、
  `DWMSaveThumbnailEnabled`、`CursorShadow`、`ListviewShadow`、`ThumbnailsOrIcon`、`ListviewAlphaSelect`、
  `DragFullWindows`、`ComboBoxAnimation`、`FontSmoothing`、`ListBoxSmoothScrolling`、`DropShadow`。

对所有未来账户生效，而不只是安装期间创建的账户。

### 开始菜单与任务栏 — *(Taskbar 界面)*

```json
"taskbar_search": "",
"start_pins": { "mode": "default", "json": null },
"start_tiles": { "mode": "default", "xml": null },
"taskbar_icons": { "mode": "default", "xml": null }
```

- `taskbar_search`：`""`/缺省（Windows 默认，显示搜索框）、`"hide"`、`"icon"`（仅图标，无搜索框）、
  `"box"`（显式指定，与默认相同）或 `"label"`（带文字标签的图标）。
- `start_pins`（仅 Windows 11，对 Windows 10 无效）：`mode` 为 `"default"`、`"empty"`（不固定任何应用）
  或 `"custom"`（要求 `json`：符合 Windows `ConfigureStartPins` 策略所需格式的原始 `{"pinnedList": [...]}` 内容）。
- `start_tiles`（仅 Windows 10，对 Windows 11 无效）：`mode` 为 `"default"`、`"empty"`（没有磁贴组）
  或 `"custom"`（要求 `xml`：原始 `LayoutModification.xml` 文档）。
- `taskbar_icons`：为所有未来账户固定到任务栏的图标。`mode` 为 `"default"`、`"empty"`（没有固定图标）
  或 `"custom"`（要求 `xml`：包含 `CustomTaskbarLayoutCollection` 的原始 `LayoutModification.xml`）。
  它通过锁定的开始布局实现，并在每个账户首次登录时解除锁定，因此用户之后可以自行调整任务栏；
  为此，生成的应答文件还会注册一个小型的 `UnlockStartLayout` 计划任务。对 XML 只检查格式是否良好，
  不检查是否符合 Windows 任务栏布局架构。

### 高级 — *(Advanced 界面)*

```json
"install_vm_guest_tools": [],
"applocker_policy_xml": null,
"use_narrator": false,
"obscure_passwords": false
```

- `install_vm_guest_tools`：可取 `VBoxGuestAdditions`、`VMwareTools`、`VirtIoGuestTools`、`ParallelsTools` 中的任意项。
  每一项运行一个静默安装程序，在 D–Z 盘符上查找对应的来宾工具 ISO，若未挂载则只写一条日志然后什么也不做，
  所以如果你事先不知道使用哪种虚拟机监控程序，把四项都列上“以防万一”也是安全的。
- `applocker_policy_xml`：原始 AppLocker 策略 XML。只检查是否为格式良好的 XML，不会对照 AppLocker 的完整架构进行验证：
  格式良好但无效的策略只会在目标机器上以错误的形式暴露。
- `use_narrator`：在安装过程中以及每个未来账户首次登录时自动启动讲述人。
- `obscure_passwords`：用 Base64 混淆生成的 XML 中的账户密码，而不是以明文写入。这是**混淆而不是加密**：
  盐是固定且公开记录的值（Microsoft 自己的 unattend 约定），所以它只能避免密码在原始文件中被简单 grep 到，仅此而已。

### 全局安装开关

```json
"hide_powershell_windows": false
```

为 `true` 时，本工具在安装期间启动的每个 PowerShell 窗口（内置调整和你自己的自定义脚本）都会隐藏运行，
而不是显示出来。这不影响基于 `cmd`/`reg`/`vbs` 的步骤，它们本来就不会显示窗口。

<a id="custom-scripts"></a>
### 自定义脚本 — *(Scripts 界面)*

```json
"system_scripts": [],
"default_user_scripts": [],
"first_logon_scripts": [],
"user_once_scripts": [],
"restart_explorer_after_scripts": false
```

共四个类别，每个都是 `{ "format": "...", "content": "..." }` 对象的列表。`format` 为 `cmd`、`ps1`、`reg`、`vbs` 之一
（`default_user_scripts` 不支持 `vbs`）。数量上限：`system_scripts` ≤4，`default_user_scripts` ≤3，
`first_logon_scripts` ≤4，`user_once_scripts` ≤4。

- **`system_scripts`** — 在系统上下文中运行一次，且在任何用户账户存在*之前*。
- **`default_user_scripts`** — 应用到默认用户注册表模板，因此会影响之后创建的每个账户，包括尚未存在的账户；
  安装期间创建的账户本身只有在它同样基于该模板创建时才会受影响（通常确实如此）。
- **`first_logon_scripts`** — 在第一个账户首次登录时运行一次。
- **`user_once_scripts`** — *每个账户*运行一次，包括未来的账户，在各自首次登录时执行。

`restart_explorer_after_scripts`：如果运行了任何 `first_logon_scripts`，之后重启资源管理器
（当脚本修改了资源管理器会缓存的内容时很有用）。

<a id="defaults"></a>
### `profile init`（不带预设）得到什么

`schema_version: 1`，语言/区域/键盘均为 `"en-US"`，`edition.mode: "interactive"`，没有账户，
`first_logon.mode: "none"`，`express_settings.mode: "interactive"`；其余全部为 JSON 零值
（空字符串/false/null/空列表），即对本指南描述的每项设置，都是“安装时问我”或“保留 Windows 默认值”。

<a id="built-in-presets"></a>
## 内置预设

二进制文件中内嵌了两个预设，通过 `profile init <name> --preset <preset-name>` 使用。

- **`minimal`** — 基本等同于不用预设：交互式选择版本，没有账户，交互式快速设置，三个原始系统调整显式设为 `false`。
  这是一个安全的、不额外做任何事的起点。
- **`single-user`** — 一个名为 `admin` 的本地管理员账户（未设置密码，真正使用前请添加），自动登录该账户
  （`first_logon.mode: "first_created_account"`），并且 `express_settings.mode: "all_disabled"`
  （完全跳过遥测/诊断提示）。适合只有你一个人使用的个人电脑，是合理的起点。

两个预设都早于上面记录的若干字段（它们是针对更早、更小的模式编写的）：没有提到的内容就取模式的零值默认值，
与其他地方未设置的字段相同。

<a id="example-profiles"></a>
## 配置文件示例

一个更完整的配置文件，综合了上面的几个部分：单用户笔记本，带一些隐私/性能调整，并移除了几个应用：

```json
{
  "schema_version": 1,
  "name": "laptop",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "interactive" },
  "computer_name": "LAPTOP-01",
  "timezone": "UTC",
  "accounts": [
    {
      "name": "alice",
      "display_name": "Alice",
      "password": "Sup3rSecret!",
      "group": "Administrators"
    }
  ],
  "first_logon": { "mode": "first_created_account" },
  "express_settings": { "mode": "all_disabled" },
  "bypass_online_account_requirement": true,
  "system_tweaks": {
    "disable_windows_update": false,
    "disable_uac": false,
    "bypass_win11_requirements": true,
    "disable_smart_screen": false,
    "enable_remote_desktop": true,
    "delete_hidden_junctions": true,
    "prevent_automatic_reboot": true,
    "left_taskbar": true,
    "hide_task_view_button": true
  },
  "file_explorer": {
    "hidden_files": "show_all",
    "show_file_extensions": true,
    "open_to_this_pc": true
  },
  "remove_apps": ["Cortana", "Skype", "Teams", "SolitaireCollection"],
  "remove_features": ["InternetExplorer"],
  "personalization": {
    "system_theme": "dark",
    "apps_theme": "dark",
    "accent_color": "0078D4"
  },
  "keep_sensitive_files": false,
  "obscure_passwords": true
}
```

一个简洁的无人值守最小安装（没有自己的账户，只跳过硬件检查并暂停更新），使用通用密钥：

```json
{
  "schema_version": 1,
  "name": "minimal-vm",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "generic_key", "edition": "Pro" },
  "accounts": [],
  "first_logon": { "mode": "none" },
  "express_settings": { "mode": "interactive" },
  "system_tweaks": {
    "bypass_win11_requirements": true,
    "disable_windows_update": true
  },
  "install_vm_guest_tools": ["VBoxGuestAdditions", "VMwareTools"]
}
```

<a id="verifying"></a>
## 验证生成的应答文件

`validate`/`generate` 只会根据模式检查 JSON 配置文件，无法确认 Windows 安装程序能否完整地接受生成的 XML，
因为这需要真正运行 Windows 安装程序。没有任何东西可以替代：把结果刻录到 U 盘（或在虚拟机中挂载），
亲眼看着它安装。如果 XML 有问题，Windows 安装程序通常会通过自己的错误对话框，或 `C:\Windows\Panther`
下的 `setupact.log`/`setuperr.log`（在 PE 环境中则是 `X:\Windows\Panther`）告诉你是哪个元素。

<a id="limitations"></a>
## 故障排查与已知限制

- **磁盘分区刻意不在范围内。** Windows 安装程序始终会停下来询问安装到哪个磁盘/分区。本工具配置的是围绕这一交互步骤的其余一切。
- **`bypass_online_account_requirement` 是尽力而为。** Microsoft 不止一次改变了它的工作方式；如果它在未来的
  Windows 版本上失效，那是 Microsoft 一侧的变化，而不是针对本工具的缺陷报告（不过仍欢迎你就此提交 issue）。
- **不支持多架构应答文件。** 每个配置文件请选择一个 `processor_architecture`。
- **每个配置文件只能配置一种语言和一种键盘布局**，不支持额外的语言和布局。
- **没有针对任意 unattend 组件的原始 XML“逃生口”。** 如果你需要某个本工具没有提供字段的组件，
  就只能在事后手工编辑生成的 `autounattend.xml`。
- **不支持图片文件桌面壁纸**，只支持纯色（`personalization.solid_color_wallpaper`）。
- **不支持锁屏图片**：Microsoft 的实现机制仅限于 Enterprise/Education/Pro-SharedPC 版本，对本工具以家庭版/专业版为主的用户来说不可靠，
  所以选择不做，而不是发布一个只能半工作的功能。
