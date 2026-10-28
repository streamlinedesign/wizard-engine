// Package scenes holds the game's screens. Sandbox is the step-1 proving
// ground: a snowy clearing to walk around in.
package scenes

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/streamlinedesign/wizard-engine/assets"
	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/render"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
	"github.com/streamlinedesign/wizard-engine/game/player"
)

// Tile ids in assets/tiles/winter.png (see tools/mkplaceholders).
const (
	tSnow = iota + 1
	tSnowSpeckled
	tPath
	tIce
	tWater
	tRock
	tPineTop
	tPineBottom
	tFence
	tSnowEdge
	tShrub
	tLamp
)

// Sandbox is a small winter clearing with a frozen pond and some pines.
type Sandbox struct {
	Map    *tilemap.Map
	Player *player.Player
	Debug  bool
}

// Enter builds the map and drops the wizard in the middle.
func (s *Sandbox) Enter(ctx *engine.Context) {
	tiles := assets.Must(assets.Image("tiles/winter.png"))
	ts := tilemap.NewTileset(tiles, 16, tWater, tRock, tPineBottom, tFence, tShrub, tLamp)
	s.Map = buildClearing(ts, 40, 24)

	spr := assets.Must(assets.Sprite("sprites/wizard.json"))
	w, h := s.Map.PixelSize()
	s.Player = player.New(spr, float64(w)/2, float64(h)/2+24)

	ctx.Camera.Bounds = &render.Rect{W: float64(w), H: float64(h)}
	ctx.Camera.Center(s.Player.X, s.Player.Y)
}

// Update advances the wizard and the camera.
func (s *Sandbox) Update(ctx *engine.Context) error {
	s.Player.Update(ctx, s.Map)
	ctx.Camera.Follow(s.Player.X, s.Player.Y-8)
	if ctx.Input.JustPressed("debug") {
		s.Debug = !s.Debug
	}
	return nil
}

// Draw renders ground, the wizard, then anything that should overlap him.
func (s *Sandbox) Draw(ctx *engine.Context, canvas *ebiten.Image) {
	ox, oy := ctx.Camera.Offset()
	v := ctx.Camera.Visible()
	s.Map.Draw(canvas, ox, oy, v.X, v.Y, v.W, v.H)
	s.Player.Draw(canvas, ox, oy)
	if s.Debug {
		ebitenutil.DebugPrint(canvas, fmt.Sprintf("tps %.0f fps %.0f\npos %.0f,%.0f", ebiten.ActualTPS(), ebiten.ActualFPS(), s.Player.X, s.Player.Y))
	}
}

// buildClearing lays out a deterministic little map: snow with a path, a
// pond with an icy rim, pines around the edge, and a fence by a lamp.
func buildClearing(ts *tilemap.Tileset, w, h int) *tilemap.Map {
	rng := rand.New(rand.NewSource(42))
	m := tilemap.NewMap(ts, w, h, 2)
	const ground, props = 0, 1

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			id := tSnow
			if rng.Intn(5) == 0 {
				id = tSnowSpeckled
			}
			m.Set(ground, x, y, id)
		}
	}

	// Rocky border so there is a visible edge to the world.
	for x := 0; x < w; x++ {
		m.Set(props, x, 0, tRock)
		m.Set(props, x, h-1, tRock)
	}
	for y := 0; y < h; y++ {
		m.Set(props, 0, y, tRock)
		m.Set(props, w-1, y, tRock)
	}

	// Pond in the top-left quarter.
	for y := 4; y < 10; y++ {
		for x := 5; x < 14; x++ {
			edge := y == 4 || y == 9 || x == 5 || x == 13
			switch {
			case edge:
				m.Set(ground, x, y, tIce)
			default:
				m.Set(ground, x, y, tWater)
			}
		}
	}
	for x := 5; x < 14; x++ {
		m.Set(ground, x, 3, tSnowEdge)
	}

	// Path from the bottom edge up through the middle and off to the right.
	for y := h / 2; y < h-1; y++ {
		m.Set(ground, w/2, y, tPath)
		m.Set(ground, w/2+1, y, tPath)
	}
	for x := w / 2; x < w-1; x++ {
		m.Set(ground, x, h/2, tPath)
		m.Set(ground, x, h/2+1, tPath)
	}

	// Pines scattered where there is nothing else.
	for i := 0; i < 28; i++ {
		x, y := 2+rng.Intn(w-4), 2+rng.Intn(h-5)
		if m.Get(ground, x, y) != tSnow && m.Get(ground, x, y) != tSnowSpeckled {
			continue
		}
		if m.Get(props, x, y) != 0 || m.Get(props, x, y+1) != 0 || m.Get(ground, x, y+1) == tPath {
			continue
		}
		if abs(x-w/2) < 3 && abs(y-h/2) < 4 {
			continue // keep the spawn clear
		}
		m.Set(props, x, y, tPineTop)
		m.Set(props, x, y+1, tPineBottom)
	}

	// A fence line and a lamp near the path junction.
	for x := w/2 + 4; x < w/2+9; x++ {
		m.Set(props, x, h/2-2, tFence)
	}
	m.Set(props, w/2-2, h/2-1, tLamp)
	m.Set(props, w/2+3, h/2+3, tShrub)
	m.Set(props, w/2-4, h/2+4, tShrub)

	return m
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
