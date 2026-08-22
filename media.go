// Package ebitter exposes the embedded game assets used by the demo build.
package ebitter

import "embed"

//go:embed media
var Media embed.FS

//go:generate go run github.com/mcbalaam/ebitter/cmd/fontgen
