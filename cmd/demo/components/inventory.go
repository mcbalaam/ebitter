package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

type InventoryEntry struct {
	Prototype string `yaml:"prototype"`
	Count     int    `yaml:"count"`
}

// Inventory stores stack counts keyed by item prototype ID.
type Inventory struct {
	Items []InventoryEntry `yaml:"items"`
}

func (*Inventory) Type() string { return "inventory" }

func (i *Inventory) Count(prototype string) int {
	for _, entry := range i.Items {
		if entry.Prototype == prototype {
			return entry.Count
		}
	}
	return 0
}

func (i *Inventory) Add(prototype string, count int) {
	if prototype == "" || count <= 0 {
		return
	}
	for n := range i.Items {
		if i.Items[n].Prototype == prototype {
			i.Items[n].Count += count
			return
		}
	}
	i.Items = append(i.Items, InventoryEntry{Prototype: prototype, Count: count})
}

func (i *Inventory) Remove(prototype string, count int) bool {
	if count <= 0 {
		return false
	}
	for n := range i.Items {
		if i.Items[n].Prototype != prototype || i.Items[n].Count < count {
			continue
		}
		i.Items[n].Count -= count
		if i.Items[n].Count == 0 {
			i.Items = append(i.Items[:n], i.Items[n+1:]...)
		}
		return true
	}
	return false
}

func init() {
	registerBattleComponent("inventory", 10, func() *Inventory { return &Inventory{} }, func(inventory *Inventory) error {
		seen := make(map[string]struct{}, len(inventory.Items))
		for _, entry := range inventory.Items {
			if entry.Prototype == "" || entry.Count <= 0 {
				return fmt.Errorf("inventory entries need a prototype and positive count")
			}
			if _, ok := seen[entry.Prototype]; ok {
				return fmt.Errorf("duplicate inventory item %q", entry.Prototype)
			}
			seen[entry.Prototype] = struct{}{}
		}
		return nil
	})
}

var _ components.Component = (*Inventory)(nil)
