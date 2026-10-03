# unattend-gen

[English](README.md) | [Русский](README.ru.md) | 简体中文 | [Español](README.es.md) | [हिन्दी](README.hi.md)

一个 CLI 与 TUI 工具，用于生成 Windows 10/11 无人值守安装所需的 `autounattend.xml` 应答文件。
它是 [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/)
的终端版本：语言与版本、计算机名与本地账户、遥测与系统调整、Wi-Fi。配置一次，
之后每次安装都可以复用结果。

配置文件（profile）就是一个包含所有设置的普通 JSON 文件。CLI 和 TUI 读写同一种配置文件格式，
并用同一份代码构建 XML，所以无论配置文件是怎么填写的，相同的配置文件总会得到相同的应答文件。

**完整文档：** [docs/USAGE.zh-CN.md](docs/USAGE.zh-CN.md)（[English](docs/USAGE.md)）——
包含每条 CLI 命令、逐屏的 TUI 介绍，以及配置文件中每个字段的完整参考。

## 功能

- 语言、区域、键盘布局、Windows 版本与产品密钥（包括存储在 BIOS/UEFI 固件中的密钥，
  以及单独的仅用于激活的密钥）、处理器架构（x64/x86/ARM64）
- 计算机名、时区和最多 5 个本地账户，可控制自动登录
- 快速设置（遥测）和 33 项系统调整（Windows 更新、UAC、绕过 Windows 11 硬件要求、SmartScreen、
  快速启动、系统还原、长路径、远程桌面、清理 junction 点、防止更新后自动重启、加固 ACL 等）
- 当配置文件已提供足够设置、无需这些界面即可完成安装时，会自动隐藏 OOBE 界面（EULA、OEM 注册、网络设置）；
  另有一个标志用于跳过要求 Microsoft 账户（尽力而为：Microsoft 不止一次堵上了这个绕过方式，
  因此并非在每个 Windows 版本上都有效）
- Wi-Fi 配置（SSID、WPA2/WPA3/开放网络、隐藏网络）
- 移除预装应用（Xbox、Teams、纸牌、OneDrive、Microsoft Store、Windows 终端等另外 36 个）、
  Windows 功能（Internet Explorer、WordPad、OpenSSH 客户端、Windows Hello 等）和旧版可选功能
  （Recall、远程桌面客户端、Media Features）
- 在四个时机运行的自定义脚本（.cmd/.ps1/.reg/.vbs）：System（创建账户之前）、DefaultUser（每个账户，包括未来账户）、
  FirstLogon（仅一次）和 UserOnce（每个账户一次）
- 密码过期与账户锁定策略
- 文件资源管理器调整（隐藏和系统文件、文件扩展名、经典右键菜单、提示、默认文件夹、任务栏“结束任务”），
  对所有账户（包括未来账户）生效
- 个性化（浅色/深色主题、强调色、透明度、纯色壁纸），同样适用于所有账户
- 视觉效果预设（最佳外观、最佳性能，或 17 个单独开关），适用于所有未来账户
- 开始菜单与任务栏：搜索框模式、左对齐（Win11）、隐藏任务视图按钮、始终显示所有托盘图标、
  禁用小组件和 Bing 结果、开始菜单固定项（Win11 JSON）和磁贴（Win10 XML）、固定的任务栏图标（空白或自定义布局 XML）
- 粘滞键（默认/禁用/自定义）以及 Caps/Num/Scroll Lock 的初始状态与行为，适用于未来账户、当前会话和登录界面
- 桌面图标可见性（此电脑、回收站等 13 项）和固定在开始菜单的文件夹（Win11），适用于所有未来账户
- 自动安装虚拟机来宾工具（VirtualBox、VMware、VirtIO、Parallels）和原始 AppLocker 策略 XML
- 通过 PowerShell 脚本动态生成计算机名、在 XML 中对账户密码做 Base64 混淆、自动启动讲述人、
  可选在安装后保留应答文件、原始导出 WLAN 配置文件 XML，以及一个在安装期间隐藏所有 PowerShell 窗口的全局开关
- 两个内置预设（`minimal`、`single-user`）作为起点
- 交互式 TUI，逐步填写配置文件，保存前可实时预览 XML
- 单个静态二进制文件：无服务器、无网络调用、除配置文件 JSON 外无需任何配置

磁盘分区刻意不在支持范围内：与普通手动安装一样，Windows 安装程序始终会询问安装到哪里。

## 技术栈

- [Go](https://go.dev/)
- [spf13/cobra](https://github.com/spf13/cobra) — CLI 命令
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)、
  [bubbles](https://github.com/charmbracelet/bubbles)、
  [lipgloss](https://github.com/charmbracelet/lipgloss) — TUI
- [go-playground/validator](https://github.com/go-playground/validator) — 配置文件校验

## 安装

需要 Go 1.23+。

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

通过 `GOOS`/`GOARCH` 交叉编译到其他操作系统：

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o unattend-gen.exe ./cmd/unattend-gen
```

## 使用

创建配置文件，校验它，并把它变成应答文件：

```sh
unattend-gen profile init demo                    # 使用默认值创建 demo.json
unattend-gen profile init demo --preset minimal    # 或者从预设开始
unattend-gen profile list                          # 列出 ./profiles 中的配置文件
unattend-gen validate demo.json                     # 退出码 0 或 1
unattend-gen generate demo.json                     # 在配置文件旁写出 autounattend.xml
unattend-gen generate demo.json -o out.xml          # 或写到指定路径
```

或者以交互方式填写配置文件：

```sh
unattend-gen tui               # 从头开始
unattend-gen tui demo.json     # 从已有的配置文件开始
```

把得到的 `autounattend.xml` 放到 Windows 启动 U 盘的根目录（或在虚拟机中挂载为虚拟软盘/光盘），
Windows 安装程序会自动识别并使用它。

## 项目结构

```text
cmd/unattend-gen/     入口点
internal/profile/     Profile 模式、JSON 加载/保存、校验
internal/xmlgen/      autounattend.xml 构建器及其组件
internal/cli/         cobra 命令：profile、validate、generate、tui
internal/tui/         bubbletea 应用：各界面与共享控件
presets/              内置配置文件预设，嵌入二进制文件
```

## 开发

```sh
make gate   # 检查 gofmt、go vet、golangci-lint、go test -race
```

CI 会在每次 push 时运行同样的检查，并验证到 Linux、macOS 和 Windows 的交叉编译。

## 测试

```sh
go test ./... -race
```

## 许可证

[GPL-3.0](LICENSE)。生成的部分脚本和列表改编自
[cschneegans/unattend-generator](https://github.com/cschneegans/unattend-generator)（MIT）；
参见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。
