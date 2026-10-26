package components

import (
	"fmt"
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Windows Setup documents a maximum length of 259 characters for
// RunSynchronousCommand/Path (and FirstLogonCommands/CommandLine). Most of
// the commands this project builds embed whole scripts as Base64, so they are
// far longer. OffloadLongCommands brings every command under the limit:
//
//   - a "cmd.exe /c a && b" chain is split into one command per part;
//   - a PowerShell command is moved into a file that is embedded in the
//     answer file's <Extensions> element and extracted by one short command at
//     the very start of the specialize pass. This is the same mechanism the
//     reference implementation uses (Main.cs EmbedFile, resource/
//     ExtractScripts.ps1, github.com/cschneegans/unattend-generator).
const (
	// MaxCommandLength is the documented maximum length of a command line.
	MaxCommandLength = 259

	// ExtensionsNamespace is the XML namespace of the <Extensions> element.
	// Windows Setup ignores foreign-namespace elements; our own extraction
	// script reads them back from the copy of the answer file that Setup
	// keeps in C:\Windows\Panther.
	ExtensionsNamespace = "https://schneegans.de/windows/unattend-generator/"

	psCommandPrefix  = `powershell -NoProfile -Command "`
	cmdCommandPrefix = `cmd.exe /c `
	cmdChainSep      = ` && `
)

// ExtractScript is copied verbatim from resource/ExtractScripts.ps1 of the
// reference implementation (see THIRD_PARTY_NOTICES.md). It writes every
// <File> of the <Extensions> element to its path.
const ExtractScript = `param(
    [xml] $Document
);

foreach( $file in $Document.unattend.Extensions.File ) {
    $path = [System.Environment]::ExpandEnvironmentVariables( $file.GetAttribute( 'path' ) );
    mkdir -Path( $path | Split-Path -Parent ) -ErrorAction 'SilentlyContinue';
    $encoding = switch( [System.IO.Path]::GetExtension( $path ) ) {
        { $_ -in '.ps1', '.xml' } { [System.Text.Encoding]::UTF8; }
        { $_ -in '.reg', '.vbs', '.js' } { [System.Text.UnicodeEncoding]::new( $false, $true ); }
        default { [System.Text.Encoding]::Default; }
    };
    $bytes = $encoding.GetPreamble() + $encoding.GetBytes( $file.InnerText.Trim() );
    [System.IO.File]::WriteAllBytes( $path, $bytes );
}`

// extractCommand runs ExtractScript against the answer file copy Setup
// keeps in C:\Windows\Panther. Taken from the reference implementation.
const extractCommand = `powershell.exe -NoProfile -Command "$xml = [xml]::new(); $xml.Load('C:\Windows\Panther\unattend.xml'); $sb = [scriptblock]::Create( $xml.unattend.Extensions.ExtractScript ); Invoke-Command -ScriptBlock $sb -ArgumentList $xml;"`

// ExtensionFile is one file embedded in the <Extensions> element.
type ExtensionFile struct {
	Path    string
	Content string
}

// extensionFilePrefix is the common name prefix of every offloaded command
// file. The sensitive-files cleanup script deletes files with this prefix.
const extensionFilePrefix = "unattend-cmd-"

type offloader struct {
	files []ExtensionFile
}

func (o *offloader) addFile(ext, content string) string {
	path := fmt.Sprintf(`%s\%s%03d.%s`, scriptsDir, extensionFilePrefix, len(o.files)+1, ext)
	o.files = append(o.files, ExtensionFile{Path: path, Content: content})
	return path
}

// splitCommand returns the commands cmd should be replaced with so that none
// of them needs a file: a long "cmd.exe /c a && b" chain becomes one command
// per part. Any other command is returned unchanged.
func splitCommand(cmd string) []string {
	if len(cmd) <= MaxCommandLength || !strings.HasPrefix(cmd, cmdCommandPrefix) {
		return []string{cmd}
	}
	parts := strings.Split(strings.TrimPrefix(cmd, cmdCommandPrefix), cmdChainSep)
	if len(parts) == 1 {
		return []string{cmd}
	}
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = cmdCommandPrefix + p
	}
	return out
}

// toFile moves a long PowerShell command (as built by wrapCommand) into an
// extension file and returns the short command that runs that file. Anything
// else is returned unchanged. The file content is the command's script text
// as PowerShell receives it, i.e. with the \" escapes of the outer
// -Command "..." string undone.
func (o *offloader) toFile(cmd string) string {
	if len(cmd) <= MaxCommandLength || !strings.HasPrefix(cmd, psCommandPrefix) || !strings.HasSuffix(cmd, `"`) {
		return cmd
	}
	body := strings.TrimSuffix(strings.TrimPrefix(cmd, psCommandPrefix), `"`)
	script := strings.ReplaceAll(body, `\"`, `"`)
	path := o.addFile("ps1", script)
	return fmt.Sprintf(`powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%s"`, path)
}

func (o *offloader) rewrite(cmd string, useFiles bool) []string {
	parts := splitCommand(cmd)
	if useFiles {
		for i, p := range parts {
			parts[i] = o.toFile(p)
		}
	}
	return parts
}

func (o *offloader) runList(list []runSynchronousCommand, useFiles bool) []runSynchronousCommand {
	var out []runSynchronousCommand
	for _, c := range list {
		for _, p := range o.rewrite(c.Path, useFiles) {
			out = append(out, newRunSynchronousCommand(len(out)+1, p))
		}
	}
	return out
}

func (o *offloader) firstLogonList(list []synchronousCommand) []synchronousCommand {
	var out []synchronousCommand
	for _, c := range list {
		for _, p := range o.rewrite(c.CommandLine, true) {
			out = append(out, synchronousCommand{Action: wcmActionAdd, Order: len(out) + 1, CommandLine: p})
		}
	}
	return out
}

// OffloadLongCommands rewrites every command of setup (windowsPE),
// deployment (specialize) and oobe (oobeSystem) so that it fits within
// MaxCommandLength, and returns the possibly-new deployment component plus
// the files that must be embedded in the answer file's <Extensions> element.
// Any argument may be nil.
//
// windowsPE commands can only be split, never moved into files: the answer
// file copy in C:\Windows\Panther that the extraction reads does not exist
// yet at that point.
func OffloadLongCommands(setup *Setup, deployment *Deployment, oobe *ShellSetupOOBE, arch profile.ProcessorArchitecture) (*Deployment, []ExtensionFile) {
	o := &offloader{}

	if setup != nil && setup.RunSynchronous != nil {
		setup.RunSynchronous.RunSynchronousCommand = o.runList(setup.RunSynchronous.RunSynchronousCommand, false)
	}
	if deployment != nil && deployment.RunSynchronous != nil {
		deployment.RunSynchronous.RunSynchronousCommand = o.runList(deployment.RunSynchronous.RunSynchronousCommand, true)
	}
	if oobe != nil && oobe.FirstLogonCommands != nil {
		oobe.FirstLogonCommands.SynchronousCommand = o.firstLogonList(oobe.FirstLogonCommands.SynchronousCommand)
	}
	if len(o.files) == 0 {
		return deployment, nil
	}

	extract := newRunSynchronousCommand(1, extractCommand)
	if deployment == nil {
		deployment = &Deployment{
			Name:           "Microsoft-Windows-Deployment",
			standardAttrs:  newStandardAttrs(arch),
			RunSynchronous: &runSynchronous{},
		}
	}
	rest := deployment.RunSynchronous.RunSynchronousCommand
	list := make([]runSynchronousCommand, 0, len(rest)+1)
	list = append(list, extract)
	for _, c := range rest {
		list = append(list, newRunSynchronousCommand(len(list)+1, c.Path))
	}
	deployment.RunSynchronous.RunSynchronousCommand = list
	return deployment, o.files
}
