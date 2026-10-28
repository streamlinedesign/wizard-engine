package tilemap

import "testing"

// newTestMap builds a 4x4 single-layer map with no image (collision only).
func newTestMap() *Map {
	ts := &Tileset{TileSize: 16, Columns: 4, Solid: map[int]bool{2: true}}
	m := NewMap(ts, 4, 4, 1)
	m.Fill(0, 1)
	m.Set(0, 2, 2, 2) // solid block at cell (2,2) => pixels 32..48
	return m
}

func TestSolidAtEdges(t *testing.T) {
	m := newTestMap()
	if !m.SolidAt(-1, 0) || !m.SolidAt(0, 4) {
		t.Fatal("outside the map should be solid")
	}
	if m.SolidAt(0, 0) || !m.SolidAt(2, 2) {
		t.Fatal("solid lookup wrong")
	}
}

func TestCollidesIsEdgeExclusive(t *testing.T) {
	m := newTestMap()
	if m.Collides(16, 16, 16, 16) {
		t.Fatal("box ending exactly at x=32 must not touch cell 2")
	}
	if !m.Collides(16, 16, 17, 17) {
		t.Fatal("box crossing into x=33 must collide")
	}
	if !m.Collides(40, 40, 2, 2) {
		t.Fatal("box fully inside solid cell must collide")
	}
	if m.Collides(48, 48, 8, 8) {
		t.Fatal("box starting at the far edge of the solid cell must not collide")
	}
}

func TestGetSetOutOfRangeIsSafe(t *testing.T) {
	m := newTestMap()
	m.Set(0, 99, 99, 5)
	if m.Get(0, 99, 99) != 0 {
		t.Fatal("out-of-range Get should be 0")
	}
	if w, h := m.PixelSize(); w != 64 || h != 64 {
		t.Fatalf("pixel size %dx%d", w, h)
	}
}
