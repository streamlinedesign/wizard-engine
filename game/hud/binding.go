package hud

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/items"
	"github.com/streamlinedesign/wizard-engine/game/spells"
)

// Ring is the number of outer slots on the binding circle.
const Ring = 12

// Binding is the ritual-circle screen: twelve slots around a centre slot.
// The centre takes an empty scroll; the ring takes anything with keywords.
// Click an inventory item to place it in the next free slot, click a slot
// to take it back, and confirm to bind. Everything placed is consumed.
type Binding struct {
	Open  bool
	Icons *tilemap.Tileset
	Table *spells.Table
	Theme ui.Theme

	Center *items.Def
	Ring   [Ring]*items.Def
	// OnBind receives the resolved spell when the player confirms.
	OnBind func(sp spells.Spell)

	message string
}

// NewBinding wires the screen to icons and the keyword table.
func NewBinding(icons *tilemap.Tileset, table *spells.Table) *Binding {
	return &Binding{Icons: icons, Table: table, Theme: ui.Parchment}
}

// Keywords lists the keywords of everything on the ring, in slot order.
func (b *Binding) Keywords() []string {
	var ids []string
	for _, d := range b.Ring {
		if d != nil {
			ids = append(ids, d.Keywords...)
		}
	}
	return ids
}

// Preview resolves what the circle would make right now.
func (b *Binding) Preview() (spells.Spell, bool) {
	ids := b.Keywords()
	if b.Center == nil || len(ids) == 0 {
		return spells.Spell{}, false
	}
	return b.Table.Resolve(ids), true
}

// Clear returns every placed item to the inventory.
func (b *Binding) Clear(inv *items.Inventory) {
	if b.Center != nil {
		inv.Add(b.Center, 1)
		b.Center = nil
	}
	for i, d := range b.Ring {
		if d != nil {
			inv.Add(d, 1)
			b.Ring[i] = nil
		}
	}
}

// place moves one of an item from the inventory onto the circle.
func (b *Binding) place(inv *items.Inventory, d *items.Def) {
	if d.Kind == items.KindScroll {
		if b.Center != nil {
			b.message = "The centre already holds a scroll."
			return
		}
		if inv.Remove(d, 1) {
			b.Center = d
		}
		return
	}
	if len(d.Keywords) == 0 {
		b.message = d.Name + " has no keywords to bind."
		return
	}
	for i := range b.Ring {
		if b.Ring[i] == nil {
			if inv.Remove(d, 1) {
				b.Ring[i] = d
				b.message = ""
			}
			return
		}
	}
	b.message = "The circle is full."
}

// layout numbers used by both Update and Draw.
type bindLayout struct {
	cx, cy, radius float64
	slot           float64 // slot size
	listX, listY   float64
	listW, rowH    float64
	confirmX       float64
	confirmY       float64
	confirmW       float64
	confirmH       float64
}

func (b *Binding) layout(canvas *ebiten.Image) bindLayout {
	bd := canvas.Bounds()
	w, h := float64(bd.Dx()), float64(bd.Dy())
	l := bindLayout{slot: 18, listW: 150, rowH: 16}
	l.radius = math.Min(h, w*0.55) * 0.32
	l.cx = w*0.36 + 10
	l.cy = h / 2
	l.listX = w - l.listW - 10
	l.listY = 24
	l.confirmW, l.confirmH = 90, 18
	l.confirmX = l.cx - l.confirmW/2
	l.confirmY = h - 28
	return l
}

func (l bindLayout) slotPos(i int) (x, y float64) {
	a := -math.Pi/2 + float64(i)*2*math.Pi/Ring
	return math.Floor(l.cx + math.Cos(a)*l.radius), math.Floor(l.cy + math.Sin(a)*l.radius)
}

// Update handles clicks. It returns true while the screen is open.
func (b *Binding) Update(ctx *engine.Context, canvas *ebiten.Image, inv *items.Inventory) bool {
	if !b.Open {
		return false
	}
	if ctx.Input.JustPressed(input.Menu) {
		b.Clear(inv)
		b.Open = false
		return true
	}
	if ctx.Input.JustPressed(input.Interact) {
		b.confirm(inv)
		return true
	}
	if !ctx.Input.Clicked() {
		return true
	}
	mx, my := ctx.Viewport.Cursor()
	fx, fy := float64(mx), float64(my)
	l := b.layout(canvas)

	// Confirm button.
	if fx >= l.confirmX && fx < l.confirmX+l.confirmW && fy >= l.confirmY && fy < l.confirmY+l.confirmH {
		b.confirm(inv)
		return true
	}
	// Slots: clicking returns the item.
	hit := func(sx, sy float64) bool {
		return math.Abs(fx-sx) <= l.slot/2 && math.Abs(fy-sy) <= l.slot/2
	}
	if hit(l.cx, l.cy) && b.Center != nil {
		inv.Add(b.Center, 1)
		b.Center = nil
		return true
	}
	for i := range b.Ring {
		sx, sy := l.slotPos(i)
		if hit(sx, sy) && b.Ring[i] != nil {
			inv.Add(b.Ring[i], 1)
			b.Ring[i] = nil
			return true
		}
	}
	// Inventory list.
	if fx >= l.listX && fx < l.listX+l.listW {
		row := int((fy - l.listY - 14) / l.rowH)
		stacks := inv.Stacks()
		if row >= 0 && row < len(stacks) {
			b.place(inv, stacks[row].Def)
		}
	}
	return true
}

func (b *Binding) confirm(inv *items.Inventory) {
	sp, ok := b.Preview()
	if !ok {
		b.message = "Needs an empty scroll in the centre and at least one item on the ring."
		return
	}
	b.Center = nil
	for i := range b.Ring {
		b.Ring[i] = nil
	}
	if b.OnBind != nil {
		b.OnBind(sp)
	}
	b.message = "Bound: " + sp.Name
}

// Draw renders the circle screen over the world.
func (b *Binding) Draw(canvas *ebiten.Image, inv *items.Inventory) {
	if !b.Open {
		return
	}
	bd := canvas.Bounds()
	vector.DrawFilledRect(canvas, 0, 0, float32(bd.Dx()), float32(bd.Dy()), color.RGBA{10, 8, 20, 200}, false)
	l := b.layout(canvas)

	// The 12-point star: every point joined to the one five steps on.
	ink := color.RGBA{230, 200, 120, 255}
	dim := color.RGBA{230, 200, 120, 90}
	vector.StrokeCircle(canvas, float32(l.cx), float32(l.cy), float32(l.radius), 1, ink, false)
	vector.StrokeCircle(canvas, float32(l.cx), float32(l.cy), float32(l.radius*0.35), 1, dim, false)
	for i := 0; i < Ring; i++ {
		x0, y0 := l.slotPos(i)
		x1, y1 := l.slotPos((i + 5) % Ring)
		vector.StrokeLine(canvas, float32(x0), float32(y0), float32(x1), float32(y1), 1, dim, false)
	}

	drawSlot := func(x, y float64, d *items.Def, label string) {
		th := b.Theme
		if d == nil {
			th.PanelFill = color.RGBA{40, 34, 50, 255}
		}
		ui.Panel(canvas, x-l.slot/2, y-l.slot/2, l.slot, l.slot, th)
		if d != nil && b.Icons != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x-8, y-8)
			canvas.DrawImage(b.Icons.Sub(d.Icon), op)
		} else if label != "" {
			ui.Print(canvas, label, x-ui.Width(label)/2, y-6, ui.Parchment.TextDim)
		}
	}
	drawSlot(l.cx, l.cy, b.Center, "scroll")
	for i := range b.Ring {
		x, y := l.slotPos(i)
		drawSlot(x, y, b.Ring[i], "")
	}

	// Preview.
	ui.Print(canvas, "Ritual circle", 8, 6, ink)
	if sp, ok := b.Preview(); ok {
		kw := ""
		for i, k := range sp.Keywords {
			if i > 0 {
				kw += " + "
			}
			kw += k
		}
		ui.Print(canvas, sp.Name, 8, 20, ui.Parchment.PanelFill)
		ui.Print(canvas, kw, 8, 32, ui.Parchment.TextDim)
		ui.Print(canvas, fmt.Sprintf("%s, %.0f mana", describe(sp), sp.Mana), 8, 44, ui.Parchment.TextDim)
	} else {
		ui.Print(canvas, "Scroll in the centre, items on the ring.", 8, 20, ui.Parchment.TextDim)
	}
	if b.message != "" {
		ui.Print(canvas, b.message, 8, float64(bd.Dy())-14, ink)
	}

	// Confirm button.
	ui.Panel(canvas, l.confirmX, l.confirmY, l.confirmW, l.confirmH, b.Theme)
	ui.Print(canvas, "Bind (E)", l.confirmX+6, l.confirmY+2, b.Theme.Accent)

	// Inventory list.
	ui.Panel(canvas, l.listX, l.listY-6, l.listW, float64(bd.Dy())-l.listY-10, b.Theme)
	ui.Print(canvas, "Inventory (click)", l.listX+4, l.listY-4, b.Theme.Accent)
	ry := l.listY + 14
	for _, s := range inv.Stacks() {
		if ry > float64(bd.Dy())-30 {
			break
		}
		if b.Icons != nil && s.Def.Icon > 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(l.listX+4, ry-1)
			canvas.DrawImage(b.Icons.Sub(s.Def.Icon), op)
		}
		label := s.Def.Name
		if s.Def.Stackable {
			label = fmt.Sprintf("%s x%d", label, s.Count)
		}
		ui.Print(canvas, label, l.listX+24, ry, b.Theme.Text)
		ry += l.rowH
	}
}

func describe(sp spells.Spell) string {
	switch sp.Shape {
	case spells.ShapeProjectile:
		return fmt.Sprintf("%d shot, %.0f damage", max(sp.Count, 1), sp.Damage)
	case spells.ShapeSelf:
		return fmt.Sprintf("heals %.0f", sp.Heal)
	case spells.ShapeBlink:
		return fmt.Sprintf("blink %.0f px", sp.Range)
	case spells.ShapeShield:
		return fmt.Sprintf("shield %.0f for %.0fs", sp.Shield, sp.Duration)
	}
	return "no shape yet"
}
