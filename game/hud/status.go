package hud

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/grimoire"
)

// Vitals is what the status bar needs to know about the player.
type Vitals struct {
	HP, MaxHP     float64
	Mana, MaxMana float64
	// BuffName and BuffLeft show an active buff; empty when none.
	BuffName string
	BuffLeft float64
	// Casting shows a cast bar: progress 0..1 and the spell name.
	Casting  bool
	CastFrac float64
	CastName string
}

// Status draws health and mana bars, the buff line, the four quick slots
// with their cooldowns, and the cast bar.
func Status(canvas *ebiten.Image, v Vitals, book *grimoire.Book, cooldownFrac func(slot int) float64) {
	const x, y, w, h = 6, 6, 80, 6
	bar(canvas, x, y, w, h, v.HP/v.MaxHP, color.RGBA{200, 60, 60, 255})
	ui.Print(canvas, fmt.Sprintf("%.0f/%.0f", v.HP, v.MaxHP), x+w+4, y-4, ui.Parchment.PanelFill)
	bar(canvas, x, y+h+3, w, h, v.Mana/v.MaxMana, color.RGBA{80, 120, 220, 255})
	ui.Print(canvas, fmt.Sprintf("%.0f/%.0f", v.Mana, v.MaxMana), x+w+4, y+h-1, ui.Parchment.PanelFill)
	if v.BuffName != "" && v.BuffLeft > 0 {
		ui.Print(canvas, fmt.Sprintf("%s %.1fs", v.BuffName, v.BuffLeft), x, y+2*h+6, color.RGBA{200, 230, 240, 255})
	}

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
		if cooldownFrac != nil {
			if f := cooldownFrac(i); f > 0 {
				vector.DrawFilledRect(canvas, float32(sx+1), float32(sy+1), float32((sw-2)*f), float32(sh-2), color.RGBA{90, 70, 60, 150}, false)
			}
		}
		label := fmt.Sprintf("%d", i+1)
		if sp := book.Slot(i); sp != nil {
			label = fmt.Sprintf("%d %s", i+1, sp.Name)
		}
		ui.Print(canvas, label, sx+3, sy+2, ui.Parchment.Text)
		sx += sw + 3
	}

	if v.Casting {
		cw := 80.0
		cx := float64(b.Dx())/2 - cw/2
		cy := float64(b.Dy()) - sh - 24
		bar(canvas, cx, cy, cw, 5, v.CastFrac, color.RGBA{230, 200, 120, 255})
		ui.Print(canvas, v.CastName, cx, cy-14, ui.Parchment.PanelFill)
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
