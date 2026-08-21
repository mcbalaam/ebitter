package tiled

import (
	"strings"

	gotiled "github.com/lafriks/go-tiled"
	"github.com/mcbalaam/ebitter/pkg/systems"
)

// TriggerMode selects how an InteractionZone fires.
type TriggerMode int

const (
	// TriggerTouch fires once when the player enters the zone.
	TriggerTouch TriggerMode = iota
	// TriggerButton fires when the player is inside the zone and presses the
	// interaction key. It can only fire again after the player has left.
	TriggerButton
)

// InteractionZone is a trigger area that emits a signal on the master signal
// bus when activated.
type InteractionZone struct {
	Name    string
	Signal  string
	Trigger TriggerMode
	Rect    ColliderBox

	active bool
}

// Check evaluates the zone against the given collider. It emits the zone's
// signal on the master signal bus when activated and reports whether it fired
// this call.
func (z *InteractionZone) Check(box *ColliderBox, interactPressed bool, source interface{}) bool {
	if z == nil || box == nil {
		return false
	}

	overlapping := z.Rect.Overlaps(box)

	switch z.Trigger {
	case TriggerTouch:
		if overlapping && !z.active {
			z.active = true
			return z.emit(source)
		}
		if !overlapping {
			z.active = false
		}

	case TriggerButton:
		if overlapping && interactPressed && !z.active {
			z.active = true
			return z.emit(source)
		}
		if !overlapping {
			z.active = false
		}
	}

	return false
}

func (z *InteractionZone) emit(source interface{}) bool {
	systems.MasterSignalBus.Emit(z.Signal, source, z)
	return true
}

// collectInteractions builds interaction zones from Tiled object groups.
// A group contributes its objects when its name or class contains
// "interaction" or it has a string "interaction" property set.
func collectInteractions(allGroups []*gotiled.ObjectGroup) []*InteractionZone {
	var out []*InteractionZone
	for _, g := range allGroups {
		if g == nil {
			continue
		}
		isInteraction := strings.Contains(strings.ToLower(g.Name), "interact") ||
			strings.Contains(strings.ToLower(g.Class), "interact")

		for _, obj := range g.Objects {
			if obj.Width <= 0 || obj.Height <= 0 {
				continue
			}

			zone := &InteractionZone{
				Name:   obj.Name,
				Signal: obj.Properties.GetString("signal"),
				Rect: ColliderBox{
					X: obj.X,
					Y: obj.Y,
					W: obj.Width,
					H: obj.Height,
				},
			}

			if !isInteraction {
				interact := obj.Properties.GetString("interaction")
				if interact == "" && zone.Signal == "" {
					continue
				}
				if zone.Name == "" {
					zone.Name = interact
				}
			}
		if zone.Name == "" {
			zone.Name = obj.Properties.GetString("name")
		}
		if zone.Name == "" {
			zone.Name = obj.Class
		}
			if zone.Name == "" {
				zone.Name = obj.Type
			}
			if zone.Signal == "" {
				zone.Signal = "interact"
			}

			switch strings.ToLower(obj.Properties.GetString("trigger")) {
			case "button":
				zone.Trigger = TriggerButton
			default:
				zone.Trigger = TriggerTouch
			}

			out = append(out, zone)
		}
	}
	return out
}
