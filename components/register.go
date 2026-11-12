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
	registerBuiltin("body", orderBody, decode(func() *Body { return &Body{} }), strictConfig(func() *Body { return &Body{} }))
	registerBuiltin("velocity", orderVelocity, decode(func() *Velocity { return &Velocity{} }), strictConfig(func() *Velocity { return &Velocity{} }))
	registerBuiltin("transform", orderTransform, newTransformFactory(), strictConfig(func() *transformConfig { return &transformConfig{} }))
	registerBuiltin("animator", orderAnimator, decode(func() *Animator { return &Animator{} }), strictConfig(func() *Animator { return &Animator{} }))
	registerBuiltin("sprite", orderSprite, newSpriteFactory(), vectorFields(strictConfig(func() *spriteConfig { return &spriteConfig{} }), "offset", "anchor"))
	registerBuiltin("hitbox", orderHitbox, newHitboxFactory(), vectorFields(strictConfig(func() *hitboxConfig { return &hitboxConfig{} }), "offset"))
	registerBuiltin("trigger", orderTrigger, newTriggerFactory(), vectorFields(strictConfig(func() *triggerCfg { return &triggerCfg{} }), "offset"))
	registerBuiltin("interactable", orderInteractable, newInteractableFactory(), vectorFields(strictConfig(func() *interactableConfig { return &interactableConfig{} }), "offset"))
	registerBuiltin("lifetime", orderLifetime, decode(func() *Lifetime { return &Lifetime{} }), strictConfig(func() *Lifetime { return &Lifetime{} }))
}

func registerBuiltin(name string, order int, factory Factory, validator Validator) {
	if err := Register(name, order, factory, validator); err != nil {
		panic(err)
	}
}

type transformConfig struct {
	X        float64 `yaml:"x"`
	Y        float64 `yaml:"y"`
	Scale    float64 `yaml:"scale"`
	ScaleX   float64 `yaml:"scaleX"`
	ScaleY   float64 `yaml:"scaleY"`
	Rotation float64 `yaml:"rotation"`
}

func newTransformFactory() Factory {
	return func(node *yaml.Node) (Component, error) {
		var c transformConfig
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

type spriteConfig struct {
	Icon   string    `yaml:"icon"`
	State  string    `yaml:"state"`
	Offset []float64 `yaml:"offset"`
	Anchor []float64 `yaml:"anchor"`
	FlipH  bool      `yaml:"flipH"`
	FlipV  bool      `yaml:"flipV"`
}

func newSpriteFactory() Factory {
	return func(node *yaml.Node) (Component, error) {
		var c spriteConfig
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

type hitboxConfig struct {
	W      float64   `yaml:"w"`
	H      float64   `yaml:"h"`
	Offset []float64 `yaml:"offset"`
	Layer  string    `yaml:"layer"`
	Mask   []string  `yaml:"mask"`
	Solid  bool      `yaml:"solid"`
}

func newHitboxFactory() Factory {
	return func(node *yaml.Node) (Component, error) {
		var c hitboxConfig
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
	return func(node *yaml.Node) (Component, error) {
		var c interactableConfig
		if err := decodeYAML(node, &c); err != nil {
			return nil, err
		}
		trigger := triggerCfg{W: c.W, H: c.H, Offset: c.Offset, Signal: c.Signal, Mode: c.Mode, Once: c.Once}
		return NewInteractable(trigger.build(), c.Prompt), nil
	}
}

type interactableConfig struct {
	W      float64   `yaml:"w"`
	H      float64   `yaml:"h"`
	Offset []float64 `yaml:"offset"`
	Signal string    `yaml:"signal"`
	Mode   string    `yaml:"mode"`
	Once   bool      `yaml:"once"`
	Prompt string    `yaml:"prompt"`
}
