# unattend-gen — Guía de uso

[English](USAGE.md) | [Русский](USAGE.ru.md) | [简体中文](USAGE.zh-CN.md) | Español | [हिन्दी](USAGE.hi.md)

Esta es la referencia de uso completa de `unattend-gen`: una herramienta CLI/TUI en Go que genera
archivos de respuestas `autounattend.xml` para Windows 10/11. Usted crea una sola vez un **perfil**
(un archivo JSON que describe cómo quiere instalar y configurar Windows) y después genera el archivo
de respuestas a partir de él cada vez que reinstale. Sin formulario web, sin acceso a la red, y se
distribuye como un único binario estático.

Para un resumen breve de las funciones consulte [README.es.md](../README.es.md). Esta guía cubre cada
comando, cada pantalla de la TUI y cada campo que puede contener un perfil.

Si algo de aquí y el XML generado discrepan, el código es la referencia. Esta guía está escrita a mano
a partir de la misma fuente de verdad que usa la propia herramienta (`internal/profile/schema.go`),
pero puede contener errores. Abra un issue si encuentra alguno.

## Contenido

- [Qué hace esta herramienta y qué no hace a propósito](#what-it-does)
- [Instalación y compilación](#installing)
- [Inicio rápido](#quick-start)
- [Referencia de la CLI](#cli-reference)
- [La TUI, pantalla por pantalla](#tui-screens)
- [Referencia del JSON del perfil](#profile-json-reference)
- [Ajustes preestablecidos incluidos](#built-in-presets)
- [Perfiles de ejemplo](#example-profiles)
- [Verificar un archivo de respuestas generado](#verifying)
- [Solución de problemas y limitaciones conocidas](#limitations)

<a id="what-it-does"></a>
## Qué hace esta herramienta y qué no hace a propósito

`unattend-gen` genera un único archivo `autounattend.xml`. Lo coloca en una memoria USB (o en la raíz
del medio de instalación) junto al instalador de Windows, y el programa de instalación de Windows lo
ejecuta automáticamente, omitiendo todas las preguntas que usted haya configurado de antemano.

**No particiona ni formatea el disco.** El programa de instalación de Windows siempre pregunta de forma
interactiva dónde instalar Windows; es intencionado, no una función que falte. La herramienta cubre todo lo
que ocurre *después* de indicar a la instalación qué disco o partición usar: idioma, cuentas, ajustes,
eliminación de aplicaciones, personalización, etcétera.

**No necesita acceso a la red para funcionar.** Sin formulario web ni servidor: es un binario local que
lee un perfil JSON y escribe un archivo XML.

<a id="installing"></a>
## Instalación y compilación

Necesita Go 1.23+ para compilar desde el código fuente (esta guía no publica binarios precompilados;
compruebe la página de Releases del repositorio por si existe uno para su plataforma).

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

Esto genera un único binario estático, `unattend-gen` (o `unattend-gen.exe` en Windows), sin dependencias
en tiempo de ejecución.

<a id="quick-start"></a>
## Inicio rápido

El camino más rápido es la TUI interactiva:

```sh
./unattend-gen tui
```

Recorra las pantallas (avance y retroceda con **Ctrl+N**/**Esc**, salte directamente al resumen con
**Ctrl+R**) y después guarde. Esto escribe un perfil JSON desde el que puede regenerar más tarde, o que
puede editar a mano.

Si prefiere trabajar directamente con JSON:

```sh
# Crear un perfil nuevo con valores predeterminados (en su mayoría interactivos)
./unattend-gen profile init my-pc

# ...edite my-pc.json a mano, o continúe en la TUI:
./unattend-gen tui my-pc.json

# Comprobar que es válido
./unattend-gen validate my-pc.json

# Generar el archivo de respuestas (escribe autounattend.xml junto al perfil
# de forma predeterminada)
./unattend-gen generate my-pc.json
```

Copie el `autounattend.xml` resultante a la raíz de su USB de instalación de Windows (junto a
`setup.exe`, o la raíz de la ISO si está creando un medio personalizado) y arranque desde él: el programa
de instalación de Windows lo encuentra y lo aplica automáticamente.

<a id="cli-reference"></a>
## Referencia de la CLI

Todos los comandos validan su entrada de antemano y fallan de forma explícita (código de salida distinto
de cero, texto del error en stderr) en lugar de generar un archivo parcial o adivinado.

### `profile init <name> [--preset minimal|single-user]`

Crea `<name>.json` en el directorio actual. Sin `--preset`, el perfil parte de los valores
predeterminados integrados (idioma/configuración regional/teclado `en-US`, todo lo demás cercano a
«preguntarme durante la instalación»; vea los [valores predeterminados](#defaults) más abajo). Con
`--preset`, parte de uno de los dos ajustes preestablecidos integrados (vea
[Ajustes preestablecidos incluidos](#built-in-presets)), y `name` siempre toma el valor que usted pasó
en la línea de comandos, sea cual sea el contenido del archivo del preestablecido.

```sh
./unattend-gen profile init laptop --preset single-user
# -> laptop.json
```

### `profile list`

Lista cada archivo `*.json` de `./profiles` (una ruta por línea y nada más; esta salida está pensada para
scripts). El directorio debe existir y contener perfiles; no hay otra configuración sobre dónde viven los
perfiles.

```sh
./unattend-gen profile list
```

### `validate <profile.json>`

Ejecuta la misma validación que `generate`, sin construir el archivo de respuestas. Escribe un error por
línea en stderr y termina con un código distinto de cero si algo está mal; si el perfil es válido escribe
`профиль корректен` (los mensajes de validación y esta línea de éxito están en ruso; vea
[Convención de idioma](#text-conventions) más abajo) y termina con 0.

```sh
./unattend-gen validate laptop.json
```

### `generate <profile.json> [-o path]`

Valida el perfil y después construye y escribe el archivo de respuestas. Sin `-o`/`--output`, el archivo
se escribe como `autounattend.xml` en el mismo directorio que el perfil. Si todo va bien, imprime la ruta
escrita.

```sh
./unattend-gen generate laptop.json
./unattend-gen generate laptop.json -o /media/usb/autounattend.xml
```

### `tui [profile.json]`

Abre la interfaz de terminal interactiva. Sin argumentos, parte de los valores predeterminados integrados.
Con la ruta de un perfil, lo carga primero (con el mismo cargador que usan `validate`/`generate`), de modo
que puede ir y venir: editar en la TUI, guardar, retocar el JSON a mano, volver a abrirlo en la TUI, etc.

```sh
./unattend-gen tui
./unattend-gen tui laptop.json
```

<a id="text-conventions"></a>
**Una nota sobre el idioma**: el código, los comentarios, los mensajes de commit y el texto de consola
sencillo (ayuda de los comandos, mensajes de éxito o fallo que ve al redirigir la salida) están en inglés.
El texto que la TUI le muestra mientras rellena un perfil, y todos los mensajes de error de validación,
están en ruso. Es una convención del proyecto, no un error.

<a id="tui-screens"></a>
## La TUI, pantalla por pantalla

Las pantallas aparecen en este orden; **Ctrl+N** pasa a la siguiente, **Esc** vuelve atrás, **Ctrl+R**
salta directamente a Review desde cualquier sitio, **Tab** / **Shift+Tab** mueven entre los campos de una
pantalla y la **barra espaciadora** marca las casillas. Cada pantalla mantiene sincronizado el mismo perfil
en memoria mientras se mueve, así que no se pierde nada al ir y venir.

1. **Welcome** — pantalla de bienvenida, nada que configurar.
2. **Language** — idioma de la interfaz / configuración regional / distribución de teclado (códigos
   BCP-47), modo de edición de Windows (interactivo / clave genérica / clave propia / clave almacenada en
   el firmware BIOS-UEFI), una clave de producto independiente solo para activación y la arquitectura de
   procesador de destino.
3. **Accounts** — nombre del equipo (o déjelo en blanco para que Windows lo genere), zona horaria, hasta 5
   cuentas locales (nombre/nombre visible/contraseña/grupo) en una tabla editable, comportamiento del primer
   inicio de sesión y la casilla, de mejor esfuerzo, «omitir la exigencia de una cuenta de Microsoft».
4. **Tweaks** — configuración rápida (telemetría/diagnósticos) más cada uno de los 33 ajustes del sistema
   (vea la [lista completa](#system-tweaks) más abajo), política de caducidad de contraseñas y de bloqueo de
   cuentas, y ajustes del Explorador de archivos. Todo en una pantalla porque son ajustes del tipo «cambiar
   este valor predeterminado».
5. **Wifi** — configura una red Wi-Fi para conectarse automáticamente en el primer arranque, ya sea rellenando
   SSID/tipo de seguridad/contraseña/oculta, o pegando el XML de un perfil WLAN exportado.
6. **Apps** — tres grupos de casillas: aplicaciones que quitar (paquetes Appx), características de Windows que
   quitar (capacidades DISM) y características opcionales heredadas que quitar (un tercer mecanismo de
   eliminación independiente); vea las [listas completas](#remove-apps) más abajo.
7. **Personalization** — tema claro/oscuro (sistema y aplicaciones por separado), color de énfasis, dónde se
   muestra (Inicio/barra de tareas, barras de título), transparencia y un color sólido de fondo de escritorio.
8. **Accessibility** — Teclas especiales (desactivadas/deshabilitadas/combinación personalizada de opciones)
   y el estado inicial de Bloq Mayús/Bloq Num/Bloq Despl, además de si pulsarlas hace algo.
9. **Desktop** — qué iconos del escritorio se muestran (Este equipo, Papelera de reciclaje, etc.; 13 en total)
   y qué carpetas especiales se anclan en el menú Inicio junto al botón de apagado (Windows 11).
10. **VisualEffects** — «Opciones de rendimiento» de Windows: un preestablecido (mejor apariencia / mejor
    rendimiento) o 17 interruptores individuales de animación/apariencia.
11. **Taskbar** — ajustes del menú Inicio y de la barra de tareas: desactivar widgets, alinear a la izquierda
    la barra de tareas (Windows 11), ocultar el botón Vista de tareas, desactivar resultados de Bing en la
    búsqueda, mostrar siempre todos los iconos de la bandeja, el modo de visualización del cuadro de búsqueda,
    elementos anclados de Inicio (JSON de Windows 11) / mosaicos (XML de Windows 10) e iconos anclados de la
    barra de tareas (vacío o un XML de diseño personalizado).
12. **Advanced** — instalar herramientas de invitado de máquinas virtuales (VirtualBox/VMware/VirtIO/
    Parallels), un XML de directiva de AppLocker en bruto, un script de PowerShell que calcula un nombre de
    equipo dinámico y tres casillas pequeñas: conservar el archivo de respuestas tras la instalación en vez de
    borrarlo, iniciar automáticamente el Narrador y ofuscar las contraseñas de las cuentas en el XML generado.
13. **Scripts** — un script personalizado por categoría (System/DefaultUser/FirstLogon/UserOnce; vea
    [Scripts personalizados](#custom-scripts) más abajo) editable en un área de texto multilínea, más una
    casilla para reiniciar el Explorador después de ejecutar los scripts.
14. **Review** — un resumen actualizado en vivo de todo el perfil y, si valida, la vista previa del XML generado.

No todos los campos de todas las pantallas se corresponden uno a uno con una clave JSON de nivel superior:
varias pantallas editan objetos anidados (por ejemplo `system_tweaks`, `personalization`). La
[Referencia del JSON del perfil](#profile-json-reference) siguiente está organizada por estructura JSON, con
una nota sobre qué pantalla edita cada parte, de modo que puede encontrar un campo desde cualquiera de los
dos lados.

<a id="profile-json-reference"></a>
## Referencia del JSON del perfil

Un perfil es un objeto JSON con `schema_version: 1`. Todos los campos siguientes son opcionales salvo que se
marquen como **obligatorios**; omitir un campo opcional (o ponerlo a `null`, o dejar un booleano en `false`)
significa «dejar tal cual el comportamiento predeterminado de Windows», con una excepción deliberada que se
indica donde aparece.

### Nivel superior

| Campo | Tipo | Notas |
|---|---|---|
| `schema_version` | int | **Obligatorio**, debe ser `1`. |
| `name` | string | **Obligatorio**. Texto libre: no se escribe en el XML, es solo una etiqueta para el propio archivo del perfil. |

### Idioma y edición — *(pantalla Language)*

```json
"language": {
  "ui_language": "en-US",
  "locale": "en-US",
  "keyboard_layout": "en-US"
}
```

Los tres son códigos BCP-47 **obligatorios** (p. ej. `en-US`, `de-DE`, `ru-RU`). `keyboard_layout`
corresponde a la configuración regional de entrada; `locale` establece tanto la configuración regional del
sistema como la del usuario.

```json
"edition": {
  "mode": "interactive",
  "edition": null,
  "product_key": null
}
```

- `mode`, uno de:
  - `"interactive"` — el programa de instalación de Windows le pregunta la edición y la clave durante la
    instalación.
  - `"generic_key"` — requiere que `edition` sea una de `"Home"`, `"Pro"`, `"Education"`, `"Enterprise"`; la
    instalación usa la clave genérica pública de Microsoft (clave de instalación de cliente KMS) para esa
    edición, y más tarde deberá activar Windows usted mismo.
  - `"custom_key"` — requiere `product_key`, una clave real de 25 caracteres con la forma
    `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX`. Esta clave también se reutiliza para la activación posterior salvo que
    establezca `activation_key` (abajo) explícitamente.
  - `"firmware"` — usa la clave de producto que ya está integrada en el firmware BIOS/UEFI del dispositivo
    (típico de las preinstalaciones OEM de Windows); nunca se pide ni se escribe ninguna clave.

```json
"activation_key": null,
"processor_architecture": ""
```

- `activation_key`: una clave independiente usada **solo** para la activación
  (`Microsoft-Windows-Shell-Setup/ProductKey`), con independencia de la clave (si la hay) que `edition` use
  para elegir qué se instala. Déjela en `null` para recurrir a `edition.product_key` (solo cuando
  `edition.mode` es `"custom_key"`) o para activar sin ninguna clave.
- `processor_architecture`: una de `"amd64"` (valor predeterminado si está vacío o ausente), `"x86"`,
  `"arm64"`. Solo se admite una arquitectura por perfil; a diferencia de otros generadores de unattend, esta
  herramienta no construye un único archivo de respuestas que instale en varias arquitecturas.

### Nombre del equipo y zona horaria — *(pantalla Accounts)*

```json
"computer_name": null,
"computer_name_script": null,
"timezone": null
```

- `computer_name`: un nombre de host estático (1–15 caracteres, letras/dígitos/guiones; no puede empezar ni
  terminar en guion ni ser solo dígitos). `null` deja que Windows genere uno aleatorio.
- `computer_name_script`: un script de PowerShell, ejecutado durante la instalación, cuya salida se convierte
  en el nombre del equipo; sirve para generar nombres dinámicamente (p. ej. a partir del número de serie o de
  un esquema de nomenclatura). **Mutuamente excluyente** con `computer_name`: definir ambos es un error de
  validación. El cambio de nombre lo hace un proceso en segundo plano que sigue reaplicando el nombre durante
  una breve ventana tras la instalación, evitando que Windows lo reescriba durante su propia pasada specialize.
- `timezone`: un identificador de zona horaria de Windows, p. ej. `"Russian Standard Time"`,
  `"Pacific Standard Time"`, `"UTC"`. `null` deja que Windows la detecte automáticamente. La herramienta no
  comprueba que la cadena sea una zona horaria real (la lista es grande y depende de la versión), solo que no
  sea una cadena vacía si se establece.

### Cuentas y primer inicio de sesión — *(pantalla Accounts)*

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

- `accounts`: hasta 5 entradas.
  - `name` — **obligatorio**, ≤20 caracteres, sin `" / \ [ ] : ; | = , + * ? < >`.
  - `display_name` — nombre amigable opcional.
  - `password` — `null` significa sin contraseña; una cadena vacía no es válida (use `null`).
  - `group` — **obligatorio**, `"Administrators"` o `"Users"`.
- `first_logon.mode`, uno de:
  - `"none"` — sin inicio de sesión automático; el primer arranque real muestra la pantalla de inicio de
    sesión normal.
  - `"first_created_account"` — inicia sesión automáticamente con la primera entrada de `accounts` (que no
    puede estar vacía).
  - `"builtin_administrator"` — inicia sesión automáticamente con la cuenta Administrador integrada y oculta;
    requiere que se establezca `first_logon.builtin_administrator_password`.

### Configuración rápida y omisiones

```json
"express_settings": { "mode": "interactive" },
"bypass_online_account_requirement": false
```

- `express_settings.mode`: `"interactive"` (la instalación pregunta por la telemetría, etc.), `"all_enabled"`
  o `"all_disabled"`. Cualquier valor distinto de `"interactive"` también oculta automáticamente varias
  pantallas de OOBE (EULA, registro OEM, configuración de red).
- `bypass_online_account_requirement`: intento de mejor esfuerzo para que la instalación termine con una
  cuenta local en lugar de exigir iniciar sesión con una cuenta de Microsoft (escribe el conocido valor de
  registro `BypassNRO`). **No está garantizado**: Microsoft ha cerrado este atajo más de una vez en
  2025–2026; no cuente con él en flotas desatendidas sin probarlo en su compilación actual de Windows. Las
  cuentas locales ya omiten esta pregunta de forma fiable en cuanto hay al menos una entrada en `accounts`,
  con independencia de este indicador.

<a id="system-tweaks"></a>
### Ajustes del sistema — *(pantalla Tweaks)*

Los 33 campos dentro de `"system_tweaks": { ... }` son booleanos simples con valor predeterminado `false`
(sin cambios), **excepto** `keep_sensitive_files`, que se explica más abajo:

| Campo | Qué hace |
|---|---|
| `disable_windows_update` | Pausa/desactiva Windows Update. |
| `disable_uac` | Desactiva los avisos de Control de cuentas de usuario (UAC). |
| `bypass_win11_requirements` | Omite las comprobaciones de hardware de TPM/Arranque seguro/RAM (windowsPE, antes de que la instalación las evalúe). |
| `disable_smart_app_control` | Desactiva Smart App Control. |
| `disable_smart_screen` | Desactiva SmartScreen (sistema y Edge). |
| `disable_fast_startup` | Desactiva el Inicio rápido (arranque híbrido). |
| `disable_system_restore` | Desactiva la Restauración del sistema. |
| `enable_long_paths` | Activa la compatibilidad con rutas largas de NTFS de más de 260 caracteres. |
| `enable_remote_desktop` | Activa Escritorio remoto y la regla del firewall. |
| `allow_powershell_scripts` | Establece la directiva de ejecución de PowerShell en `RemoteSigned`. |
| `disable_last_access_timestamp` | `fsutil behavior set disablelastaccess 1`. |
| `prevent_device_encryption` | Impide el cifrado automático de dispositivo de BitLocker. |
| `disable_auto_sign_on_last_user` | Desactiva el inicio de sesión automático del último usuario interactivo tras un reinicio. |
| `disable_wpbt` | Desactiva la ejecución de Windows Platform Binary Table. |
| `audit_process_creation` | Activa la auditoría de creación de procesos, incluida la línea de comandos. |
| `hide_edge_first_run` | Omite la experiencia de primera ejecución de Edge. |
| `disable_edge_startup_boost` | Desactiva el impulso de inicio y el modo en segundo plano de Edge. |
| `delete_hidden_junctions` | Elimina puntos de unión NTFS heredados (p. ej. `C:\Documents and Settings`); se aplica a la cuenta de instalación y a todas las cuentas futuras. |
| `prevent_automatic_reboot` | Impide que Windows Update reinicie un equipo en uso activo (registra una tarea programada que mantiene las «horas activas» ajustadas a la hora actual). |
| `turn_off_system_sounds` | Establece el esquema de sonido en «Sin sonidos», para la cuenta de instalación y todas las cuentas futuras. |
| `disable_app_suggestions` | Desactiva las aplicaciones sugeridas que Content Delivery Manager instala en silencio. |
| `disable_pointer_precision` | Desactiva «Mejorar la precisión del puntero» (aceleración del ratón). |
| `prevent_device_apps` | Impide que Windows descargue/instale aplicaciones asociadas a dispositivos de hardware concretos. |
| `harden_system_drive_acl` | Quita al grupo «Usuarios autenticados» el acceso de escritura a `C:\`. |
| `make_edge_uninstallable` | Cambia el indicador de directiva interno que permite a Edge mostrar la opción «Desinstalar». |
| `delete_windows_old` | Elimina `C:\Windows.old` (solo relevante en actualizaciones in situ; no hace nada en una instalación limpia). |
| `disable_core_isolation` | Desactiva la Integridad de memoria / seguridad basada en virtualización (útil en algunos invitados de VM o con controladores antiguos). |
| `delete_edge_desktop_icon` | Elimina el acceso directo de Microsoft Edge del escritorio, para la cuenta de instalación y todas las cuentas futuras. |
| `disable_widgets` | Desactiva el panel de Widgets. |
| `left_taskbar` | Alinea a la izquierda la barra de tareas (Windows 11; por defecto está centrada). |
| `hide_task_view_button` | Oculta el botón Vista de tareas de la barra de tareas. |
| `disable_bing_results` | Desactiva los resultados web de Bing en la búsqueda de la barra de tareas. |
| `show_all_tray_icons` | Muestra siempre todos los iconos del área de notificación en lugar de contraer los inactivos (el mecanismo difiere entre Windows 10 y 11; se gestiona automáticamente). |

`keep_sensitive_files` está en el nivel superior del perfil, no dentro de `system_tweaks`; lea la nota
siguiente, porque su comportamiento predeterminado es la única excepción deliberada a «false/ausente significa
sin cambios» en todo este esquema.

```json
"keep_sensitive_files": false
```

Por defecto (`false`, es decir, también **ausente en el JSON**), la herramienta borra
`C:\Windows\Panther\unattend.xml` / `unattend-original.xml` (las copias que el programa de instalación de
Windows conserva de su archivo de respuestas, incluidas las contraseñas en texto plano si no establece también
`obscure_passwords`) y su propio archivo temporal de perfil Wi-Fi cuando termina la instalación. Póngalo en
`true` para dejar esos archivos en su sitio. Es el único campo de todo el esquema donde el valor predeterminado
*hace algo* en lugar de *no cambiar nada*: tener las contraseñas en texto plano en el disco tras la instalación
es peor que romper la convención habitual.

### Caducidad de contraseñas y bloqueo de cuentas — *(pantalla Tweaks)*

```json
"password_expiration": { "mode": "default", "days": null },
"account_lockout": {
  "mode": "default",
  "threshold": null,
  "window_minutes": null,
  "duration_minutes": null
}
```

- `password_expiration.mode`: `"default"` (el valor predeterminado de Windows, 42 días; no se emite ningún
  comando), `"never"` (las contraseñas nunca caducan) o `"custom"` (requiere `days` ≥ 1).
- `account_lockout.mode`: `"default"` (el valor predeterminado de Windows: 10 intentos fallidos / ventana de
  10 min / bloqueo de 10 min), `"disabled"` (bloqueo desactivado) o `"custom"` (requiere los tres campos
  numéricos, cada uno ≥ 1).

### Ajustes del Explorador de archivos — *(pantalla Tweaks)*

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

- `hidden_files`: `"default"`, `"show_hidden"` (mostrar archivos ocultos) o `"show_all"` (mostrar archivos
  ocultos y archivos protegidos del sistema operativo).
- Los demás booleanos: mostrar extensiones de archivo; restaurar el menú contextual clásico (al estilo
  Windows 10) en Windows 11; ocultar las descripciones emergentes de carpetas/iconos del escritorio; abrir el
  Explorador de archivos en «Este equipo» en lugar de «Acceso rápido»/«Inicio»; mostrar «Finalizar tarea»
  directamente en el menú contextual de la barra de tareas.

Todos estos ajustes se aplican tanto a la cuenta creada durante la instalación como a todas las cuentas
futuras del equipo.

### Wi-Fi — *(pantalla Wifi)*

```json
"wifi": {
  "ssid": "MyNetwork",
  "authentication": "WPA2Personal",
  "password": "hunter2000",
  "connect_hidden": false,
  "raw_profile_xml": null
}
```

`wifi` en conjunto es opcional: omítalo (o déjelo en `null`) para no configurar el Wi-Fi.

- `authentication`: `"Open"`, `"WPA2Personal"` o `"WPA3Personal"`. `password` es obligatorio (≥8
  caracteres) salvo que `authentication` sea `"Open"`.
- `raw_profile_xml`: si se establece, este XML de perfil WLAN en bruto (exportado con
  `netsh wlan export profile key=clear`) se usa tal cual en lugar de construir uno a partir de
  `ssid`/`authentication`/`password`/`connect_hidden`; en ese caso esos cuatro pasan a ser opcionales.

<a id="remove-apps"></a>
### Quitar aplicaciones, características y características opcionales — *(pantalla Apps)*

Tres listas separadas, con tres mecanismos de eliminación subyacentes distintos; se mantienen separadas
porque un nombre de una lista no es válido en otra.

```json
"remove_apps": ["OneDrive", "Terminal", "Store"],
"remove_features": ["InternetExplorer"],
"remove_optional_features": ["Recall"]
```

**`remove_apps`** (paquetes Appx, eliminados con `Remove-AppxProvisionedPackage`), cualquiera de:

`3DViewer`, `BingSearch`, `Calculator`, `Camera`, `Clipchamp`, `Clock`,
`Copilot`, `Cortana`, `DevHome`, `Family`, `FeedbackHub`, `GameAssist`,
`GetHelp`, `MailAndCalendar`, `Maps`, `MediaPlayerModern`, `MixedReality`,
`MoviesAndTV`, `News`, `Notepad`, `Office`, `OneDrive`, `OneNote`,
`Outlook`, `Paint`, `Paint3D`, `People`, `PhoneLink`, `PowerAutomate`,
`QuickAssist`, `Skype`, `SnippingTool`, `SolitaireCollection`,
`StickyNotes`, `Store`, `Teams`, `Terminal`, `Tips`, `ToDo`,
`VoiceRecorder`, `Wallet`, `Weather`, `XboxApps`.

`OneDrive` se gestiona de forma distinta internamente (no se distribuye como paquete Appx: la herramienta
elimina su acceso directo sobrante y los ejecutables de instalación, y quita su entrada de inicio
automático), pero se usa igual: simplemente por su nombre en esta misma lista.

**`remove_features`** (*capacidades* opcionales de Windows, eliminadas con `Remove-WindowsCapability`),
cualquiera de: `InternetExplorer`, `WordPad`, `PowerShellISE`, `OpenSSHClient`, `MediaPlayer`, `Speech`,
`Handwriting`, `WindowsHello`, `MathInputPanel`, `OneSync`, `StepsRecorder`.

**`remove_optional_features`** (características opcionales heredadas de Windows, eliminadas con
`Disable-WindowsOptionalFeature`, un tercer mecanismo distinto), cualquiera de: `Recall`, `MediaFeatures`,
`RemoteDesktopClient`.

### Personalización — *(pantalla Personalization)*

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

- `system_theme` / `apps_theme`: `"light"` o `"dark"`; vacío/ausente deja el valor predeterminado de Windows.
- `accent_color` / `solid_color_wallpaper`: colores hexadecimales de 6 dígitos (p. ej. `"FF8800"`), sin
  prefijo `#`. `solid_color_wallpaper` sustituye el fondo del escritorio por un color liso; no se admite un
  fondo de pantalla a partir de un archivo de imagen.
- Se aplican a todas las cuentas futuras, no solo a la creada durante la instalación.

### Accesibilidad — *(pantalla Accessibility)*

```json
"sticky_keys": { "mode": "default", "flags": [] },
"lock_keys": null
```

- `sticky_keys.mode`: `"default"` (el valor predeterminado de Windows: activado, 5×Shift lo activa),
  `"disabled"` (desactiva el propio atajo 5×Shift, no solo los extras visuales) o `"custom"` (use `flags`,
  cualquier subconjunto de `HotKeyActive`, `Indicator`, `TriState`, `TwoKeysOff`, `AudibleFeedback`,
  `HotKeySound`).
- `lock_keys`: `null` deja intacto el comportamiento de Bloq Mayús/Bloq Num/Bloq Despl. Defina un objeto para
  configurar las tres:

  ```json
  "lock_keys": {
    "caps_lock":   { "initial": "off", "behavior": "toggle" },
    "num_lock":    { "initial": "on",  "behavior": "toggle" },
    "scroll_lock": { "initial": "off", "behavior": "ignore" }
  }
  ```

  `initial` es `"off"` u `"on"` (estado justo después de iniciar Windows); `behavior` es `"toggle"` (normal)
  o `"ignore"` (pulsar la tecla no hace nada; surte efecto tras reiniciar).

### Iconos del escritorio y carpetas de Inicio — *(pantalla Desktop)*

```json
"desktop_icons": { "ThisPC": true, "RecycleBin": false },
"start_folders": ["Settings", "Documents", "Downloads"]
```

- `desktop_icons`: un mapa; las claves que no menciona se dejan en el valor predeterminado de Windows, las que
  menciona se establecen explícitamente en `true` (visible) o `false` (oculto). Claves válidas: `ThisPC`,
  `UserFiles`, `Network`, `RecycleBin`, `ControlPanel`, `Desktop`, `Documents`, `Downloads`, `Music`,
  `Pictures`, `Videos`, `Gallery`, `Home`.
- `start_folders`: una lista ordenada de carpetas especiales ancladas junto al botón de apagado del menú Inicio
  de Windows 11; el orden de la lista es el orden en que se anclan. Valores válidos: `Settings`,
  `FileExplorer`, `Documents`, `Downloads`, `Music`, `Pictures`, `Videos`, `Network`, `PersonalFolder`. Una
  lista vacía o ausente significa «dejar tal cual el conjunto predeterminado de Windows»: con este campo no hay
  forma de expresar «anclar cero carpetas».

Ambos se aplican a todas las cuentas futuras, no solo a la creada durante la instalación.

### Efectos visuales — *(pantalla VisualEffects)*

```json
"visual_effects": { "mode": "default", "custom": {} }
```

- `mode`: `"default"` (sin cambios), `"best_appearance"` (todos los efectos activados), `"best_performance"`
  (todos los efectos desactivados) o `"custom"` (use `custom`).
- `custom`: un mapa de nombre de efecto → `true`/`false`; los efectos que no menciona conservan el valor
  predeterminado de Windows. Claves válidas: `ControlAnimations`, `AnimateMinMax`, `TaskbarAnimations`,
  `DWMAeroPeekEnabled`, `MenuAnimation`, `TooltipAnimation`, `SelectionFade`, `DWMSaveThumbnailEnabled`,
  `CursorShadow`, `ListviewShadow`, `ThumbnailsOrIcon`, `ListviewAlphaSelect`, `DragFullWindows`,
  `ComboBoxAnimation`, `FontSmoothing`, `ListBoxSmoothScrolling`, `DropShadow`.

Se aplica a todas las cuentas futuras, no solo a la creada durante la instalación.

### Menú Inicio y barra de tareas — *(pantalla Taskbar)*

```json
"taskbar_search": "",
"start_pins": { "mode": "default", "json": null },
"start_tiles": { "mode": "default", "xml": null },
"taskbar_icons": { "mode": "default", "xml": null }
```

- `taskbar_search`: `""`/ausente (valor predeterminado de Windows, cuadro de búsqueda visible), `"hide"`,
  `"icon"` (solo icono, sin cuadro), `"box"` (explícito, igual que el predeterminado) o `"label"` (icono con
  etiqueta de texto).
- `start_pins` (solo Windows 11, sin efecto en Windows 10): `mode` es `"default"`, `"empty"` (ninguna
  aplicación anclada) o `"custom"` (requiere `json`: una carga `{"pinnedList": [...]}` en bruto con el formato
  que espera la directiva `ConfigureStartPins` de Windows).
- `start_tiles` (solo Windows 10, sin efecto en Windows 11): `mode` es `"default"`, `"empty"` (sin grupos de
  mosaicos) o `"custom"` (requiere `xml`: un documento `LayoutModification.xml` en bruto).
- `taskbar_icons`: iconos anclados a la barra de tareas para todas las cuentas futuras. `mode` es `"default"`,
  `"empty"` (sin iconos anclados) o `"custom"` (requiere `xml`: un `LayoutModification.xml` en bruto que
  contenga un `CustomTaskbarLayoutCollection`). Funciona mediante un diseño de Inicio bloqueado que se
  desbloquea de nuevo en el primer inicio de sesión de cada cuenta, de modo que los usuarios pueden reorganizar
  después la barra de tareas; por eso el archivo de respuestas generado también registra una pequeña tarea
  programada `UnlockStartLayout`. Solo se comprueba que el XML esté bien formado, no el esquema de diseño de
  la barra de tareas de Windows.

### Avanzado — *(pantalla Advanced)*

```json
"install_vm_guest_tools": [],
"applocker_policy_xml": null,
"use_narrator": false,
"obscure_passwords": false
```

- `install_vm_guest_tools`: cualquiera de `VBoxGuestAdditions`, `VMwareTools`, `VirtIoGuestTools`,
  `ParallelsTools`. Cada uno ejecuta un instalador silencioso que busca su ISO de herramientas de invitado en
  las letras de unidad D–Z y no hace nada (con un mensaje en el registro) si no está conectada; es seguro
  listar las cuatro «por si acaso» si no sabe de antemano en qué hipervisor estará.
- `applocker_policy_xml`: un XML de directiva de AppLocker en bruto. Solo se comprueba que esté bien formado,
  no se valida contra el esquema completo de AppLocker: una directiva bien formada pero inválida solo se
  manifestará como un error en el equipo de destino.
- `use_narrator`: inicia automáticamente el Narrador durante la propia instalación y en el primer inicio de
  sesión de cada cuenta futura.
- `obscure_passwords`: ofusca con Base64 las contraseñas de las cuentas en el XML generado en lugar de
  escribirlas en texto plano. Es **ofuscación, no cifrado**: la sal es un valor fijo y documentado
  públicamente (la propia convención de unattend de Microsoft), así que solo evita que las contraseñas se
  encuentren trivialmente con grep en el archivo en bruto; nada más.

### Interruptor global de instalación

```json
"hide_powershell_windows": false
```

Con `true`, cada ventana de PowerShell que esta herramienta lanza durante la instalación (tanto los ajustes
integrados como sus propios scripts personalizados) se ejecuta oculta en lugar de visible. No afecta a los
pasos basados en `cmd`/`reg`/`vbs`, que no muestran ventana en ningún caso.

<a id="custom-scripts"></a>
### Scripts personalizados — *(pantalla Scripts)*

```json
"system_scripts": [],
"default_user_scripts": [],
"first_logon_scripts": [],
"user_once_scripts": [],
"restart_explorer_after_scripts": false
```

Cuatro categorías, cada una una lista de objetos `{ "format": "...", "content": "..." }`. `format` es uno de
`cmd`, `ps1`, `reg`, `vbs` (`default_user_scripts` no admite `vbs`). Límites: `system_scripts` ≤4,
`default_user_scripts` ≤3, `first_logon_scripts` ≤4, `user_once_scripts` ≤4.

- **`system_scripts`** — se ejecutan una vez, en el contexto del sistema, *antes* de que exista cualquier
  cuenta de usuario.
- **`default_user_scripts`** — se aplican a la plantilla de registro del usuario predeterminado, de modo que
  afectan a todas las cuentas creadas después, incluidas las que aún no existen; no a la cuenta creada durante
  la propia instalación (salvo que también se cree a partir de esta plantilla, lo cual es lo habitual).
- **`first_logon_scripts`** — se ejecutan una vez, cuando inicia sesión la primera cuenta.
- **`user_once_scripts`** — se ejecutan una vez *por cuenta*, incluidas las futuras, la primera vez que inicia
  sesión cada una.

`restart_explorer_after_scripts`: si se ejecutó algún `first_logon_scripts`, reinicia el Explorador después
(útil si un script cambió algo que el Explorador guarda en caché).

<a id="defaults"></a>
### Qué le da `profile init` (sin preestablecido)

`schema_version: 1`, idioma/configuración regional/teclado todos `"en-US"`, `edition.mode: "interactive"`,
sin cuentas, `first_logon.mode: "none"`, `express_settings.mode: "interactive"`; todo lo demás es el valor
cero de JSON (cadena vacía/false/null/lista vacía), es decir, «preguntarme durante la instalación» o «dejar
el valor predeterminado de Windows» para cada ajuste que describe esta guía.

<a id="built-in-presets"></a>
## Ajustes preestablecidos incluidos

Dos preestablecidos vienen integrados en el binario; úselos con
`profile init <name> --preset <preset-name>`.

- **`minimal`** — básicamente lo mismo que no usar preestablecido: edición interactiva, sin cuentas,
  configuración rápida interactiva, `false` explícito para los tres ajustes originales del sistema. Un punto
  de partida seguro que no hace nada extra, sobre el que ir construyendo.
- **`single-user`** — una cuenta local de Administrador llamada `admin` (sin contraseña; añada una antes de
  usarlo en serio), con inicio de sesión automático en esa cuenta (`first_logon.mode: "first_created_account"`)
  y `express_settings.mode: "all_disabled"` (se omiten por completo las preguntas de telemetría/diagnósticos).
  Un punto de partida razonable para un equipo personal que solo usted utiliza.

Ambos preestablecidos son anteriores a varios de los campos documentados arriba (se escribieron para un
esquema anterior y más pequeño): lo que no mencionan toma el valor cero predeterminado del esquema, igual que
un campo sin definir en cualquier otro sitio.

<a id="example-profiles"></a>
## Perfiles de ejemplo

Un perfil más completo, que combina varias de las secciones anteriores: un portátil de un solo usuario con
algunos ajustes de privacidad/rendimiento y un par de aplicaciones eliminadas:

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

Una instalación mínima desatendida sin florituras (sin cuentas propias, solo omitir las comprobaciones de
hardware y pausar las actualizaciones) con una clave genérica:

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
## Verificar un archivo de respuestas generado

`validate`/`generate` solo comprueban el perfil JSON contra el esquema: no pueden confirmar que el programa de
instalación de Windows aceptará de principio a fin el XML resultante, porque eso requiere una ejecución real de
la instalación de Windows. No hay sustituto para grabar el resultado en una memoria USB (o montarlo en una VM)
y ver cómo se instala. Si algo del XML está mal, el programa de instalación de Windows normalmente le indicará
qué elemento mediante su propio cuadro de diálogo de error o en `setupact.log`/`setuperr.log` dentro de
`C:\Windows\Panther` (o `X:\Windows\Panther` mientras sigue en el entorno PE).

<a id="limitations"></a>
## Solución de problemas y limitaciones conocidas

- **El particionado de discos queda fuera de alcance, a propósito.** El programa de instalación de Windows
  siempre se detendrá y le preguntará en qué disco/partición instalar. Esta herramienta configura todo lo demás
  alrededor de ese paso interactivo.
- **`bypass_online_account_requirement` es de mejor esfuerzo.** Microsoft ha cambiado su funcionamiento más de
  una vez; si deja de funcionar en una compilación futura de Windows, es un cambio por parte de Microsoft, no un
  informe de error contra esta herramienta (aunque un issue al respecto sigue siendo bienvenido).
- **No se admiten archivos de respuestas multiarquitectura.** Elija un `processor_architecture` por perfil.
- **Solo se puede configurar un idioma y una distribución de teclado** por perfil; no se admiten idiomas ni
  distribuciones adicionales.
- **No hay una «válvula de escape» de XML en bruto para componentes unattend arbitrarios.** Si necesita un
  componente para el que esta herramienta no expone un campo, tendrá que editar a mano el `autounattend.xml`
  generado después.
- **No se admite el fondo de escritorio a partir de un archivo de imagen**: solo un color liso
  (`personalization.solid_color_wallpaper`).
- **No se admite la imagen de la pantalla de bloqueo**: el mecanismo de Microsoft para ello está limitado a las
  ediciones Enterprise/Education/Pro-SharedPC, lo que lo hace poco fiable para el público de esta herramienta
  (sobre todo Home/Pro), así que se dejó fuera en lugar de publicarlo a medias.
