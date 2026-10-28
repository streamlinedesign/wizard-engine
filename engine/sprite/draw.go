package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite pairs a parsed Sheet with its pixels so it can be drawn.
type Sprite struct {
	Sheet *Sheet
	Image *ebiten.Image
}

// Draw blits one frame so that the top-left of its untrimmed cell lands at
// (x, y). With flipX the cell is mirrored in place, which is how a single
// "side" animation serves both left and right.
func (s *Sprite) Draw(dst *ebiten.Image, fr Frame, x, y float64, flipX bool) {
	sub := s.Image.SubImage(fr.Rect).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	w := float64(fr.Rect.Dx())
	offX := float64(fr.Offset.X)
	if flipX {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(w, 0)
		// Mirror the trim offset too, so the frame stays inside its cell.
		offX = float64(fr.Size.X) - offX - w
	}
	op.GeoM.Translate(x+offX, y+float64(fr.Offset.Y))
	dst.DrawImage(sub, op)
}
