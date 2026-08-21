// Package assetfs provides a cross-platform filesystem for game assets.
// Desktop: reads from the real filesystem via os.DirFS.
// WASM: reads from embedded data set by the main package.
package assetfs

import "io/fs"

var FS fs.FS

func SetFS(fsys fs.FS) {
	FS = fsys
}
