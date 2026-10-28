package npc

import "testing"

func TestInRangeAndOverlaps(t *testing.T) {
	n := &NPC{X: 100, Y: 100, BoxW: 10, BoxH: 6}
	if !n.InRange(110, 100, 12) || n.InRange(120, 100, 12) {
		t.Fatal("InRange wrong")
	}
	// Box spans x 95..105, y 94..100.
	if !n.Overlaps(104, 93, 4, 4) {
		t.Fatal("corner overlap missed")
	}
	if n.Overlaps(105, 100, 4, 4) {
		t.Fatal("edge-touching should not overlap")
	}
	sign := &NPC{X: 0, Y: 0}
	if sign.Overlaps(-1, -1, 2, 2) {
		t.Fatal("a sign without a box never blocks")
	}
}
