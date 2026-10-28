package player

import "testing"

func TestFacingFrom(t *testing.T) {
	cases := []struct {
		dx, dy float64
		prev   Facing
		want   Facing
	}{
		{1, 0, Down, Right},
		{-1, 0, Down, Left},
		{0, -1, Down, Up},
		{0, 1, Up, Down},
		{1, 1, Down, Down},   // keep facing down while still heading down
		{1, -1, Down, Right}, // diagonal that reverses vertical goes horizontal
		{1, -1, Up, Up},
		{0, 0, Left, Left},
	}
	for _, c := range cases {
		if got := facingFrom(c.dx, c.dy, c.prev); got != c.want {
			t.Errorf("facingFrom(%v,%v,%v) = %v, want %v", c.dx, c.dy, c.prev, got, c.want)
		}
	}
}
