package text

import (
	"image/color"
	"testing"
)

func TestParserMarkupCommands(t *testing.T) {
	p := &TextParser{
		Text:         "Hi $p500$cff0000!$n2",
		StartX:       10,
		StartY:       20,
		ScaleX:       1,
		ScaleY:       1,
		FontHeight:   24,
		LineSpacing:  10,
		Delay:        0.1,
		CharWidth:    make(map[string]int),
		DefaultColor: color.White,
	}
	cmds := p.Parse()

	if len(cmds) != 5 {
		t.Fatalf("expected 5 commands, got %d: %+v", len(cmds), cmds)
	}
	if cmds[0].Char != "H" || cmds[0].X != 10 || cmds[0].Y != 20 {
		t.Errorf("first char misplaced: %+v", cmds[0])
	}
	if cmds[3].Char != "!" || cmds[3].Color != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Errorf("color not applied: %+v", cmds[3])
	}
	if cmds[4].Char != "2" || cmds[4].Y <= 20 {
		t.Errorf("newline not applied: %+v", cmds[4])
	}
}

func TestParserInvalidColorKeepsPrevious(t *testing.T) {
	p := &TextParser{
		Text:         "A$cZZZZZZB",
		StartX:       0,
		StartY:       0,
		ScaleX:       1,
		ScaleY:       1,
		Delay:        0,
		CharWidth:    make(map[string]int),
		DefaultColor: color.White,
	}
	cmds := p.Parse()
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0].Char != "A" || cmds[1].Char != "B" {
		t.Fatalf("unexpected chars: %+v, %+v", cmds[0], cmds[1])
	}
	if cmds[1].Color != color.White {
		t.Errorf("invalid hex should keep previous color, got %v", cmds[1].Color)
	}
}

func TestParserPauseAdvancesTrigger(t *testing.T) {
	p := &TextParser{
		Text:         "AB$p1000C",
		StartX:       0,
		StartY:       0,
		ScaleX:       1,
		ScaleY:       1,
		Delay:        0.1,
		CharWidth:    make(map[string]int),
		DefaultColor: color.White,
	}
	cmds := p.Parse()
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[2].TriggerAt < 1.1 {
		t.Errorf("pause not applied to trigger time: %f", cmds[2].TriggerAt)
	}
}

func TestParserEndMarkers(t *testing.T) {
	p := &TextParser{
		Text:         "A$eB$f",
		StartX:       0,
		StartY:       0,
		ScaleX:       1,
		ScaleY:       1,
		Delay:        0,
		CharWidth:    make(map[string]int),
		DefaultColor: color.White,
	}
	cmds := p.Parse()
	if len(cmds) != 4 {
		t.Fatalf("expected 4 commands, got %d: %+v", len(cmds), cmds)
	}
	if cmds[0].Type != CmdChar || cmds[1].Type != CmdEnd || cmds[2].Type != CmdChar || cmds[3].Type != CmdEndNoWait {
		t.Errorf("unexpected command types: %+v", cmds)
	}
}
