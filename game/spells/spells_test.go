package spells

import (
	"os"
	"testing"
)

func table(t *testing.T) *Table {
	data, err := os.ReadFile("../../assets/data/spells.json")
	if err != nil {
		t.Fatal(err)
	}
	tb, err := ParseTable(data)
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

func TestStarterSpellsResolve(t *testing.T) {
	tb := table(t)
	cases := []struct {
		ids   []string
		name  string
		shape Shape
	}{
		{[]string{"fire", "ball"}, "Fireball", ShapeProjectile},
		{[]string{"life", "mend"}, "Small heal", ShapeSelf},
		{[]string{"void", "step"}, "Minor teleport", ShapeBlink},
		{[]string{"ward", "shell"}, "Barrier", ShapeShield},
	}
	for _, c := range cases {
		sp := tb.Resolve(c.ids)
		if sp.Name != c.name || sp.Shape != c.shape || sp.Mana <= 0 {
			t.Errorf("%v -> %q %q mana %v", c.ids, sp.Name, sp.Shape, sp.Mana)
		}
	}
}

func TestEmergentCombination(t *testing.T) {
	tb := table(t)
	sp := tb.Resolve([]string{"ball", "moon", "fire"})
	if sp.Name != "Fireball" {
		t.Fatalf("moon fireball should still match the Fireball synergy, got %q", sp.Name)
	}
	plain := tb.Resolve([]string{"fire", "ball"})
	if sp.Mana >= plain.Mana {
		t.Fatalf("moon should make it cheaper: %v vs %v", sp.Mana, plain.Mana)
	}
	// No synergy: name assembles from labels in order.
	sp = tb.Resolve([]string{"ball", "fire", "life"})
	if sp.Synergy == "" || sp.Name != "Fireball" {
		// fire+ball is present, so the Fireball synergy still names it.
		t.Fatalf("got %q synergy %q", sp.Name, sp.Synergy)
	}
	sp = tb.Resolve([]string{"ball", "life"})
	if sp.Name != "Life Ball" {
		t.Fatalf("default name should be ordered labels, got %q", sp.Name)
	}
}

func TestMostSpecificSynergyWins(t *testing.T) {
	tb := table(t)
	sp := tb.Resolve([]string{"fire", "ball", "fire"})
	if sp.Name != "Greater fireball" {
		t.Fatalf("two fires should be Greater fireball, got %q", sp.Name)
	}
	plain := tb.Resolve([]string{"fire", "ball"})
	if sp.Damage <= plain.Damage*1.5 {
		t.Fatalf("damage should stack and multiply: %v vs %v", sp.Damage, plain.Damage)
	}
}

func TestUnknownKeywordShowsInName(t *testing.T) {
	tb := table(t)
	sp := tb.Resolve([]string{"fire", "glorp"})
	if sp.Name != "Fire glorp" {
		t.Fatalf("got %q", sp.Name)
	}
}

func TestParseRejectsUnknownSynergyKeyword(t *testing.T) {
	_, err := ParseTable([]byte(`{"keywords":[{"id":"a"}],"synergies":[{"keywords":["a","zzz"],"name":"x"}]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
