// Package combat holds the things a cast creates and the things it hits:
// projectiles, shields, floating damage numbers, and training dummies. All
// drawn with shapes until there is art for them.
package combat

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/streamlinedesign/wizard-engine/engine/ui"
)

// Projectile flies in a straight line until it hits or runs out of range.
type Projectile struct {
	X, Y     float64
	VX, VY   float64
	Size     float64
	Damage   float64
	Left     float64 // range remaining in pixels
	Tint     color.RGBA
	Pierce   int
	age      float64
	hitThing map[*Dummy]bool
}

// Shield absorbs damage around its owner for a while.
type Shield struct {
	HP, MaxHP float64
	Left      float64
	Tint      color.RGBA
}

// Active reports whether the shield still stands.
func (s *Shield) Active() bool { return s != nil && s.HP > 0 && s.Left > 0 }

// FloatText is a number or word drifting up from a hit.
type FloatText struct {
	X, Y float64
	Text string
	Tint color.Color
	Left float64
}

// Dummy is a training target with hit points that resets when destroyed.
type Dummy struct {
	X, Y      float64 // feet
	HP, MaxHP float64
	hitFlash  float64
	respawn   float64
}

// NewDummy stands a dummy at a feet position.
func NewDummy(x, y float64) *Dummy {
	return &Dummy{X: x, Y: y, HP: 40, MaxHP: 40}
}

// Box is the dummy's hit box in world pixels.
func (d *Dummy) Box() (x, y, w, h float64) { return d.X - 6, d.Y - 20, 12, 20 }

// Alive is false while the dummy waits to respawn.
func (d *Dummy) Alive() bool { return d.respawn <= 0 }

// Hit applies damage and returns true if the dummy fell.
func (d *Dummy) Hit(dmg float64) bool {
	if !d.Alive() {
		return false
	}
	d.HP -= dmg
	d.hitFlash = 0.15
	if d.HP <= 0 {
		d.HP = 0
		d.respawn = 2.5
		return true
	}
	return false
}

// Update ticks flash and respawn timers.
func (d *Dummy) Update(dt float64) {
	if d.hitFlash > 0 {
		d.hitFlash -= dt
	}
	if d.respawn > 0 {
		d.respawn -= dt
		if d.respawn <= 0 {
			d.HP = d.MaxHP
		}
	}
}

// Draw renders a post with a straw body and an HP bar.
func (d *Dummy) Draw(dst *ebiten.Image, ox, oy float64) {
	x, y := float32(d.X+ox), float32(d.Y+oy)
	if !d.Alive() {
		// A pile.
		vector.DrawFilledRect(dst, x-6, y-3, 12, 3, color.RGBA{150, 120, 60, 255}, false)
		return
	}
	straw := color.RGBA{210, 180, 90, 255}
	post := color.RGBA{100, 70, 40, 255}
	if d.hitFlash > 0 {
		straw = color.RGBA{255, 240, 220, 255}
	}
	vector.DrawFilledRect(dst, x-1, y-20, 3, 20, post, false)
	vector.DrawFilledRect(dst, x-6, y-18, 13, 10, straw, false)
	vector.DrawFilledRect(dst, x-4, y-24, 9, 6, straw, false)
	vector.DrawFilledRect(dst, x-8, y-15, 4, 3, straw, false)
	vector.DrawFilledRect(dst, x+5, y-15, 4, 3, straw, false)
	// HP bar
	w := float32(16)
	vector.DrawFilledRect(dst, x-8, y-29, w, 3, color.RGBA{40, 30, 40, 255}, false)
	vector.DrawFilledRect(dst, x-8, y-29, w*float32(d.HP/d.MaxHP), 3, color.RGBA{200, 60, 60, 255}, false)
}

// World owns every transient combat thing in a scene.
type World struct {
	Projectiles []*Projectile
	Texts       []*FloatText
	Dummies     []*Dummy
}

// Fire adds a projectile from (x, y) heading along (dx, dy).
func (w *World) Fire(x, y, dx, dy, speed, size, dmg, rng float64, tint [3]uint8) {
	l := math.Hypot(dx, dy)
	if l == 0 {
		dx, dy, l = 0, 1, 1
	}
	w.Projectiles = append(w.Projectiles, &Projectile{
		X: x, Y: y, VX: dx / l * speed, VY: dy / l * speed,
		Size: size, Damage: dmg, Left: rng,
		Tint:     color.RGBA{tint[0], tint[1], tint[2], 255},
		hitThing: map[*Dummy]bool{},
	})
}

// Say adds floating text at a world point.
func (w *World) Say(x, y float64, text string, tint color.Color) {
	w.Texts = append(w.Texts, &FloatText{X: x, Y: y, Text: text, Tint: tint, Left: 0.9})
}

// Update moves projectiles, resolves hits against dummies and ages text.
// blocked reports whether a world point is inside a wall.
func (w *World) Update(dt float64, blocked func(x, y float64) bool) {
	for _, d := range w.Dummies {
		d.Update(dt)
	}
	kept := w.Projectiles[:0]
	for _, p := range w.Projectiles {
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.Left -= math.Hypot(p.VX, p.VY) * dt
		p.age += dt
		alive := p.Left > 0 && !blocked(p.X, p.Y)
		if alive {
			for _, d := range w.Dummies {
				if !d.Alive() || p.hitThing[d] {
					continue
				}
				bx, by, bw, bh := d.Box()
				if p.X+p.Size > bx && p.X-p.Size < bx+bw && p.Y+p.Size > by && p.Y-p.Size < by+bh {
					p.hitThing[d] = true
					fell := d.Hit(p.Damage)
					w.Say(d.X, d.Y-28, fmt.Sprintf("%.0f", p.Damage), color.RGBA{255, 230, 120, 255})
					if fell {
						w.Say(d.X, d.Y-38, "down!", color.RGBA{255, 120, 120, 255})
					}
					if p.Pierce > 0 {
						p.Pierce--
					} else {
						alive = false
						break
					}
				}
			}
		}
		if alive {
			kept = append(kept, p)
		}
	}
	w.Projectiles = kept

	texts := w.Texts[:0]
	for _, t := range w.Texts {
		t.Left -= dt
		t.Y -= 18 * dt
		if t.Left > 0 {
			texts = append(texts, t)
		}
	}
	w.Texts = texts
}

// DrawBelow draws dummies (they are world objects with feet, so the scene
// depth-sorts them; this is called per dummy from the scene).
func (w *World) DrawEffects(dst *ebiten.Image, ox, oy float64) {
	for _, p := range w.Projectiles {
		x, y := float32(p.X+ox), float32(p.Y+oy)
		r := float32(p.Size)
		// Trail
		tx, ty := float32(p.X-p.VX*0.05+ox), float32(p.Y-p.VY*0.05+oy)
		trail := p.Tint
		trail.A = 120
		vector.DrawFilledCircle(dst, tx, ty, r*0.7, trail, false)
		vector.DrawFilledCircle(dst, x, y, r, p.Tint, false)
		core := color.RGBA{255, 250, 220, 255}
		vector.DrawFilledCircle(dst, x, y, r*0.45, core, false)
	}
	for _, t := range w.Texts {
		ui.Print(dst, t.Text, math.Floor(t.X+ox-ui.Width(t.Text)/2), math.Floor(t.Y+oy), t.Tint)
	}
}

// DrawShield draws a bubble around a feet point.
func DrawShield(dst *ebiten.Image, s *Shield, x, y float64) {
	if !s.Active() {
		return
	}
	c := s.Tint
	c.A = 90
	vector.DrawFilledCircle(dst, float32(x), float32(y-14), 16, c, false)
	edge := s.Tint
	edge.A = 220
	vector.StrokeCircle(dst, float32(x), float32(y-14), 16, 1, edge, false)
}
