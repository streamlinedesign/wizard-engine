package combat

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine/sprite"
)

// FX is a one-shot sprite animation (an impact, a heal, a teleport puff)
// drawn with the centre of its cell on (X, Y).
type FX struct {
	X, Y float64
	Spr  *sprite.Sprite
	Anim *sprite.Animator
}

// Play starts effect name's tag at a world point. Unknown names are ignored,
// so a scene without effect art still runs.
func (w *World) Play(name, tag string, x, y float64) {
	spr := w.Effects[name]
	if spr == nil {
		return
	}
	if _, ok := spr.Sheet.Tags[tag]; !ok {
		return
	}
	w.FX = append(w.FX, &FX{X: x, Y: y, Spr: spr, Anim: sprite.NewAnimator(spr.Sheet, tag)})
}

func (w *World) updateFX(dt float64) {
	kept := w.FX[:0]
	for _, f := range w.FX {
		f.Anim.Advance(dt)
		if !f.Anim.Finished() {
			kept = append(kept, f)
		}
	}
	w.FX = kept
}

func (w *World) drawFX(dst *ebiten.Image, ox, oy float64) {
	for _, f := range w.FX {
		fr := f.Anim.Frame()
		x := math.Floor(f.X + ox - float64(fr.Size.X)/2)
		y := math.Floor(f.Y + oy - float64(fr.Size.Y)/2)
		f.Spr.Draw(dst, fr, x, y, false)
	}
}

// frameAt picks the frame of a looping tag after t seconds, for things
// that keep their own clock (a projectile's age, a shield's time left).
func frameAt(s *sprite.Sheet, tag string, t float64) (sprite.Frame, bool) {
	tg, ok := s.Tags[tag]
	if !ok {
		return sprite.Frame{}, false
	}
	total := 0.0
	for i := tg.From; i <= tg.To; i++ {
		total += float64(s.Frames[i].DurationMs)
	}
	if total <= 0 {
		return s.Frames[tg.From], true
	}
	ms := math.Mod(t*1000, total)
	for i := tg.From; i <= tg.To; i++ {
		ms -= float64(s.Frames[i].DurationMs)
		if ms < 0 {
			return s.Frames[i], true
		}
	}
	return s.Frames[tg.To], true
}

// drawRotated blits a frame centred on (x, y), turned by angle radians.
func drawRotated(dst *ebiten.Image, spr *sprite.Sprite, fr sprite.Frame, x, y, angle float64) {
	sub := spr.Image.SubImage(fr.Rect).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(fr.Size.X)/2+float64(fr.Offset.X), -float64(fr.Size.Y)/2+float64(fr.Offset.Y))
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(math.Floor(x), math.Floor(y))
	dst.DrawImage(sub, op)
}

// projectileArt maps an element to the effect sheet that draws its
// projectile. Elements without art keep the drawn circle.
var projectileArt = map[string]string{"fire": "fireball"}
