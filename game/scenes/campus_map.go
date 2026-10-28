package scenes

import (
	"math/rand"

	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
)

// buildCampus lays out the hub in code: a 64x48 tile campus with a gate at
// the bottom, a central quad with a fountain, the old wizard's study west
// of the quad, scriptorium and library facades to the north, and the harbour
// to the east. Each building is a facade for now; doors open in later steps.
func buildCampus(ts *tilemap.Tileset) *tilemap.Map {
	const w, h = 64, 48
	rng := rand.New(rand.NewSource(3))
	m := tilemap.NewMap(ts, w, h, 3)

	// Ground: grass with variation.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			id := cGrass
			switch rng.Intn(12) {
			case 0:
				id = cGrassWorn
			case 1:
				id = cGrassFlowers
			}
			m.Set(layerGround, x, y, id)
		}
	}

	// Harbour water down the east side, with a shore row.
	for y := 0; y < h; y++ {
		for x := 58; x < w; x++ {
			m.Set(layerGround, x, y, cWater)
		}
		m.Set(layerGround, 57, y, cWater)
	}
	for y := 0; y < h; y++ {
		m.Set(layerGround, 56, y, cShore)
	}
	// Dock planks out into the water.
	for x := 56; x < 61; x++ {
		m.Set(layerGround, x, 31, cWoodFloor)
		m.Set(layerGround, x, 32, cWoodFloor)
	}

	// Hedge border on the other three sides.
	for x := 0; x < 57; x++ {
		m.Set(layerProps, x, 0, cHedge)
		m.Set(layerProps, x, h-1, cHedge)
	}
	for y := 0; y < h; y++ {
		m.Set(layerProps, 0, y, cHedge)
	}
	// Gate: gap in the bottom hedge.
	for x := 29; x <= 32; x++ {
		m.Set(layerProps, x, h-1, 0)
	}

	// Quad: a plaza in the middle with a fountain.
	fillRect(m, layerGround, 24, 20, 40, 30, cPlaza)
	m.Set(layerProps, 32, 25, cFountain)
	for _, p := range [][2]int{{26, 22}, {38, 22}, {26, 28}, {38, 28}} {
		m.Set(layerProps, p[0], p[1], cLamp)
	}
	for _, p := range [][2]int{{29, 29}, {35, 29}} {
		m.Set(layerProps, p[0], p[1], cBench)
	}

	// Paths: gate to quad, quad north to the scriptorium and library, quad
	// east to the harbour, quad west to the study.
	fillRect(m, layerGround, 30, 30, 32, h, cPath)
	fillRect(m, layerGround, 30, 12, 32, 20, cPath)
	fillRect(m, layerGround, 16, 12, 46, 14, cPath)
	fillRect(m, layerGround, 40, 24, 56, 26, cPath)
	fillRect(m, layerGround, 14, 24, 24, 26, cPath)

	// Buildings (facades).
	building(m, 16, 4, 10, 7) // scriptorium
	building(m, 38, 4, 10, 7) // library
	building(m, 6, 17, 8, 6)  // old wizard's study
	building(m, 46, 36, 8, 5) // harbourmaster
	building(m, 8, 36, 10, 6) // shops

	// Signs by the doors and at the gate.
	m.Set(layerProps, 20, 12, cSignpost)
	m.Set(layerProps, 42, 12, cSignpost)
	m.Set(layerProps, 28, 42, cSignpost)
	m.Set(layerProps, 54, 30, cSignpost)

	// Trees wherever there is plain grass and nothing nearby.
	for i := 0; i < 90; i++ {
		x, y := 2+rng.Intn(53), 2+rng.Intn(h-5)
		if !clearFor(m, x, y) {
			continue
		}
		m.Set(layerOverhead, x, y, cTreeTop)
		m.Set(layerProps, x, y+1, cTreeBottom)
	}
	// Keep the spawn and the old wizard's spot clear.
	for y := 40; y < 47; y++ {
		for x := 28; x <= 34; x++ {
			m.Set(layerProps, x, y, 0)
			m.Set(layerOverhead, x, y, 0)
		}
	}
	return m
}

func fillRect(m *tilemap.Map, layer, x0, y0, x1, y1, id int) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			m.Set(layer, x, y, id)
		}
	}
}

// building draws a facade: roof rows on the overhead layer so the player
// can walk "behind" the top of it, then a wall-top band, walls with a door
// in the middle and windows either side, all on the props layer (solid).
func building(m *tilemap.Map, x, y, w, h int) {
	roofRows := h - 3
	for ry := 0; ry < roofRows; ry++ {
		id := cRoof
		if ry == 0 {
			id = cRoofEdge
		}
		for bx := x; bx < x+w; bx++ {
			m.Set(layerOverhead, bx, y+ry, id)
			m.Set(layerProps, bx, y+ry, cWall) // solid under the roof
		}
	}
	for bx := x; bx < x+w; bx++ {
		m.Set(layerProps, bx, y+roofRows, cWallTop)
		m.Set(layerProps, bx, y+roofRows+1, cWall)
		m.Set(layerProps, bx, y+roofRows+2, cWall)
	}
	door := x + w/2
	m.Set(layerProps, door, y+h-1, cDoor)
	m.Set(layerProps, x+1, y+h-2, cWindow)
	m.Set(layerProps, x+w-2, y+h-2, cWindow)
	// Doorstep.
	m.Set(layerGround, door, y+h, cPath)
}

// clearFor reports whether a 3x3 around (x, y) and the cell below are plain
// grass with no props, so a tree fits without crowding anything.
func clearFor(m *tilemap.Map, x, y int) bool {
	for dy := -1; dy <= 2; dy++ {
		for dx := -1; dx <= 1; dx++ {
			g := m.Get(layerGround, x+dx, y+dy)
			if g != cGrass && g != cGrassWorn && g != cGrassFlowers {
				return false
			}
			if m.Get(layerProps, x+dx, y+dy) != 0 || m.Get(layerOverhead, x+dx, y+dy) != 0 {
				return false
			}
		}
	}
	return true
}
