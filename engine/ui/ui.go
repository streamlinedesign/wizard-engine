// Package ui draws the placeholder interface: panels, text and a typewriter
// text box. It is intentionally plain so real UI art can replace it piece by
// piece (the panel becomes a 9-slice, the font becomes a bitmap sheet).
package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/bitmapfont/v3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Theme is the colour set every widget uses.
type Theme struct {
	PanelFill   color.Color
	PanelBorder color.Color
	Text        color.Color
	TextDim     color.Color
	Accent      color.Color
}

// Parchment is the default look: dark ink on a warm panel.
var Parchment = Theme{
	PanelFill:   color.RGBA{236, 226, 200, 255},
	PanelBorder: color.RGBA{96, 56, 48, 255},
	Text:        color.RGBA{40, 30, 40, 255},
	TextDim:     color.RGBA{120, 104, 96, 255},
	Accent:      color.RGBA{170, 50, 60, 255},
}

// Face is the font every widget draws with. bitmapfont is 12px tall, which
// is large for a 320x180 canvas but readable; a 5x7 pixel font replaces it.
var Face text.Face = text.NewGoXFace(bitmapfont.Face)

// LineHeight is the vertical advance between lines of Face.
const LineHeight = 12

// Panel draws a bordered box.
func Panel(dst *ebiten.Image, x, y, w, h float64, th Theme) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), th.PanelBorder, false)
	vector.DrawFilledRect(dst, float32(x+1), float32(y+1), float32(w-2), float32(h-2), th.PanelFill, false)
}

// Print draws text with its top-left at (x, y).
func Print(dst *ebiten.Image, s string, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	op.LineSpacing = LineHeight
	text.Draw(dst, s, Face, op)
}

// Width measures a single line of text in pixels.
func Width(s string) float64 {
	w, _ := text.Measure(s, Face, LineHeight)
	return w
}

// Wrap breaks text into lines no wider than maxWidth pixels, on spaces.
func Wrap(s string, maxWidth float64) []string {
	var lines []string
	for _, para := range strings.Split(s, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if Width(line+" "+w) <= maxWidth {
				line += " " + w
				continue
			}
			lines = append(lines, line)
			line = w
		}
		lines = append(lines, line)
	}
	return lines
}
