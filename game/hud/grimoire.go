package hud

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/grimoire"
)

// GrimoirePanel is the spellbook: every bound spell on the left page, the
// selected spell's ten stats on the right, and 1 to 4 to put it on the bar.
type GrimoirePanel struct {
	Open  bool
	Theme ui.Theme

	sel int
}

// NewGrimoirePanel makes a closed book.
func NewGrimoirePanel() *GrimoirePanel {
	return &GrimoirePanel{Theme: ui.Parchment}
}

// Update toggles the book and handles selection and slot assignment. It
// reports whether the book consumed input this tick.
func (p *GrimoirePanel) Update(ctx *engine.Context, book *grimoire.Book) bool {
	if ctx.Input.JustPressed(input.Grimoire) {
		p.Open = !p.Open
	}
	if !p.Open {
		return false
	}
	if ctx.Input.JustPressed(input.Menu) {
		p.Open = false
		return true
	}
	n := len(book.Spells)
	if ctx.Input.JustPressed(input.MoveDown) && p.sel < n-1 {
		p.sel++
	}
	if ctx.Input.JustPressed(input.MoveUp) && p.sel > 0 {
		p.sel--
	}
	if p.sel >= n {
		p.sel = max(n-1, 0)
	}
	for i, a := range []input.Action{input.Slot1, input.Slot2, input.Slot3, input.Slot4} {
		if ctx.Input.JustPressed(a) && n > 0 {
			book.Assign(i, p.sel)
		}
	}
	// Clicking a row selects it.
	if ctx.Input.Clicked() {
		mx, my := ctx.Viewport.Cursor()
		l := p.layout(ctx.Viewport.Canvas())
		row := int((float64(my) - l.listY) / l.rowH)
		if float64(mx) >= l.x && float64(mx) < l.x+l.listW && row >= 0 && row < n {
			p.sel = row
		}
	}
	return true
}

type bookLayout struct {
	x, y, w, h   float64
	listW, listY float64
	rowH         float64
}

func (p *GrimoirePanel) layout(canvas *ebiten.Image) bookLayout {
	b := canvas.Bounds()
	l := bookLayout{w: 420, h: 220, rowH: 16}
	l.x = float64(b.Dx())/2 - l.w/2
	l.y = float64(b.Dy())/2 - l.h/2
	l.listW = l.w / 2
	l.listY = l.y + 34
	return l
}

// Draw renders the open book.
func (p *GrimoirePanel) Draw(canvas *ebiten.Image, book *grimoire.Book) {
	if !p.Open {
		return
	}
	l := p.layout(canvas)
	ui.Panel(canvas, l.x, l.y, l.w, l.h, p.Theme)
	// Spine.
	ui.Panel(canvas, l.x+l.listW-1, l.y+4, 2, l.h-8, ui.Theme{PanelFill: p.Theme.PanelBorder, PanelBorder: p.Theme.PanelBorder})

	ui.Print(canvas, fmt.Sprintf("Grimoire of %s", book.Owner), l.x+8, l.y+6, p.Theme.Accent)
	ui.Print(canvas, fmt.Sprintf("%d spells bound", len(book.Spells)), l.x+8, l.y+18, p.Theme.TextDim)

	if len(book.Spells) == 0 {
		ui.Print(canvas, "No spells yet. Bind one in the ritual circle.", l.x+8, l.listY, p.Theme.TextDim)
		return
	}

	// Left page: the list, with slot numbers for spells on the bar.
	visible := int((l.h - 44) / l.rowH)
	first := 0
	if p.sel >= visible {
		first = p.sel - visible + 1
	}
	ry := l.listY
	for i := first; i < len(book.Spells) && i < first+visible; i++ {
		sp := book.Spells[i]
		clr := p.Theme.Text
		if i == p.sel {
			ui.Panel(canvas, l.x+4, ry-1, l.listW-10, l.rowH, ui.Theme{PanelFill: p.Theme.PanelBorder, PanelBorder: p.Theme.PanelBorder})
			clr = p.Theme.PanelFill
		}
		slot := "  "
		for s, idx := range book.Slots {
			if idx == i {
				slot = fmt.Sprintf("%d ", s+1)
			}
		}
		ui.Print(canvas, slot+sp.Name, l.x+8, ry, clr)
		ry += l.rowH
	}

	// Right page: the selected spell.
	sp := book.Spells[p.sel]
	rx := l.x + l.listW + 8
	ry = l.y + 6
	ui.Print(canvas, sp.Name, rx, ry, p.Theme.Accent)
	ry += ui.LineHeight
	ui.Print(canvas, strings.Join(sp.Keywords, " + "), rx, ry, p.Theme.TextDim)
	ry += ui.LineHeight
	ui.Print(canvas, sp.Describe(), rx, ry, p.Theme.Text)
	ry += ui.LineHeight + 4
	lines := sp.StatLines()
	// Two columns of five.
	for i, line := range lines {
		cx := rx + float64(i/5)*(l.listW/2-4)
		cy := ry + float64(i%5)*ui.LineHeight
		ui.Print(canvas, line, cx, cy, p.Theme.Text)
	}
	ui.Print(canvas, "1-4: put on the bar   Esc: close", rx, l.y+l.h-16, p.Theme.TextDim)
}
