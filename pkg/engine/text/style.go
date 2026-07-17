package text

import "image/color"

type TextStyle struct {
	FontName     string
	StartX       float64
	StartY       float64
	ScaleX       float64
	ScaleY       float64
	FontHeight   float64
	LineSpacing  float64
	DefaultDelay float64
	Instant      bool
	CharSpacing  float64
	Color        color.Color
}

func (s TextStyle) WithInstant(instant bool) TextStyle {
	s.Instant = instant
	return s
}
