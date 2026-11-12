package components

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	enginecomponents "github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/text"
	"gopkg.in/yaml.v3"
)

// IconStateLabel draws the current Aseprite state above its entity.
type IconStateLabel struct {
	label   *text.TextString
	shown   string
	visible bool
}

func (*IconStateLabel) Type() string { return "iconStateLabel" }

func (c *IconStateLabel) Visible() bool { return c.visible }

func (c *IconStateLabel) Init(e *enginecomponents.Entity) {
	c.visible = false
	c.label = text.NewTextString("", 0, 0, text.TextStyle{
		FontName:    "mars",
		ScaleX:      0.22,
		ScaleY:      0.22,
		FontHeight:  130,
		LineSpacing: 10,
		Color:       color.White,
	})
	events.MasterSignalBus.Subscribe("demo.iconStateLabel.toggle", e, func(sig events.Signal) {
		if sig.Source == e {
			c.visible = !c.visible
		}
	})
}

func (c *IconStateLabel) Update(*enginecomponents.Entity, time.Duration) {}

func (c *IconStateLabel) Draw(e *enginecomponents.Entity, screen *ebiten.Image) {
	if !c.visible || c.label == nil || e.Sprite() == nil || e.Sprite().Icon == nil || e.Sprite().Icon.CurrentState == nil {
		return
	}
	state := e.Sprite().Icon.CurrentState.Name
	if state == "" {
		return
	}
	if state != c.shown {
		c.shown = state
		c.label.SetText(state)
	}
	if t := e.Transform(); t != nil {
		c.label.SetPosition(t.X+30, t.Y-60)
		c.label.Draw(screen)
	}
}

// ToggleIconStateLabel toggles the state label attached to e, if present.
func ToggleIconStateLabel(e *enginecomponents.Entity) {
	events.MasterSignalBus.Emit("demo.iconStateLabel.toggle", e)
}

func registerIconStateLabel() {
	err := enginecomponents.Register("iconStateLabel", 40,
		func(node *yaml.Node) (enginecomponents.Component, error) {
			return &IconStateLabel{}, nil
		}, func(node *yaml.Node) error {
			if node == nil {
				return nil
			}
			if node.Kind != yaml.MappingNode {
				return fmt.Errorf("expected mapping")
			}
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value != "order" {
					return fmt.Errorf("unknown field %q", node.Content[i].Value)
				}
			}
			return nil
		})
	if err != nil {
		panic(err)
	}
}

func init() { registerIconStateLabel() }
