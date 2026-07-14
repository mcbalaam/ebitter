package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Input struct {
	justPressed map[ebiten.Key]struct{}
}

func NewInput() *Input {
	return &Input{justPressed: make(map[ebiten.Key]struct{})}
}

func (in *Input) Update() {
	for k := range in.justPressed {
		delete(in.justPressed, k)
	}
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		in.justPressed[k] = struct{}{}
	}
}

func (in *Input) JustPressed(key ebiten.Key) bool {
	_, ok := in.justPressed[key]
	return ok
}
