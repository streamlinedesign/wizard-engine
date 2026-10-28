package hud

import (
	"os"
	"testing"

	"github.com/streamlinedesign/wizard-engine/game/items"
	"github.com/streamlinedesign/wizard-engine/game/spells"
)

func load(t *testing.T) (*items.Registry, *spells.Table) {
	id, err := os.ReadFile("../../assets/data/items.json")
	if err != nil {
		t.Fatal(err)
	}
	reg, err := items.ParseRegistry(id)
	if err != nil {
		t.Fatal(err)
	}
	sd, err := os.ReadFile("../../assets/data/spells.json")
	if err != nil {
		t.Fatal(err)
	}
	tb, err := spells.ParseTable(sd)
	if err != nil {
		t.Fatal(err)
	}
	return reg, tb
}

func TestBindingPlaceAndConfirm(t *testing.T) {
	reg, tb := load(t)
	var inv items.Inventory
	inv.Add(reg.Must("scroll_empty"), 2)
	inv.Add(reg.Must("spell_fireball"), 1)

	b := NewBinding(nil, tb)
	var bound []spells.Spell
	b.OnBind = func(sp spells.Spell) { bound = append(bound, sp) }

	b.confirm(&inv)
	if len(bound) != 0 || b.message == "" {
		t.Fatal("empty circle must not bind")
	}
	b.place(&inv, reg.Must("spell_fireball"))
	b.place(&inv, reg.Must("scroll_empty"))
	if b.Center == nil || b.Ring[0] == nil || inv.Count(reg.Must("scroll_empty")) != 1 || inv.Has(reg.Must("spell_fireball")) {
		t.Fatal("placing should move items out of the inventory")
	}
	sp, ok := b.Preview()
	if !ok || sp.Name != "Fireball" {
		t.Fatalf("preview %v %v", sp, ok)
	}
	b.confirm(&inv)
	if len(bound) != 1 || bound[0].Name != "Fireball" || b.Center != nil || b.Ring[0] != nil {
		t.Fatal("confirm should bind and consume")
	}
	// A second scroll cannot go in a full centre; Clear gives things back.
	b.place(&inv, reg.Must("scroll_empty"))
	b.place(&inv, reg.Must("scroll_empty"))
	if inv.Count(reg.Must("scroll_empty")) != 0 || b.Center == nil {
		t.Fatal("second scroll should stay in the inventory")
	}
	b.Clear(&inv)
	if inv.Count(reg.Must("scroll_empty")) != 1 {
		t.Fatal("Clear should return the scroll")
	}
}

func TestBindingRejectsKeywordlessItems(t *testing.T) {
	reg, tb := load(t)
	var inv items.Inventory
	inv.Add(reg.Must("coin"), 5)
	b := NewBinding(nil, tb)
	b.place(&inv, reg.Must("coin"))
	if b.Ring[0] != nil || inv.Count(reg.Must("coin")) != 5 {
		t.Fatal("coins have no keywords and must not be placed")
	}
}
