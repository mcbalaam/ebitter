package components

import (
	"bytes"
	"fmt"
	"time"

	enginecomponents "github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/tiled"
	"gopkg.in/yaml.v3"
)

type PlayerController struct {
	Speed   float64 `yaml:"speed"`
	dx, dy  float64
	facing  string
	solids  []*tiled.ColliderBox
	pressed map[string]bool
}

type playerInput struct {
	Key     string
	Pressed bool
}

func (p *PlayerController) Type() string { return "playerController" }

func (p *PlayerController) Init(e *enginecomponents.Entity) {
	p.pressed = map[string]bool{}
	e.Signals().Subscribe("player.input", e, func(sig events.Signal) {
		if sig.Source != e {
			return
		}
		if input, ok := sig.Data.(playerInput); ok {
			p.pressed[input.Key] = input.Pressed
		}
	})
}

func (p *PlayerController) Update(e *enginecomponents.Entity, dt time.Duration) {
	p.dx, p.dy = 0, 0
	if p.pressed["left"] {
		p.dx--
	}
	if p.pressed["right"] {
		p.dx++
	}
	if p.pressed["up"] {
		p.dy--
	}
	if p.pressed["down"] {
		p.dy++
	}
	if p.Speed == 0 {
		p.Speed = 160
	}
	t := e.Transform()
	h := e.Hitbox()
	if t == nil || h == nil {
		return
	}
	if p.dx < 0 {
		p.facing = "left"
	} else if p.dx > 0 {
		p.facing = "right"
	} else if p.dy < 0 {
		p.facing = "up"
	} else if p.dy > 0 {
		p.facing = "down"
	} else if p.facing == "" {
		p.facing = "down"
	}
	state := p.facing
	if p.dx != 0 || p.dy != 0 {
		state += "_walk"
	}
	if s := e.Sprite(); s != nil && s.Icon != nil {
		_ = s.Icon.SetIconState(state)
	}
	seconds := dt.Seconds()
	p.moveAxis(e, 0, p.dx*p.Speed*seconds)
	p.moveAxis(e, 1, p.dy*p.Speed*seconds)
}

func (p *PlayerController) Movement() (float64, float64) { return p.dx, p.dy }

func (p *PlayerController) Facing() string { return p.facing }

func (p *PlayerController) SetSolids(solids []*tiled.ColliderBox) { p.solids = solids }

func (p *PlayerController) moveAxis(e *enginecomponents.Entity, axis int, delta float64) {
	t, h := e.Transform(), e.Hitbox()
	if delta == 0 || t == nil || h == nil || h.Collider == nil {
		return
	}
	if axis == 0 {
		t.X += delta
	} else {
		t.Y += delta
	}
	w := h.Width * t.ScaleX
	height := h.Height * t.ScaleY
	offX, offY := h.OffsetX*t.ScaleX, h.OffsetY*t.ScaleY
	for _, c := range p.solids {
		box := tiled.ColliderBox{X: t.X + offX - w/2, Y: t.Y + offY - height/2, W: w, H: height}
		if !box.Overlaps(c) {
			continue
		}
		if axis == 0 {
			if delta > 0 {
				t.X = c.X - w/2 - offX
			} else {
				t.X = c.X + c.W + w/2 - offX
			}
		} else {
			if delta > 0 {
				t.Y = c.Y - height/2 - offY
			} else {
				t.Y = c.Y + c.H + height/2 - offY
			}
		}
	}
}

func registerPlayerController() {
	err := enginecomponents.Register("playerController", 5,
		func(node *yaml.Node) (enginecomponents.Component, error) {
			p := &PlayerController{Speed: 160}
			if node != nil {
				if err := node.Decode(p); err != nil {
					return nil, err
				}
			}
			return p, nil
		}, func(node *yaml.Node) error {
			if node == nil {
				return nil
			}
			if node.Kind != yaml.MappingNode {
				return fmt.Errorf("expected mapping")
			}
			copyNode := *node
			copyNode.Content = make([]*yaml.Node, 0, len(node.Content))
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value == "order" {
					continue
				}
				copyNode.Content = append(copyNode.Content, node.Content[i], node.Content[i+1])
			}
			data, err := yaml.Marshal(&copyNode)
			if err != nil {
				return err
			}
			dec := yaml.NewDecoder(bytes.NewReader(data))
			dec.KnownFields(true)
			return dec.Decode(&struct {
				Speed float64 `yaml:"speed"`
			}{})
		})
	if err != nil {
		panic(err)
	}
}

// EmitInput sends a directional key state to the player's controller.
func EmitInput(e *enginecomponents.Entity, key string, pressed bool) {
	e.Signals().Emit("player.input", e, playerInput{Key: key, Pressed: pressed})
}

func init() { registerPlayerController() }
