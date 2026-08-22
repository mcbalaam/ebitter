//go:build !wasm

package assetfs

import "os"

func init() {
	FS = os.DirFS(".")
}
