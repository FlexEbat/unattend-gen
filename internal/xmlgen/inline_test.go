package xmlgen

import "github.com/FlexEbat/unattend-gen/internal/profile"

// buildInline builds the answer file without moving long commands into
// <Extensions>. Most tests inspect what one feature's command does, which is
// only readable while the command is still a single string in the document;
// offload_test.go checks that offloading those commands loses nothing.
func buildInline(p *profile.Profile) (string, error) {
	return buildAnswerFile(p, false)
}
