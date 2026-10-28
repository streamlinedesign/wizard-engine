package scenes

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/streamlinedesign/wizard-engine/assets"
	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/render"
	"github.com/streamlinedesign/wizard-engine/engine/sprite"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
	"github.com/streamlinedesign/wizard-engine/game/dialogue"
	"github.com/streamlinedesign/wizard-engine/game/hud"
	"github.com/streamlinedesign/wizard-engine/game/items"
	"github.com/streamlinedesign/wizard-engine/game/npc"
	"github.com/streamlinedesign/wizard-engine/game/player"
)

// Tile ids in assets/tiles/campus.png (see tools/mkplaceholders/campus.go).
const (
	cGrass = iota + 1
	cGrassWorn
	cGrassFlowers
	cPath
	cPlaza
	cWater
	cShore
	cWall
	cWallTop
	cRoof
	cRoofEdge
	cDoor
	cWindow
	cTreeTop
	cTreeBottom
	cHedge
	cLamp
	cFountain
	cBench
	cSignpost
	cBookshelf
	cDesk
	cWoodFloor
	cRug
)

// Layers of the campus map.
const (
	layerGround = iota
	layerProps
	layerOverhead // drawn above everything: roofs, tree tops
)

// Campus is the hub: an open university of magic. Step 2 has the grounds,
// the old wizard, signs, and the inventory; buildings are facades until
// their systems exist.
type Campus struct {
	Map    *tilemap.Map
	Player *player.Player
	NPCs   []*npc.NPC

	Items     *items.Registry
	Inventory items.Inventory
	Dialogues *dialogue.Registry
	Talk      *dialogue.Runner
	Panel     *hud.InventoryPanel

	// StartX, StartY override the spawn tile when non-zero (dev flag).
	StartX, StartY int
	// StartDialogue opens a dialogue on entry, for previewing text.
	StartDialogue string
	// StartInventory grants the old wizard's gifts and opens the inventory
	// on entry, for previewing the panel.
	StartInventory bool

	nearest *npc.NPC
	Debug   bool
}

// Enter loads data, builds the map and places everyone.
func (s *Campus) Enter(ctx *engine.Context) {
	s.Items = assets.Must(items.ParseRegistry(assets.Must(assets.Data("data/items.json"))))
	s.Dialogues = assets.Must(dialogue.ParseRegistry(assets.Must(assets.Data("data/dialogue.json"))))

	tiles := assets.Must(assets.Image("tiles/campus.png"))
	ts := tilemap.NewTileset(tiles, 16,
		cWater, cWall, cWallTop, cDoor, cWindow, cTreeBottom, cHedge, cLamp, cFountain, cBench, cSignpost, cBookshelf, cDesk)
	s.Map = buildCampus(ts)

	icons := tilemap.NewTileset(assets.Must(assets.Image("icons/items.png")), 16)
	s.Panel = hud.NewInventoryPanel(icons)

	spr := assets.Must(assets.Sprite("sprites/wizard.json"))
	s.Player = player.New(spr, 30*16+8, 44*16)
	if s.StartX != 0 || s.StartY != 0 {
		s.Player.X, s.Player.Y = float64(s.StartX*16+8), float64(s.StartY*16+12)
	}
	s.Player.Blocked = s.blockedByNPC

	old := assets.Must(assets.Sprite("sprites/oldwizard.json"))
	s.NPCs = []*npc.NPC{
		{Name: "Old wizard", X: 34*16 + 8, Y: 24*16 + 12, Dialogue: "old_wizard_intro",
			Sprite: old, Anim: sprite.NewAnimator(old.Sheet, "idle_down"), BoxW: 10, BoxH: 6},
		{Name: "Sign", X: 28*16 + 8, Y: 42*16 + 12, Dialogue: "sign_gate"},
		{Name: "Sign", X: 20*16 + 8, Y: 12*16 + 12, Dialogue: "sign_scriptorium"},
		{Name: "Sign", X: 42*16 + 8, Y: 12*16 + 12, Dialogue: "sign_library"},
		{Name: "Sign", X: 54*16 + 8, Y: 30*16 + 12, Dialogue: "sign_harbour"},
	}

	s.Talk = dialogue.NewRunner(float64(ctx.Viewport.Width) - 20)
	s.Talk.OnGive = dialogue.GiftsTo(s.Items, &s.Inventory)
	s.Talk.OnEnd = func(d *dialogue.Def) {
		if d.Then != "" && s.nearest != nil {
			s.nearest.Dialogue = d.Then
		}
	}

	w, h := s.Map.PixelSize()
	ctx.Camera.Bounds = &render.Rect{W: float64(w), H: float64(h)}
	ctx.Camera.Center(s.Player.X, s.Player.Y)

	if d, ok := s.Dialogues.Get(s.StartDialogue); ok {
		s.Talk.Start(d)
	}
	if s.StartInventory {
		if d, ok := s.Dialogues.Get("old_wizard_intro"); ok {
			for _, g := range d.Give {
				s.Talk.OnGive(g)
			}
		}
		s.Panel.Open = true
	}
}

func (s *Campus) blockedByNPC(x, y, w, h float64) bool {
	for _, n := range s.NPCs {
		if n.Overlaps(x, y, w, h) {
			return true
		}
	}
	return false
}

// Update runs dialogue or the inventory when they are open, otherwise the
// world.
func (s *Campus) Update(ctx *engine.Context) error {
	if ctx.Input.JustPressed("debug") {
		s.Debug = !s.Debug
	}
	for _, n := range s.NPCs {
		n.Update(ctx.DT, s.Player.X, s.Player.Y)
	}
	if s.Talk.Active() {
		s.Talk.Update(ctx)
		return nil
	}
	if s.Panel.Update(ctx, &s.Inventory) {
		return nil
	}

	s.Player.Update(ctx, s.Map)
	ctx.Camera.Follow(s.Player.X, s.Player.Y-8)

	s.nearest = nil
	best := 26.0 // talking reach in pixels, a little under two tiles
	for _, n := range s.NPCs {
		if n.InRange(s.Player.X, s.Player.Y, best) {
			s.nearest = n
			best = dist(n.X, n.Y, s.Player.X, s.Player.Y)
		}
	}
	if s.nearest != nil && ctx.Input.JustPressed(input.Interact) {
		if d, ok := s.Dialogues.Get(s.nearest.Dialogue); ok {
			s.Talk.Start(d)
		}
	}
	return nil
}

func dist(ax, ay, bx, by float64) float64 {
	return math.Hypot(ax-bx, ay-by)
}

// Draw renders ground, then props and characters interleaved by row so the
// wizard walks behind what is below him on screen, then the overhead layer.
func (s *Campus) Draw(ctx *engine.Context, canvas *ebiten.Image) {
	ox, oy := ctx.Camera.Offset()
	v := ctx.Camera.Visible()
	s.Map.DrawLayerRows(canvas, layerGround, ox, oy, v.X, v.Y, v.W, v.H, 0, s.Map.Height-1)

	// Characters sorted by feet y, merged with prop rows.
	chars := make([]drawable, 0, len(s.NPCs)+1)
	chars = append(chars, drawable{s.Player.Y, func() { s.Player.Draw(canvas, ox, oy) }})
	for _, n := range s.NPCs {
		if n.Sprite != nil {
			nn := n
			chars = append(chars, drawable{n.Y, func() { nn.Draw(canvas, ox, oy) }})
		}
	}
	sortDrawables(chars)
	row := s.Map.RowOf(v.Y)
	for _, c := range chars {
		// Props whose row is at or above the character's feet row draw first;
		// a prop on the same row is "behind" when the character stands lower
		// in it, which is the common case for a tall sprite.
		cr := s.Map.RowOf(c.y - 1)
		if cr >= row {
			s.Map.DrawLayerRows(canvas, layerProps, ox, oy, v.X, v.Y, v.W, v.H, row, cr)
			row = cr + 1
		}
		c.draw()
	}
	s.Map.DrawLayerRows(canvas, layerProps, ox, oy, v.X, v.Y, v.W, v.H, row, s.Map.Height-1)
	s.Map.DrawLayerRows(canvas, layerOverhead, ox, oy, v.X, v.Y, v.W, v.H, 0, s.Map.Height-1)

	if s.nearest != nil && !s.Talk.Active() && !s.Panel.Open {
		label := "E: Talk"
		if s.nearest.Sprite == nil {
			label = "E: Read"
		}
		hud.Prompt(canvas, label, s.nearest.X+ox, s.nearest.Y+oy)
	}
	s.Talk.Draw(canvas)
	s.Panel.Draw(canvas, &s.Inventory)

	if s.Debug {
		ebitenutil.DebugPrint(canvas, fmt.Sprintf("tps %.0f fps %.0f\npos %.0f,%.0f tile %d,%d",
			ebiten.ActualTPS(), ebiten.ActualFPS(), s.Player.X, s.Player.Y, int(s.Player.X)/16, int(s.Player.Y)/16))
	}
}

type drawable struct {
	y    float64
	draw func()
}

func sortDrawables(d []drawable) {
	// Insertion sort: there are a handful of characters.
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j].y < d[j-1].y; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}
