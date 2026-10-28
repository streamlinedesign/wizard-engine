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

	// Vitals. Mana refills over time; health only through spells for now.
	HP, MaxHP     float64
	Mana, MaxMana float64
	ManaRegen     float64 // per second

	Sprite *sprite.Sprite
	Anim   *sprite.Animator
	// Blocked, when set, reports whether a world rectangle hits something
	// that is not a tile (an NPC, a prop). Such hits stop the move outright.
	Blocked func(x, y, w, h float64) bool
	moving  bool
}

// New places a wizard at a feet position.
func New(spr *sprite.Sprite, x, y float64) *Player {
	return &Player{
		X: x, Y: y,
		Speed: 64,
		BoxW:  10,
		BoxH:  6,
		HP:    30, MaxHP: 30,
		Mana: 20, MaxMana: 20,
		ManaRegen: 2.5,
		Sprite:    spr,
		Anim:      sprite.NewAnimator(spr.Sheet, "idle_down"),
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
	p.Mana = math.Min(p.MaxMana, p.Mana+p.ManaRegen*ctx.DT)
}

// Facing as a unit vector.
func (p *Player) Dir() (dx, dy float64) {
	switch p.Facing {
	case Up:
		return 0, -1
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	}
	return 0, 1
}

// Spend takes mana if there is enough and reports whether it did.
func (p *Player) Spend(mana float64) bool {
	if p.Mana < mana {
		return false
	}
	p.Mana -= mana
	return true
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
		if p.Blocked != nil && p.Blocked(bx, by, bw, bh) {
			return
		}
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
