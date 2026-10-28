// Package hud draws the in-game overlays: for now the inventory panel and
// the interaction prompt.
package hud

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/items"
)

// InventoryPanel lists the inventory with icons and shows the selected
// item's keywords and description.
type InventoryPanel struct {
	Open  bool
	Icons *tilemap.Tileset
	Theme ui.Theme

	sel int
}

// NewInventoryPanel wires the panel to the icon sheet.
func NewInventoryPanel(icons *tilemap.Tileset) *InventoryPanel {
	return &InventoryPanel{Icons: icons, Theme: ui.Parchment}
}

// Update toggles the panel and moves the selection. It reports whether the
// panel consumed input this tick (so the player does not walk while it is
// open).
func (p *InventoryPanel) Update(ctx *engine.Context, inv *items.Inventory) bool {
	if ctx.Input.JustPressed(input.Inventory) {
		p.Open = !p.Open
	}
	if !p.Open {
		return false
	}
	if ctx.Input.JustPressed(input.Menu) {
		p.Open = false
		return true
	}
	n := len(inv.Stacks())
	if ctx.Input.JustPressed(input.MoveDown) && p.sel < n-1 {
		p.sel++
	}
	if ctx.Input.JustPressed(input.MoveUp) && p.sel > 0 {
		p.sel--
	}
	if p.sel >= n {
		p.sel = max(n-1, 0)
	}
	return true
}

// Draw renders the panel over the canvas.
func (p *InventoryPanel) Draw(canvas *ebiten.Image, inv *items.Inventory) {
	if !p.Open {
		return
	}
	b := canvas.Bounds()
	const w, h = 300, 150
	x := float64(b.Dx()-w) / 2
	y := float64(b.Dy()-h) / 2
	ui.Panel(canvas, x, y, w, h, p.Theme)
	ui.Print(canvas, "Inventory", x+6, y+3, p.Theme.Accent)

	const listW = 150
	stacks := inv.Stacks()
	if len(stacks) == 0 {
		ui.Print(canvas, "Nothing yet.", x+6, y+20, p.Theme.TextDim)
		return
	}
	const rowH = 16
	// Scroll so the selection stays visible.
	visible := (h - 24) / rowH
	first := 0
	if p.sel >= visible {
		first = p.sel - visible + 1
	}
	ry := y + 18
	for i := first; i < len(stacks) && i < first+visible; i++ {
		s := stacks[i]
		if i == p.sel {
			ui.Panel(canvas, x+4, ry-1, listW-8, rowH, ui.Theme{PanelFill: p.Theme.PanelBorder, PanelBorder: p.Theme.PanelBorder})
		}
		if p.Icons != nil && s.Def.Icon > 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x+6, ry-1)
			canvas.DrawImage(p.Icons.Sub(s.Def.Icon), op)
		}
		label := s.Def.Name
		if s.Def.Stackable {
			label = fmt.Sprintf("%s x%d", label, s.Count)
		}
		clr := p.Theme.Text
		if i == p.sel {
			clr = p.Theme.PanelFill
		}
		ui.Print(canvas, label, x+26, ry, clr)
		ry += rowH
	}

	// Detail pane.
	d := stacks[p.sel].Def
	dx := x + listW + 4
	dw := w - listW - 10
	ui.Print(canvas, d.Name, dx, y+18, p.Theme.Accent)
	cy := y + 18 + ui.LineHeight
	if len(d.Keywords) > 0 {
		kw := ""
		for i, k := range d.Keywords {
			if i > 0 {
				kw += " + "
			}
			kw += k
		}
		ui.Print(canvas, kw, dx, cy, p.Theme.TextDim)
		cy += ui.LineHeight
	}
	for _, line := range ui.Wrap(d.Description, float64(dw)) {
		ui.Print(canvas, line, dx, cy, p.Theme.Text)
		cy += ui.LineHeight
	}
}

// Prompt draws a small "E: Talk" style hint above a world point.
func Prompt(canvas *ebiten.Image, label string, wx, wy float64) {
	w := ui.Width(label) + 8
	x := wx - w/2
	y := wy - 40
	ui.Panel(canvas, x, y, w, ui.LineHeight+4, ui.Parchment)
	ui.Print(canvas, label, x+4, y+1, ui.Parchment.Text)
}
