package engine

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

var (
	StyleNarrative = TextStyle{
		FontName:     "determination",
		StartX:       60.0,
		StartY:       732.0,
		ScaleX:       0.5,
		ScaleY:       0.5,
		FontHeight:   24.0,
		LineSpacing:  90.0,
		DefaultDelay: 0.03,
		Instant:      false,
		CharSpacing:  2.0,
	}
)
