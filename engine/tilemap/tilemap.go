// Package tilemap is a grid of fixed-size tiles with a solidity flag per tile
// kind, drawn from a tileset image. Deliberately small: maps will later come
// from an editor export, but the in-memory shape stays this.
package tilemap

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tileset slices one image into TileSize squares, numbered left to right,
// top to bottom. Tile 0 is "empty" and is never drawn.
type Tileset struct {
	Image    *ebiten.Image
	TileSize int
	Columns  int
	// Solid marks which tile ids block movement.
	Solid map[int]bool
}

// NewTileset wraps an image whose width is a multiple of tileSize.
func NewTileset(img *ebiten.Image, tileSize int, solid ...int) *Tileset {
	ts := &Tileset{
		Image:    img,
		TileSize: tileSize,
		Columns:  img.Bounds().Dx() / tileSize,
		Solid:    map[int]bool{},
	}
	for _, id := range solid {
		ts.Solid[id] = true
	}
	return ts
}

// Sub returns the sub-image for a tile id (1-based; 0 is empty).
func (ts *Tileset) Sub(id int) *ebiten.Image {
	i := id - 1
	x := (i % ts.Columns) * ts.TileSize
	y := (i / ts.Columns) * ts.TileSize
	return ts.Image.SubImage(image.Rect(x, y, x+ts.TileSize, y+ts.TileSize)).(*ebiten.Image)
}

// Map is a set of layers over one tileset. Layers draw bottom to top.
type Map struct {
	Tileset       *Tileset
	Width, Height int
	Layers        [][]int // each layer is Width*Height ids, row-major
}

// NewMap makes an empty map with the given number of layers.
func NewMap(ts *Tileset, w, h, layers int) *Map {
	m := &Map{Tileset: ts, Width: w, Height: h}
	for i := 0; i < layers; i++ {
		m.Layers = append(m.Layers, make([]int, w*h))
	}
	return m
}

// PixelSize is the map size in world pixels.
func (m *Map) PixelSize() (w, h int) {
	return m.Width * m.Tileset.TileSize, m.Height * m.Tileset.TileSize
}

// Get returns the tile id at a cell in a layer, or 0 outside the map.
func (m *Map) Get(layer, x, y int) int {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return 0
	}
	return m.Layers[layer][y*m.Width+x]
}

// Set writes a tile id; writes outside the map are ignored.
func (m *Map) Set(layer, x, y, id int) {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return
	}
	m.Layers[layer][y*m.Width+x] = id
}

// Fill sets every cell of a layer to id.
func (m *Map) Fill(layer, id int) {
	l := m.Layers[layer]
	for i := range l {
		l[i] = id
	}
}

// SolidAt reports whether any layer has a solid tile at the cell. Cells
// outside the map count as solid so entities cannot walk off the edge.
func (m *Map) SolidAt(x, y int) bool {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return true
	}
	for l := range m.Layers {
		if m.Tileset.Solid[m.Get(l, x, y)] {
			return true
		}
	}
	return false
}

// Collides reports whether a world-pixel rectangle overlaps any solid cell.
func (m *Map) Collides(x, y, w, h float64) bool {
	ts := float64(m.Tileset.TileSize)
	x0 := int(math.Floor(x / ts))
	y0 := int(math.Floor(y / ts))
	// Right/bottom edges are exclusive: a box ending exactly on a cell
	// boundary does not touch the next cell.
	x1 := int(math.Ceil((x+w)/ts)) - 1
	y1 := int(math.Ceil((y+h)/ts)) - 1
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			if m.SolidAt(cx, cy) {
				return true
			}
		}
	}
	return false
}

// Draw renders the tiles visible inside view (a world-pixel rectangle),
// translated by (ox, oy), which is normally the camera offset.
func (m *Map) Draw(dst *ebiten.Image, ox, oy float64, viewX, viewY, viewW, viewH float64) {
	ts := m.Tileset.TileSize
	tsf := float64(ts)
	x0 := clamp(int(math.Floor(viewX/tsf)), 0, m.Width-1)
	y0 := clamp(int(math.Floor(viewY/tsf)), 0, m.Height-1)
	x1 := clamp(int(math.Ceil((viewX+viewW)/tsf)), 0, m.Width-1)
	y1 := clamp(int(math.Ceil((viewY+viewH)/tsf)), 0, m.Height-1)

	op := &ebiten.DrawImageOptions{}
	for l := range m.Layers {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				id := m.Get(l, x, y)
				if id == 0 {
					continue
				}
				op.GeoM.Reset()
				op.GeoM.Translate(float64(x*ts)+ox, float64(y*ts)+oy)
				dst.DrawImage(m.Tileset.Sub(id), op)
			}
		}
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
