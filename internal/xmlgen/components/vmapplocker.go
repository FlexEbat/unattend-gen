package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 21 (tech.md backlog group C): VM guest tools and AppLocker policy.
// Scripts sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, resource/*.ps1,
// modifier/Optimizations.cs and modifier/AppLocker.cs), not invented from
// memory.

const vboxGuestAdditionsScript = `foreach( $letter in 'DEFGHIJKLMNOPQRSTUVWXYZ'.ToCharArray() ) {
	$exe = "${letter}:\VBoxWindowsAdditions.exe";
	if( Test-Path -LiteralPath $exe ) {
		$certs = "${letter}:\cert";
		Start-Process -FilePath "${certs}\VBoxCertUtil.exe" -ArgumentList "add-trusted-publisher ${certs}\vbox*.cer", "--root ${certs}\vbox*.cer"  -Wait;
		Start-Process -FilePath $exe -ArgumentList '/with_wddm', '/S' -Wait;
		return;
	}
}
'VBoxGuestAdditions.iso is not attached to this VM.';
`

const vmwareToolsScript = `foreach( $letter in 'DEFGHIJKLMNOPQRSTUVWXYZ'.ToCharArray() ) {
	$exe = "${letter}:\setup.exe";
	if( ( Get-Item -LiteralPath $exe -ErrorAction 'SilentlyContinue' | Select-Object -ExpandProperty 'VersionInfo' | Select-Object -ExpandProperty 'ProductName' ) -eq 'VMware Tools' ) {
		Start-Process -FilePath $exe -ArgumentList '/s /v /qn REBOOT=R' -Wait;
		return;
	}
}
'VMware Tools image (windows.iso) is not attached to this VM.';
`

const virtIOGuestToolsScript = `foreach( $letter in 'DEFGHIJKLMNOPQRSTUVWXYZ'.ToCharArray() ) {
	$exe = "${letter}:\virtio-win-guest-tools.exe";
	if( Test-Path -LiteralPath $exe ) {
		Start-Process -FilePath $exe -ArgumentList '/passive', '/norestart' -Wait;
		return;
	}
}
'VirtIO Guest Tools image (virtio-win-*.iso) is not attached to this VM.';
`

const parallelsToolsScript = `foreach( $letter in 'DEFGHIJKLMNOPQRSTUVWXYZ'.ToCharArray() ) {
	$exe = "${letter}:\PTAgent.exe";
	if( Test-Path -LiteralPath $exe ) {
		Start-Process -FilePath $exe -ArgumentList '/install_silent' -Wait;
		return;
	}
}
'Parallels Tools image (prl-tools-win-*.iso) is not attached to this VM.';
`

// vmGuestToolScripts maps a VMGuestTool to its silent-install script. Each
// script scans drive letters D-Z for the tool's installer file (the
// mounted guest-tools ISO) and no-ops with a message if it isn't attached
// — the same detection approach the reference implementation uses,
// because none of these hypervisors expose the drive letter another way
// that's stable across host configurations.
var vmGuestToolScripts = map[profile.VMGuestTool]string{
	profile.VMGuestToolVBoxGuestAdditions: vboxGuestAdditionsScript,
	profile.VMGuestToolVMwareTools:        vmwareToolsScript,
	profile.VMGuestToolVirtIOGuestTools:   virtIOGuestToolsScript,
	profile.VMGuestToolParallelsTools:     parallelsToolsScript,
}

// VMGuestToolsFirstLogonCommands returns one oobeSystem/FirstLogonCommands
// entry per tool in tools (not combined into one command), so a missing
// ISO for one tool doesn't prevent the others from running. The reference
// implementation runs VBox/VMware/Parallels tools at specialize instead
// when Windows Defender is disabled (a silent installer is less likely to
// get blocked before first logon) — this project has no equivalent
// "Defender disabled" signal to key off, so all four always run at first
// logon, the reference implementation's own fallback behavior.
func VMGuestToolsFirstLogonCommands(tools []profile.VMGuestTool, hidePowerShellWindows bool) []string {
	var commands []string
	for _, t := range tools {
		script, ok := vmGuestToolScripts[t]
		if !ok {
			continue
		}
		path := scriptsDir + `\unattend-vm-` + string(t) + `.ps1`
		commands = append(commands, wrapCommand([]string{
			ensureScriptsDirStatement(),
			writeFileStatement(path, []byte(script)),
			invokeCommand(profile.ScriptPs1, path, hidePowerShellWindows),
		}))
	}
	return commands
}

// AppLockerCommand returns one specialize-pass command that embeds
// policyXML as a file and applies it with Set-AppLockerPolicy, first
// enabling and starting the Application Identity service it depends on
// (AppLocker silently no-ops without it). Returns "" if policyXML is nil.
// Unlike the reference implementation, the policy XML is not validated
// against AppLocker's XSD schema here (profile.ValidateProfile only checks
// it's well-formed XML) — Go's standard library has no XSD validator, and
// this project doesn't ship a custom one; a malformed-but-well-formed
// policy will only surface as a Set-AppLockerPolicy error on the target
// machine.
func AppLockerCommand(policyXML *string) string {
	if policyXML == nil {
		return ""
	}
	path := scriptsDir + `\unattend-applocker-policy.xml`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(path, []byte(*policyXML)),
		`Get-Service -Name 'AppIDSvc' | Set-Service -StartupType 'Automatic'`,
		`Get-Service -Name 'AppIDSvc' | Start-Service`,
		fmt.Sprintf(`Set-AppLockerPolicy -XmlPolicy "%s"`, path),
	})
}
