# unattend-gen

[English](README.md) | [Русский](README.ru.md) | [简体中文](README.zh-CN.md) | Español | [हिन्दी](README.hi.md)

Una herramienta CLI y TUI que genera archivos de respuestas `autounattend.xml` para instalaciones desatendidas
de Windows 10/11. Es el equivalente en terminal de
[schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/): idioma y
edición, nombre del equipo y cuentas locales, telemetría y ajustes del sistema, Wi-Fi. Configure una vez y
reutilice el resultado en cada instalación.

Un perfil es un archivo JSON normal con todos los ajustes. La CLI y la TUI leen y escriben el mismo formato de
perfil y construyen el XML con el mismo código, de modo que un mismo perfil siempre produce el mismo archivo de
respuestas, sin importar cómo se rellenó.

**Documentación completa:** [docs/USAGE.es.md](docs/USAGE.es.md) ([English](docs/USAGE.md)): cada comando de
la CLI, un recorrido pantalla por pantalla de la TUI y una referencia completa de cada campo que puede contener
un perfil.

## Funciones

- Idioma, configuración regional, distribución de teclado, edición de Windows y clave de producto (incluidas las
  claves almacenadas en el firmware BIOS/UEFI y una clave independiente solo para activación), arquitectura de
  procesador (x64/x86/ARM64)
- Nombre del equipo, zona horaria y hasta 5 cuentas locales, con control del inicio de sesión automático
- Configuración rápida (telemetría) y 33 ajustes del sistema (Windows Update, UAC, omitir los requisitos de
  hardware de Windows 11, SmartScreen, inicio rápido, restauración del sistema, rutas largas, Escritorio remoto,
  limpieza de puntos de unión, evitar reinicios tras actualizaciones, refuerzo de ACL y más)
- Las pantallas de OOBE (EULA, registro OEM, configuración de red) se ocultan automáticamente cuando el perfil
  ya aporta lo necesario para instalar sin ellas; un indicador aparte permite omitir la exigencia de una cuenta
  de Microsoft (de mejor esfuerzo: Microsoft ha cerrado este atajo más de una vez, así que no funciona en todas
  las compilaciones de Windows)
- Perfil de Wi-Fi (SSID, WPA2/WPA3/red abierta, redes ocultas)
- Eliminación de aplicaciones preinstaladas (Xbox, Teams, Solitario, OneDrive, Microsoft Store, Terminal de
  Windows y 36 más), de características de Windows (Internet Explorer, WordPad, cliente OpenSSH, Windows Hello y
  otras) y de características opcionales heredadas (Recall, cliente de Escritorio remoto, Media Features)
- Scripts personalizados (.cmd/.ps1/.reg/.vbs) en cuatro puntos: System (antes de crear las cuentas),
  DefaultUser (cada cuenta, incluidas las futuras), FirstLogon (una vez) y UserOnce (una vez por cuenta)
- Directiva de caducidad de contraseñas y de bloqueo de cuentas
- Ajustes del Explorador de archivos (archivos ocultos y de sistema, extensiones de archivo, menú contextual
  clásico, descripciones emergentes, carpeta predeterminada, «Finalizar tarea» en la barra de tareas) aplicados a
  todas las cuentas, incluidas las futuras
- Personalización (tema claro/oscuro, color de énfasis, transparencia, fondo de color liso) con el mismo
  mecanismo para todas las cuentas
- Preestablecidos de efectos visuales (mejor apariencia, mejor rendimiento o 17 interruptores individuales) para
  todas las cuentas futuras
- Menú Inicio y barra de tareas: modo del cuadro de búsqueda, alineación a la izquierda (Win11), ocultar el
  botón Vista de tareas, mostrar siempre todos los iconos de la bandeja, desactivar widgets y resultados de Bing,
  elementos anclados de Inicio (JSON de Win11) y mosaicos (XML de Win10), iconos anclados de la barra de tareas
  (vacío o un XML de diseño personalizado)
- Teclas especiales (predeterminado/desactivadas/personalizado) y el estado inicial y comportamiento de Bloq
  Mayús/Bloq Num/Bloq Despl, para cuentas futuras, la sesión actual y la pantalla de inicio de sesión
- Visibilidad de los iconos del escritorio (Este equipo, Papelera de reciclaje y 11 más) y carpetas ancladas al
  menú Inicio (Win11) para todas las cuentas futuras
- Instalación automática de herramientas de invitado de máquinas virtuales (VirtualBox, VMware, VirtIO,
  Parallels) y XML de directiva de AppLocker en bruto
- Nombre de equipo dinámico mediante un script de PowerShell, ofuscación Base64 de las contraseñas de las cuentas
  en el XML, inicio automático del Narrador, conservación opcional del archivo de respuestas tras la instalación,
  XML de perfil WLAN exportado en bruto y un interruptor global que oculta todas las ventanas de PowerShell
  durante la instalación
- Dos preestablecidos integrados (`minimal`, `single-user`) como punto de partida
- TUI interactiva para ir rellenando el perfil paso a paso con una vista previa del XML en vivo antes de guardar
- Un único binario estático: sin servidor, sin llamadas de red, sin más configuración que el JSON del perfil

El particionado de discos no se admite a propósito: el programa de instalación de Windows siempre pregunta dónde
instalar, igual que en una instalación manual normal.

## Tecnologías

- [Go](https://go.dev/)
- [spf13/cobra](https://github.com/spf13/cobra): comandos de la CLI
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea),
  [bubbles](https://github.com/charmbracelet/bubbles),
  [lipgloss](https://github.com/charmbracelet/lipgloss): TUI
- [go-playground/validator](https://github.com/go-playground/validator): validación del perfil

## Instalación

Requiere Go 1.23+.

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

Compilación cruzada para otro sistema operativo con `GOOS`/`GOARCH`:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o unattend-gen.exe ./cmd/unattend-gen
```

## Uso

Cree un perfil, valídelo y conviértalo en un archivo de respuestas:

```sh
unattend-gen profile init demo                    # demo.json con los valores predeterminados
unattend-gen profile init demo --preset minimal    # o parta de un preestablecido
unattend-gen profile list                          # lista los perfiles de ./profiles
unattend-gen validate demo.json                     # código de salida 0 o 1
unattend-gen generate demo.json                     # escribe autounattend.xml junto al perfil
unattend-gen generate demo.json -o out.xml          # o en la ruta que elija
```

O rellene el perfil de forma interactiva:

```sh
unattend-gen tui               # desde cero
unattend-gen tui demo.json     # desde un perfil existente
```

Coloque el `autounattend.xml` resultante en la raíz de una memoria USB de arranque de Windows (o móntelo como
disquete/CD virtual en una máquina virtual) y el programa de instalación de Windows lo recogerá
automáticamente.

## Estructura del proyecto

```text
cmd/unattend-gen/     punto de entrada
internal/profile/     esquema de Profile, carga/guardado de JSON, validación
internal/xmlgen/      constructor de autounattend.xml y sus componentes
internal/cli/         comandos cobra: profile, validate, generate, tui
internal/tui/         aplicación bubbletea: pantallas y widgets compartidos
presets/              preestablecidos de perfil integrados, incrustados en el binario
```

## Desarrollo

```sh
make gate   # comprueba gofmt, go vet, golangci-lint, go test -race
```

La CI ejecuta la misma comprobación en cada push y además verifica la compilación cruzada para Linux, macOS y
Windows.

## Pruebas

```sh
go test ./... -race
```

## Licencia

[GPL-3.0](LICENSE). Partes de los scripts y listas generados están adaptadas de
[cschneegans/unattend-generator](https://github.com/cschneegans/unattend-generator) (MIT); vea
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
