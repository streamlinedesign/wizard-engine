package input

import (
	"math"
	"testing"
)

func TestNormalize(t *testing.T) {
	x, y := Normalize(0, 0)
	if x != 0 || y != 0 {
		t.Fatalf("zero vector should stay zero, got (%v,%v)", x, y)
	}
	x, y = Normalize(1, 1)
	if l := math.Hypot(x, y); math.Abs(l-1) > 1e-9 {
		t.Fatalf("diagonal should have length 1, got %v", l)
	}
	x, y = Normalize(-3, 0)
	if x != -1 || y != 0 {
		t.Fatalf("straight should normalise to (-1,0), got (%v,%v)", x, y)
	}
}

func TestDefaultBindingsCoverEveryAction(t *testing.T) {
	b := DefaultBindings()
	for _, a := range []Action{MoveUp, MoveDown, MoveLeft, MoveRight, Interact, Cast, Menu, Inventory} {
		if len(b[a]) == 0 {
			t.Errorf("action %q has no default key", a)
		}
	}
}
