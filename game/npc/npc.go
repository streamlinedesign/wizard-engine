// Package npc is a character who stands somewhere, faces you when you are
// close, and talks when you press interact. Signs are NPCs without a
// sprite.
package npc

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine/sprite"
)

// NPC is a talking thing in the world.
type NPC struct {
	Name string
	// X, Y is the feet position, like the player.
	X, Y float64
	// Dialogue is the id of the current conversation; dialogues can
	// replace it via their "then".
	Dialogue string
	// Sprite and Anim are nil for signs and other sprite-less talkers.
	Sprite *sprite.Sprite
	Anim   *sprite.Animator
	// Solid NPCs block movement with a box like the player's.
	BoxW, BoxH float64

	facingLeft bool
}

// Box returns the collision rectangle at the feet.
func (n *NPC) Box() (x, y, w, h float64) {
	return n.X - n.BoxW/2, n.Y - n.BoxH, n.BoxW, n.BoxH
}

// InRange reports whether a point (the player's feet) is close enough to
// talk: within reach pixels of the NPC's feet.
func (n *NPC) InRange(px, py, reach float64) bool {
	return math.Hypot(px-n.X, py-n.Y) <= reach
}

// Update turns the NPC toward the player when they are near.
func (n *NPC) Update(dt, px, py float64) {
	if n.Anim == nil {
		return
	}
	dx, dy := px-n.X, py-n.Y
	tag := "idle_down"
	if math.Hypot(dx, dy) < 40 {
		switch {
		case math.Abs(dx) > math.Abs(dy):
			tag = "idle_side"
			n.facingLeft = dx < 0
		case dy < 0:
			tag = "idle_up"
		}
	}
	n.Anim.Play(tag)
	n.Anim.Advance(dt)
}

// Draw blits the NPC with its feet on (X, Y).
func (n *NPC) Draw(dst *ebiten.Image, ox, oy float64) {
	if n.Sprite == nil {
		return
	}
	fr := n.Anim.Frame()
	x := math.Floor(n.X - float64(fr.Size.X)/2 + ox)
	y := math.Floor(n.Y - float64(fr.Size.Y) + oy)
	n.Sprite.Draw(dst, fr, x, y, n.facingLeft)
}

// Overlaps reports whether a world rectangle touches this NPC's box.
func (n *NPC) Overlaps(x, y, w, h float64) bool {
	if n.BoxW == 0 || n.BoxH == 0 {
		return false
	}
	bx, by, bw, bh := n.Box()
	return x < bx+bw && x+w > bx && y < by+bh && y+h > by
}
