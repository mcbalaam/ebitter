//go:build !wasm

package main

import "github.com/mcbalaam/ebitter/pkg/assets"

func init() {
	assets.ProcessFonts()
}
