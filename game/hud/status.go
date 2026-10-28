package hud

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/grimoire"
)

// Status draws health and mana bars and the four quick slots.
func Status(canvas *ebiten.Image, hp, maxHP, mana, maxMana float64, book *grimoire.Book) {
	const x, y, w, h = 6, 6, 80, 6
	bar(canvas, x, y, w, h, hp/maxHP, color.RGBA{200, 60, 60, 255})
	ui.Print(canvas, fmt.Sprintf("%.0f/%.0f", hp, maxHP), x+w+4, y-4, ui.Parchment.PanelFill)
	bar(canvas, x, y+h+3, w, h, mana/maxMana, color.RGBA{80, 120, 220, 255})
	ui.Print(canvas, fmt.Sprintf("%.0f/%.0f", mana, maxMana), x+w+4, y+h-1, ui.Parchment.PanelFill)

	// Quick slots along the bottom-left.
	b := canvas.Bounds()
	const sw, sh = 92, 18
	sx := 6.0
	sy := float64(b.Dy()) - sh - 6
	for i := 0; i < 4; i++ {
		th := ui.Parchment
		if i == book.Selected {
			th.PanelBorder = ui.Parchment.Accent
		}
		ui.Panel(canvas, sx, sy, sw, sh, th)
		label := fmt.Sprintf("%d", i+1)
		if sp := book.Slot(i); sp != nil {
			label = fmt.Sprintf("%d %s", i+1, sp.Name)
		}
		ui.Print(canvas, label, sx+3, sy+2, ui.Parchment.Text)
		sx += sw + 3
	}
}

func bar(dst *ebiten.Image, x, y, w, h float64, frac float64, fill color.Color) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	vector.DrawFilledRect(dst, float32(x-1), float32(y-1), float32(w+2), float32(h+2), color.RGBA{40, 30, 40, 255}, false)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w*frac), float32(h), fill, false)
}
