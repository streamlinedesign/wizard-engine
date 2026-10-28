// Package render holds the pixel-perfect presentation layer: a fixed-size
// virtual canvas that the game draws into, integer-scaled onto whatever
// window the player has, plus a world camera.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Viewport owns the virtual canvas. Everything in the game is drawn at the
// canvas resolution (e.g. 320x180) and then blown up by a whole-number factor
// so pixels stay crisp and square on any monitor. Leftover space is
// letterboxed with Border.
type Viewport struct {
	Width, Height int
	Border        color.Color

	canvas *ebiten.Image
	// winW, winH remember the last window size Present saw, so cursor
	// positions can be mapped back onto the canvas.
	winW, winH int
}

// NewViewport allocates a canvas of the given virtual size.
func NewViewport(width, height int) *Viewport {
	return &Viewport{
		Width:  width,
		Height: height,
		Border: color.Black,
		canvas: ebiten.NewImage(width, height),
	}
}

// Canvas is the image the game should draw into each frame.
func (v *Viewport) Canvas() *ebiten.Image { return v.canvas }

// Fit returns the integer scale factor and the pixel offset that centre the
// canvas inside a window of the given size. Scale is never below 1, so a
// window smaller than the canvas simply crops.
func Fit(canvasW, canvasH, windowW, windowH int) (scale, offX, offY int) {
	scale = windowW / canvasW
	if s := windowH / canvasH; s < scale {
		scale = s
	}
	if scale < 1 {
		scale = 1
	}
	offX = (windowW - canvasW*scale) / 2
	offY = (windowH - canvasH*scale) / 2
	return scale, offX, offY
}

// Present draws the canvas onto the real screen using Fit.
func (v *Viewport) Present(screen *ebiten.Image) {
	screen.Fill(v.Border)
	b := screen.Bounds()
	v.winW, v.winH = b.Dx(), b.Dy()
	scale, offX, offY := Fit(v.Width, v.Height, b.Dx(), b.Dy())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(offX), float64(offY))
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(v.canvas, op)
}

// WindowToCanvas converts a window-space point (such as the mouse cursor)
// into canvas pixels. Points in the letterbox map outside the canvas.
func (v *Viewport) WindowToCanvas(windowW, windowH, x, y int) (cx, cy int) {
	scale, offX, offY := Fit(v.Width, v.Height, windowW, windowH)
	return floorDiv(x-offX, scale), floorDiv(y-offY, scale)
}

// Cursor returns the mouse position in canvas pixels, using the window
// size from the last Present. Before the first frame it assumes 1:1.
func (v *Viewport) Cursor() (cx, cy int) {
	x, y := ebiten.CursorPosition()
	if v.winW == 0 || v.winH == 0 {
		return x, y
	}
	return v.WindowToCanvas(v.winW, v.winH, x, y)
}

// floorDiv divides rounding toward negative infinity, so points left of
// or above the canvas stay negative instead of collapsing onto column 0.
func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}
