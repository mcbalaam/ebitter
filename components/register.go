package components

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mcbalaam/ebitter/render"
)

// Built-in default update/draw order. Lower values run first. The order is
// overridable per component through a blueprint's optional "order" field.
//
//	body        -20  # forces before integration
//	velocity    -10  # integrates into transform
//	transform     0
//	animator     10  # switches sprite before it draws
//	sprite       20
//	hitbox       30
//	trigger      40
//	interactable 50
//	lifetime    100  # last
const (
	orderBody         = -20
	orderVelocity     = -10
	orderTransform    = 0
	orderAnimator     = 10
	orderSprite       = 20
	orderHitbox       = 30
	orderTrigger      = 40
	orderInteractable = 50
	orderLifetime     = 100
)

func init() {
	Register("body", orderBody, decode(func() *Body { return &Body{} }))
	Register("velocity", orderVelocity, decode(func() *Velocity { return &Velocity{} }))
	Register("transform", orderTransform, newTransformFactory())
	Register("animator", orderAnimator, decode(func() *Animator { return &Animator{} }))
	Register("sprite", orderSprite, newSpriteFactory())
	Register("hitbox", orderHitbox, newHitboxFactory())
	Register("trigger", orderTrigger, newTriggerFactory())
	Register("interactable", orderInteractable, newInteractableFactory())
	Register("lifetime", orderLifetime, decode(func() *Lifetime { return &Lifetime{} }))
}

func newTransformFactory() Factory {
	type cfg struct {
		X, Y     float64
		Scale    float64
		ScaleX   float64
		ScaleY   float64
		Rotation float64
	}
	return func(node *yaml.Node) (Component, error) {
		var c cfg
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		sx, sy := c.ScaleX, c.ScaleY
		if c.Scale != 0 {
			if sx == 0 {
				sx = c.Scale
			}
			if sy == 0 {
				sy = c.Scale
			}
		}
		if sx == 0 {
			sx = 1
		}
		if sy == 0 {
			sy = 1
		}
		return &Transform{X: c.X, Y: c.Y, ScaleX: sx, ScaleY: sy, Rotation: c.Rotation}, nil
	}
}

func newSpriteFactory() Factory {
	type cfg struct {
		Icon   string
		State  string
		Offset []float64
		Anchor []float64
		FlipH  bool
		FlipV  bool
	}
	return func(node *yaml.Node) (Component, error) {
		var c cfg
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		s := &Sprite{FlipH: c.FlipH, FlipV: c.FlipV}
		if len(c.Offset) == 2 {
			s.OffsetX, s.OffsetY = c.Offset[0], c.Offset[1]
		}
		if len(c.Anchor) == 2 {
			s.AnchorX, s.AnchorY = c.Anchor[0], c.Anchor[1]
		}
		if c.Icon != "" {
			icon, err := render.NewAnimatedIconFromPath(spritePath(c.Icon), c.State)
			if err != nil {
				return nil, err
			}
			s.Icon = icon
		}
		return s, nil
	}
}

func newHitboxFactory() Factory {
	type cfg struct {
		W, H   float64
		Offset []float64
		Layer  string
		Mask   []string
		Solid  bool
	}
	return func(node *yaml.Node) (Component, error) {
		var c cfg
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		var ox, oy float64
		if len(c.Offset) == 2 {
			ox, oy = c.Offset[0], c.Offset[1]
		}
		h := NewHitbox(c.W, c.H, ox, oy)
		h.Layer = c.Layer
		h.Mask = c.Mask
		h.Solid = c.Solid
		return h, nil
	}
}

type triggerCfg struct {
	W      float64   `yaml:"w"`
	H      float64   `yaml:"h"`
	Offset []float64 `yaml:"offset"`
	Signal string    `yaml:"signal"`
	Mode   string    `yaml:"mode"`
	Once   bool      `yaml:"once"`
}

func (c triggerCfg) build() *Trigger {
	var ox, oy float64
	if len(c.Offset) == 2 {
		ox, oy = c.Offset[0], c.Offset[1]
	}
	return &Trigger{
		W: c.W, H: c.H,
		OffsetX: ox, OffsetY: oy,
		Signal: c.Signal,
		Mode:   strings.ToLower(c.Mode),
		Once:   c.Once,
	}
}

func newTriggerFactory() Factory {
	return func(node *yaml.Node) (Component, error) {
		var c triggerCfg
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		return c.build(), nil
	}
}

func newInteractableFactory() Factory {
	type cfg struct {
		Trigger triggerCfg `yaml:",inline"`
		Prompt  string     `yaml:"prompt"`
	}
	return func(node *yaml.Node) (Component, error) {
		var c cfg
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		return NewInteractable(c.Trigger.build(), c.Prompt), nil
	}
}
