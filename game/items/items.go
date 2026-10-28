// Package items defines what an item is and holds the player's inventory.
// Definitions come from assets/data/items.json so adding an item is a data
// change. Keywords are the hook for the spell system in step 3: an item's
// keywords are what a bound scroll turns into a spell.
package items

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Kind groups items for UI and rules.
type Kind string

const (
	KindSpell    Kind = "spell"    // a spell item: bindable into the grimoire
	KindScroll   Kind = "scroll"   // empty or bound scroll
	KindBook     Kind = "book"     // the grimoire and other books
	KindMaterial Kind = "material" // crafting ingredients
	KindMisc     Kind = "misc"
)

// Def is one item definition.
type Def struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        Kind     `json:"kind"`
	Icon        int      `json:"icon"` // 1-based tile id in icons/items.png
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
	// Stackable items share one slot with a count.
	Stackable bool `json:"stackable"`
	// Legendary items cast as spells without a scroll (step 3).
	Legendary bool `json:"legendary"`
}

// Registry is every known item by id.
type Registry struct {
	defs map[string]*Def
}

// ParseRegistry decodes the items.json format: a top-level "items" array.
func ParseRegistry(data []byte) (*Registry, error) {
	var file struct {
		Items []*Def `json:"items"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("items.json: %w", err)
	}
	r := &Registry{defs: map[string]*Def{}}
	for _, d := range file.Items {
		if d.ID == "" {
			return nil, fmt.Errorf("items.json: item with empty id")
		}
		if _, dup := r.defs[d.ID]; dup {
			return nil, fmt.Errorf("items.json: duplicate id %q", d.ID)
		}
		r.defs[d.ID] = d
	}
	return r, nil
}

// Get looks up a definition; ok is false for unknown ids.
func (r *Registry) Get(id string) (*Def, bool) {
	d, ok := r.defs[id]
	return d, ok
}

// Must looks up a definition or panics; for ids written in code.
func (r *Registry) Must(id string) *Def {
	d, ok := r.defs[id]
	if !ok {
		panic("unknown item id " + id)
	}
	return d
}

// Len is the number of definitions.
func (r *Registry) Len() int { return len(r.defs) }

// Stack is a slot in an inventory.
type Stack struct {
	Def   *Def
	Count int
}

// Inventory is an ordered list of stacks. Stackable items merge; others
// each take a slot. Order is insertion order so the UI is stable.
type Inventory struct {
	stacks []*Stack
}

// Add puts count of an item in, merging into an existing stack when the
// item is stackable.
func (inv *Inventory) Add(d *Def, count int) {
	if count <= 0 {
		return
	}
	if d.Stackable {
		for _, s := range inv.stacks {
			if s.Def == d {
				s.Count += count
				return
			}
		}
		inv.stacks = append(inv.stacks, &Stack{Def: d, Count: count})
		return
	}
	for i := 0; i < count; i++ {
		inv.stacks = append(inv.stacks, &Stack{Def: d, Count: 1})
	}
}

// Remove takes count of an item out; it reports false and changes nothing
// if there are not enough.
func (inv *Inventory) Remove(d *Def, count int) bool {
	if inv.Count(d) < count {
		return false
	}
	left := count
	kept := inv.stacks[:0]
	for _, s := range inv.stacks {
		if s.Def != d || left == 0 {
			kept = append(kept, s)
			continue
		}
		take := s.Count
		if take > left {
			take = left
		}
		s.Count -= take
		left -= take
		if s.Count > 0 {
			kept = append(kept, s)
		}
	}
	inv.stacks = kept
	return true
}

// Count totals an item across stacks.
func (inv *Inventory) Count(d *Def) int {
	n := 0
	for _, s := range inv.stacks {
		if s.Def == d {
			n += s.Count
		}
	}
	return n
}

// Has reports whether at least one of the item is held.
func (inv *Inventory) Has(d *Def) bool { return inv.Count(d) > 0 }

// Stacks returns the slots in display order.
func (inv *Inventory) Stacks() []*Stack { return inv.stacks }

// ByKind returns the slots of one kind, in display order.
func (inv *Inventory) ByKind(k Kind) []*Stack {
	var out []*Stack
	for _, s := range inv.stacks {
		if s.Def.Kind == k {
			out = append(out, s)
		}
	}
	return out
}

// IDs lists held item ids, sorted, for saving and tests.
func (inv *Inventory) IDs() []string {
	seen := map[string]bool{}
	var ids []string
	for _, s := range inv.stacks {
		if !seen[s.Def.ID] {
			seen[s.Def.ID] = true
			ids = append(ids, s.Def.ID)
		}
	}
	sort.Strings(ids)
	return ids
}
