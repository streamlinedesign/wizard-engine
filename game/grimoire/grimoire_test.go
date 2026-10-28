package grimoire

import (
	"testing"

	"github.com/streamlinedesign/wizard-engine/game/spells"
)

func TestBindFillsSlotsInOrder(t *testing.T) {
	b := New("Uriel")
	for i := 0; i < 5; i++ {
		b.Bind(spells.Spell{Name: string(rune('a' + i))})
	}
	if b.Slots != [4]int{0, 1, 2, 3} {
		t.Fatalf("slots %v", b.Slots)
	}
	if b.Active().Name != "a" {
		t.Fatalf("active should be slot 0, got %q", b.Active().Name)
	}
	b.Assign(0, 4)
	if b.Active().Name != "e" {
		t.Fatalf("assign failed: %q", b.Active().Name)
	}
	b.Selected = 9
	if b.Active() != nil {
		t.Fatal("out-of-range selection should be nil")
	}
}
