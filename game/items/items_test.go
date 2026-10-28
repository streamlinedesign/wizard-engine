package items

import "testing"

const testJSON = `{"items":[
 {"id":"scroll_empty","name":"Empty scroll","kind":"scroll","icon":1,"stackable":true},
 {"id":"spell_fireball","name":"Fireball","kind":"spell","icon":3,"keywords":["fire","ball"]},
 {"id":"grimoire","name":"Grimoire","kind":"book","icon":7}
]}`

func reg(t *testing.T) *Registry {
	r, err := ParseRegistry([]byte(testJSON))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestParseRegistry(t *testing.T) {
	r := reg(t)
	if r.Len() != 3 {
		t.Fatalf("len %d", r.Len())
	}
	fb := r.Must("spell_fireball")
	if fb.Kind != KindSpell || len(fb.Keywords) != 2 {
		t.Fatalf("fireball parsed wrong: %+v", fb)
	}
	if _, ok := r.Get("nope"); ok {
		t.Fatal("unknown id should not be found")
	}
}

func TestParseRejectsDuplicates(t *testing.T) {
	_, err := ParseRegistry([]byte(`{"items":[{"id":"a"},{"id":"a"}]}`))
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestInventoryStacking(t *testing.T) {
	r := reg(t)
	var inv Inventory
	scroll, fb := r.Must("scroll_empty"), r.Must("spell_fireball")

	inv.Add(scroll, 3)
	inv.Add(scroll, 2)
	inv.Add(fb, 2)
	if len(inv.Stacks()) != 3 {
		t.Fatalf("expected 1 scroll stack + 2 fireball slots, got %d", len(inv.Stacks()))
	}
	if inv.Count(scroll) != 5 || inv.Count(fb) != 2 {
		t.Fatalf("counts wrong: %d %d", inv.Count(scroll), inv.Count(fb))
	}

	if inv.Remove(scroll, 6) {
		t.Fatal("removing more than held must fail")
	}
	if !inv.Remove(scroll, 5) || inv.Has(scroll) {
		t.Fatal("removing all scrolls should empty the stack")
	}
	if !inv.Remove(fb, 1) || inv.Count(fb) != 1 || len(inv.Stacks()) != 1 {
		t.Fatalf("removing one unstackable should leave one slot, got %d", len(inv.Stacks()))
	}
	if ids := inv.IDs(); len(ids) != 1 || ids[0] != "spell_fireball" {
		t.Fatalf("ids %v", ids)
	}
}
