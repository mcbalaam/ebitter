//go:build !wasm

package embedfs

import "os"

func init() {
	FS = os.DirFS(".")
}
