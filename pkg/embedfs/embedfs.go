// Package embedfs provides a cross-platform filesystem for embedded assets.
// Desktop: reads from the real filesystem via os.DirFS.
// WASM: reads from embedded data set by the main package.
package embedfs

import "io/fs"

var FS fs.FS

func SetFS(fsys fs.FS) {
	FS = fsys
}
