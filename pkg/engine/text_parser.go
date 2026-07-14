package engine

import (
	"image/color"
	"strconv"
)

type CommandType int

const (
	CmdChar CommandType = iota
	CmdEnd
	CmdEndNoWait
)

type DialogueCommand struct {
	Type      CommandType
	Char      string
	Color     color.Color
	X         float64
	Y         float64
	TriggerAt float64
}

type TextParser struct {
	Text         string
	StartX       float64
	StartY       float64
	ScaleX       float64
	ScaleY       float64
	FontHeight   float64
	LineSpacing  float64
	Delay        float64
	CharWidth    map[string]int
	CharSpacing  float64
	DefaultColor color.Color
}

func (p *TextParser) Parse() []DialogueCommand {
	runes := []rune(p.Text)
	var cmds []DialogueCommand

	currentTime := 0.0
	curX := p.StartX
	curY := p.StartY
	curColor := p.DefaultColor
	if curColor == nil {
		curColor = color.White
	}
	curDelay := p.Delay

	i := 0
	for i < len(runes) {
		if runes[i] == '$' && i+1 < len(runes) {
			cmdChar := runes[i+1]
			i += 2

			switch cmdChar {
			case 'p':
				valStr := ""
				for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
					valStr += string(runes[i])
					i++
				}
				ms, _ := strconv.Atoi(valStr)
				currentTime += float64(ms) / 1000.0

			case 'c':
				if i+6 <= len(runes) {
					hexStr := string(runes[i : i+6])
					i += 6
					r, _ := strconv.ParseUint(hexStr[0:2], 16, 8)
					g, _ := strconv.ParseUint(hexStr[2:4], 16, 8)
					b, _ := strconv.ParseUint(hexStr[4:6], 16, 8)
					curColor = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
				}

			case 's':
				valStr := ""
				for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
					valStr += string(runes[i])
					i++
				}
				ms, _ := strconv.Atoi(valStr)
				curDelay = float64(ms) / 1000.0

			case 'n':
				curX = p.StartX
				curY += (p.FontHeight + p.LineSpacing) * p.ScaleY

			case 'e':
				cmds = append(cmds, DialogueCommand{
					Type:      CmdEnd,
					TriggerAt: currentTime,
				})
			case 'f':
				cmds = append(cmds, DialogueCommand{
					Type:      CmdEndNoWait,
					TriggerAt: currentTime,
				})
			}
			continue
		}

		char := string(runes[i])
		charWidth := p.CharWidth[char]
		if charWidth == 0 {
			charWidth = 20
		}

		extraDelay := 0.0
		if char == "." || char == "," {
			extraDelay = curDelay * 10
		}

		cmds = append(cmds, DialogueCommand{
			Type:      CmdChar,
			Char:      char,
			Color:     curColor,
			X:         curX,
			Y:         curY,
			TriggerAt: currentTime,
		})

		currentTime += curDelay + extraDelay
		spacing := p.CharSpacing
		if spacing == 0 {
			spacing = 2
		}
		curX += ((float64(charWidth) + spacing) * 3) * p.ScaleX
		i++
	}

	return cmds
}
