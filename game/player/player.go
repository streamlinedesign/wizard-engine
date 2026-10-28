// Package player is the wizard the person controls: movement, facing and
// animation. Spells, inventory and the grimoire will hang off this later.
package player

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/sprite"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
)

// Facing is the direction the wizard last moved.
type Facing int

const (
	Down Facing = iota
	Up
	Left
	Right
)

// Player is positioned by its feet: (X, Y) is the bottom-centre of the
// sprite, which is what you want for depth sorting and tile collision.
type Player struct {
	X, Y   float64
	Speed  float64 // pixels per second
	Facing Facing

	// Collision box, relative to the feet point.
	BoxW, BoxH float64

	Sprite *sprite.Sprite
	Anim   *sprite.Animator
	moving bool
}

// New places a wizard at a feet position.
func New(spr *sprite.Sprite, x, y float64) *Player {
	return &Player{
		X: x, Y: y,
		Speed:  64,
		BoxW:   10,
		BoxH:   6,
		Sprite: spr,
		Anim:   sprite.NewAnimator(spr.Sheet, "idle_down"),
	}
}

// Box returns the collision rectangle in world pixels.
func (p *Player) Box() (x, y, w, h float64) {
	return p.X - p.BoxW/2, p.Y - p.BoxH, p.BoxW, p.BoxH
}

// Update reads input, moves with sliding collision, and picks an animation.
func (p *Player) Update(ctx *engine.Context, m *tilemap.Map) {
	dx, dy := ctx.Input.Axis()
	p.moving = dx != 0 || dy != 0

	if p.moving {
		step := p.Speed * ctx.DT
		p.move(m, dx*step, 0)
		p.move(m, 0, dy*step)
		p.Facing = facingFrom(dx, dy, p.Facing)
	}

	tag := "idle_"
	if p.moving {
		tag = "walk_"
	}
	switch p.Facing {
	case Up:
		tag += "up"
	case Left, Right:
		tag += "side"
	default:
		tag += "down"
	}
	p.Anim.Play(tag)
	p.Anim.Advance(ctx.DT)
}

// move tries a displacement along one axis and backs off to the last free
// pixel if it hits something, so the wizard slides along walls.
func (p *Player) move(m *tilemap.Map, dx, dy float64) {
	if dx == 0 && dy == 0 {
		return
	}
	nx, ny := p.X+dx, p.Y+dy
	bx, by, bw, bh := nx-p.BoxW/2, ny-p.BoxH, p.BoxW, p.BoxH
	if !m.Collides(bx, by, bw, bh) {
		p.X, p.Y = nx, ny
		return
	}
	// Snap to the tile edge in the direction of travel.
	ts := float64(m.Tileset.TileSize)
	switch {
	case dx > 0:
		edge := math.Floor((bx+bw)/ts) * ts
		p.X = edge - p.BoxW/2 - 0.001
	case dx < 0:
		edge := math.Ceil(bx/ts) * ts
		p.X = edge + p.BoxW/2
	case dy > 0:
		edge := math.Floor((by+bh)/ts) * ts
		p.Y = edge - 0.001
	case dy < 0:
		edge := math.Ceil(by/ts) * ts
		p.Y = edge + p.BoxH
	}
}

// facingFrom prefers the horizontal direction on diagonals, unless the
// player was already facing vertically and is still moving that way.
func facingFrom(dx, dy float64, prev Facing) Facing {
	if dx != 0 && dy != 0 {
		if (prev == Up && dy < 0) || (prev == Down && dy > 0) {
			return prev
		}
	}
	switch {
	case dx < 0:
		return Left
	case dx > 0:
		return Right
	case dy < 0:
		return Up
	case dy > 0:
		return Down
	}
	return prev
}

// Draw blits the current frame so the feet sit on (X, Y), offset by the
// camera translation.
func (p *Player) Draw(dst *ebiten.Image, ox, oy float64) {
	fr := p.Anim.Frame()
	x := math.Floor(p.X - float64(fr.Size.X)/2 + ox)
	y := math.Floor(p.Y - float64(fr.Size.Y) + oy)
	p.Sprite.Draw(dst, fr, x, y, p.Facing == Left)
}
