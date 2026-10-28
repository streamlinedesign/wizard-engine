package tilemap

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawLayerRows renders one layer, but only the tile rows from fromRow to
// toRow inclusive, culled to the view. Scenes use it to interleave
// characters with props: rows above a character's feet first, then the
// character, then the rows below, so he walks behind trees and in front of
// fences.
func (m *Map) DrawLayerRows(dst *ebiten.Image, layer int, ox, oy float64, viewX, viewY, viewW, viewH float64, fromRow, toRow int) {
	ts := m.Tileset.TileSize
	tsf := float64(ts)
	x0 := clamp(int(math.Floor(viewX/tsf)), 0, m.Width-1)
	y0 := clamp(int(math.Floor(viewY/tsf)), 0, m.Height-1)
	x1 := clamp(int(math.Ceil((viewX+viewW)/tsf)), 0, m.Width-1)
	y1 := clamp(int(math.Ceil((viewY+viewH)/tsf)), 0, m.Height-1)
	if fromRow > y0 {
		y0 = fromRow
	}
	if toRow < y1 {
		y1 = toRow
	}

	op := &ebiten.DrawImageOptions{}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			id := m.Get(layer, x, y)
			if id == 0 {
				continue
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x*ts)+ox, float64(y*ts)+oy)
			dst.DrawImage(m.Tileset.Sub(id), op)
		}
	}
}

// RowOf converts a world y coordinate to a tile row.
func (m *Map) RowOf(y float64) int {
	return int(math.Floor(y / float64(m.Tileset.TileSize)))
}
