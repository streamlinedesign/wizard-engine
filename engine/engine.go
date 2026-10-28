// Package engine is the thin game loop that ties the subsystems together:
// a fixed-tick Update, a Draw into the pixel canvas, and one active Scene.
// Ebitengine supplies the window, input and GPU; everything game-specific
// lives in scenes and the packages under engine/.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/render"
)

// Config describes the window and canvas.
type Config struct {
	Title string
	// CanvasWidth/Height is the virtual resolution the game draws at.
	CanvasWidth, CanvasHeight int
	// WindowScale is the initial integer zoom of the window.
	WindowScale int
	// TPS is simulation ticks per second. 0 means 60.
	TPS int
}

// Context is what a scene gets every tick: input, camera, timing.
type Context struct {
	Input    *input.State
	Camera   *render.Camera
	Viewport *render.Viewport
	// DT is the fixed tick length in seconds.
	DT float64
	// Tick counts simulation ticks since start.
	Tick uint64

	game *Game
}

// SetScene swaps the active scene at the end of the current tick.
func (c *Context) SetScene(s Scene) { c.game.next = s }

// Scene is one screen of the game: the hub, a dungeon, a menu.
type Scene interface {
	// Enter runs once when the scene becomes active.
	Enter(ctx *Context)
	Update(ctx *Context) error
	Draw(ctx *Context, canvas *ebiten.Image)
}

// Game implements ebiten.Game.
type Game struct {
	cfg      Config
	ctx      *Context
	scene    Scene
	next     Scene
	viewport *render.Viewport

	// AfterDraw, if set, is called with the finished canvas each frame.
	// Used for screenshots and tests.
	AfterDraw func(canvas *ebiten.Image, tick uint64)
}

// New prepares a game around an initial scene.
func New(cfg Config, scene Scene) *Game {
	if cfg.TPS == 0 {
		cfg.TPS = 60
	}
	if cfg.WindowScale < 1 {
		cfg.WindowScale = 3
	}
	g := &Game{cfg: cfg}
	g.viewport = render.NewViewport(cfg.CanvasWidth, cfg.CanvasHeight)
	g.ctx = &Context{
		Input:    input.New(),
		Camera:   render.NewCamera(cfg.CanvasWidth, cfg.CanvasHeight),
		Viewport: g.viewport,
		DT:       1 / float64(cfg.TPS),
		game:     g,
	}
	g.next = scene
	return g
}

// Context exposes the shared context (for tests and tools).
func (g *Game) Context() *Context { return g.ctx }

// Run opens the window and blocks until the game exits.
func (g *Game) Run() error {
	ebiten.SetWindowTitle(g.cfg.Title)
	ebiten.SetWindowSize(g.cfg.CanvasWidth*g.cfg.WindowScale, g.cfg.CanvasHeight*g.cfg.WindowScale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(g.cfg.TPS)
	return ebiten.RunGame(g)
}

// Update is one simulation tick.
func (g *Game) Update() error {
	if g.next != nil {
		g.scene, g.next = g.next, nil
		g.scene.Enter(g.ctx)
	}
	if err := g.scene.Update(g.ctx); err != nil {
		return err
	}
	g.ctx.Tick++
	return nil
}

// Draw renders the scene into the canvas, then the canvas onto the window.
func (g *Game) Draw(screen *ebiten.Image) {
	canvas := g.viewport.Canvas()
	canvas.Clear()
	if g.scene != nil {
		g.scene.Draw(g.ctx, canvas)
	}
	if g.AfterDraw != nil {
		g.AfterDraw(canvas, g.ctx.Tick)
	}
	g.viewport.Present(screen)
}

// Layout tells Ebitengine the screen is the real window size; the viewport
// does its own integer scaling from there.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
