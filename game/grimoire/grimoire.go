// Package grimoire is the player's spellbook: the spells bound so far and
// which four sit on the quick slots.
package grimoire

import "github.com/streamlinedesign/wizard-engine/game/spells"

// Book holds bound spells. Owner is the name on the cover.
type Book struct {
	Owner  string
	Spells []spells.Spell
	// Slots index into Spells; -1 is empty.
	Slots [4]int
	// Selected is the active quick slot.
	Selected int
}

// New makes an empty book with every slot free.
func New(owner string) *Book {
	return &Book{Owner: owner, Slots: [4]int{-1, -1, -1, -1}}
}

// Bind adds a spell and drops it into the first free slot, if any. It
// returns the spell's index.
func (b *Book) Bind(sp spells.Spell) int {
	b.Spells = append(b.Spells, sp)
	idx := len(b.Spells) - 1
	for i, s := range b.Slots {
		if s < 0 {
			b.Slots[i] = idx
			break
		}
	}
	return idx
}

// Active returns the spell on the selected slot, or nil.
func (b *Book) Active() *spells.Spell {
	return b.Slot(b.Selected)
}

// Slot returns the spell on a quick slot, or nil.
func (b *Book) Slot(i int) *spells.Spell {
	if i < 0 || i >= len(b.Slots) || b.Slots[i] < 0 || b.Slots[i] >= len(b.Spells) {
		return nil
	}
	return &b.Spells[b.Slots[i]]
}

// Assign puts a bound spell on a quick slot.
func (b *Book) Assign(slot, spell int) {
	if slot < 0 || slot >= len(b.Slots) || spell < 0 || spell >= len(b.Spells) {
		return
	}
	b.Slots[slot] = spell
}
