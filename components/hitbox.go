package components

import "time"

// Hitbox is an axis-aligned SAT collision box attached to an entity. It embeds
// Collider for the geometry and adds data-driven layer/mask filtering.
type Hitbox struct {
	*Collider
	Layer string
	Mask  []string
	Solid bool
}

// NewHitbox creates a hitbox of the given size centred on the entity's
// transform plus (offsetX, offsetY).
func NewHitbox(width, height, offsetX, offsetY float64) *Hitbox {
	return &Hitbox{Collider: NewCollider(width, height, offsetX, offsetY)}
}

func (h *Hitbox) Type() string { return "hitbox" }

func (h *Hitbox) Update(e *Entity, dt time.Duration) {
	if h.Collider == nil {
		return
	}
	if t := e.Transform(); t != nil {
		h.Collider.UpdateWorldVerts(t)
	}
}

// AABB returns the hitbox's world-space bounding box.
func (h *Hitbox) AABB() AABB {
	if h == nil || h.Collider == nil {
		return AABB{}
	}
	minX, minY, maxX, maxY := h.Collider.GetBounds()
	return AABB{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

// MatchesLayer reports whether this hitbox and other could collide under their
// layer/mask rules. An empty mask matches any layer.
func (h *Hitbox) MatchesLayer(other *Hitbox) bool {
	if h == nil || other == nil {
		return false
	}
	return layerAllowed(h.Layer, other.Mask) && layerAllowed(other.Layer, h.Mask)
}

func layerAllowed(layer string, mask []string) bool {
	if len(mask) == 0 {
		return true
	}
	for _, m := range mask {
		if m == layer {
			return true
		}
	}
	return false
}
