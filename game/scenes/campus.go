package scenes

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/streamlinedesign/wizard-engine/assets"
	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/render"
	"github.com/streamlinedesign/wizard-engine/engine/sprite"
	"github.com/streamlinedesign/wizard-engine/engine/tilemap"
	"github.com/streamlinedesign/wizard-engine/game/combat"
	"github.com/streamlinedesign/wizard-engine/game/dialogue"
	"github.com/streamlinedesign/wizard-engine/game/grimoire"
	"github.com/streamlinedesign/wizard-engine/game/hud"
	"github.com/streamlinedesign/wizard-engine/game/items"
	"github.com/streamlinedesign/wizard-engine/game/npc"
	"github.com/streamlinedesign/wizard-engine/game/player"
	"github.com/streamlinedesign/wizard-engine/game/spells"
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

// The ritual circle: centre tile and radius in pixels.
const (
	circleTileX, circleTileY = 24, 17
	circleRadius             = 36.0
)

// Campus is the hub: an open university of magic. The grounds, the old
// wizard, signs, the inventory, the ritual circle for binding spells, and
// training dummies to cast at. Buildings are facades until their systems
// exist.
type Campus struct {
	Map    *tilemap.Map
	Player *player.Player
	NPCs   []*npc.NPC

	Items     *items.Registry
	Inventory items.Inventory
	Dialogues *dialogue.Registry
	Spells    *spells.Table
	Book      *grimoire.Book
	Talk      *dialogue.Runner
	Panel     *hud.InventoryPanel
	Circle    *hud.Binding
	Combat    combat.World
	Shield    *combat.Shield

	// StartX, StartY override the spawn tile when non-zero (dev flag).
	StartX, StartY int
	// StartDialogue opens a dialogue on entry, for previewing text.
	StartDialogue string
	// StartInventory grants the old wizard's gifts and opens the inventory
	// on entry, for previewing the panel.
	StartInventory bool
	// StartBinding grants the gifts and opens the circle with a fireball and
	// a scroll placed, for previewing the screen.
	StartBinding bool
	// StartBound grants the gifts and binds the four starters so casting
	// works immediately.
	StartBound bool
	// AutoCast casts the active spell every 40 ticks (screenshots).
	AutoCast bool

	nearest  *npc.NPC
	inCircle bool
	circleX  float64
	circleY  float64
	Debug    bool
}

// Enter loads data, builds the map and places everyone.
func (s *Campus) Enter(ctx *engine.Context) {
	s.Items = assets.Must(items.ParseRegistry(assets.Must(assets.Data("data/items.json"))))
	s.Dialogues = assets.Must(dialogue.ParseRegistry(assets.Must(assets.Data("data/dialogue.json"))))
	s.Spells = assets.Must(spells.ParseTable(assets.Must(assets.Data("data/spells.json"))))
	s.Book = grimoire.New("Uriel")

	tiles := assets.Must(assets.Image("tiles/campus.png"))
	ts := tilemap.NewTileset(tiles, 16,
		cWater, cWall, cWallTop, cDoor, cWindow, cTreeBottom, cHedge, cLamp, cFountain, cBench, cSignpost, cBookshelf, cDesk)
	s.Map = buildCampus(ts)
	s.circleX, s.circleY = circleTileX*16+8, circleTileY*16+8

	icons := tilemap.NewTileset(assets.Must(assets.Image("icons/items.png")), 16)
	s.Panel = hud.NewInventoryPanel(icons)
	s.Circle = hud.NewBinding(icons, s.Spells)
	s.Circle.OnBind = func(sp spells.Spell) {
		s.Book.Bind(sp)
	}

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
	s.Combat.Dummies = []*combat.Dummy{
		combat.NewDummy(36*16+8, 23*16+14),
		combat.NewDummy(37*16+8, 26*16+14),
		combat.NewDummy(36*16+8, 29*16+14),
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

	s.applyDevStart()
}

// applyDevStart handles the preview flags.
func (s *Campus) applyDevStart() {
	if d, ok := s.Dialogues.Get(s.StartDialogue); ok {
		s.Talk.Start(d)
	}
	if s.StartInventory || s.StartBinding || s.StartBound {
		if d, ok := s.Dialogues.Get("old_wizard_intro"); ok {
			for _, g := range d.Give {
				s.Talk.OnGive(g)
			}
		}
	}
	if s.StartInventory {
		s.Panel.Open = true
	}
	if s.StartBound {
		for _, id := range []string{"spell_fireball", "spell_small_heal", "spell_minor_teleport", "spell_barrier"} {
			d := s.Items.Must(id)
			if s.Inventory.Remove(d, 1) && s.Inventory.Remove(s.Items.Must("scroll_empty"), 1) {
				s.Book.Bind(s.Spells.Resolve(d.Keywords))
			}
		}
	}
	if s.StartBinding {
		s.Circle.Open = true
		s.Circle.Center = s.Items.Must("scroll_empty")
		s.Inventory.Remove(s.Circle.Center, 1)
		fb := s.Items.Must("spell_fireball")
		s.Inventory.Remove(fb, 1)
		s.Circle.Ring[0] = fb
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

// Update runs whichever screen is open, otherwise the world.
func (s *Campus) Update(ctx *engine.Context) error {
	if ctx.Input.JustPressed("debug") {
		s.Debug = !s.Debug
	}
	for _, n := range s.NPCs {
		n.Update(ctx.DT, s.Player.X, s.Player.Y)
	}
	s.Combat.Update(ctx.DT, s.solidAt)
	if s.Shield.Active() {
		s.Shield.Left -= ctx.DT
	}

	if s.Talk.Active() {
		s.Talk.Update(ctx)
		return nil
	}
	if s.Circle.Update(ctx, ctx.Viewport.Canvas(), &s.Inventory) {
		return nil
	}
	if s.Panel.Update(ctx, &s.Inventory) {
		return nil
	}

	s.Player.Update(ctx, s.Map)
	ctx.Camera.Follow(s.Player.X, s.Player.Y-8)

	// Standing in the circle's centre tile.
	s.inCircle = math.Abs(s.Player.X-s.circleX) < 8 && math.Abs(s.Player.Y-s.circleY-4) < 8
	if s.inCircle && ctx.Input.JustPressed(input.Interact) {
		s.Circle.Open = true
		return nil
	}

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

	// Quick slots and casting.
	for i, a := range []input.Action{input.Slot1, input.Slot2, input.Slot3, input.Slot4} {
		if ctx.Input.JustPressed(a) {
			s.Book.Selected = i
		}
	}
	if ctx.Input.JustPressed(input.Cast) || ctx.Input.Clicked() || (s.AutoCast && ctx.Tick%40 == 0) {
		s.cast(ctx)
	}
	return nil
}

// solidAt reports whether a world point is inside a solid tile.
func (s *Campus) solidAt(x, y float64) bool {
	return s.Map.SolidAt(int(math.Floor(x/16)), int(math.Floor(y/16)))
}

// aim returns a unit direction from the player toward the mouse when the
// cursor is on the canvas, otherwise the facing direction. AutoCast aims at
// the nearest standing dummy so screenshots show hits.
func (s *Campus) aim(ctx *engine.Context) (dx, dy float64) {
	if s.AutoCast {
		var best *combat.Dummy
		bd := math.Inf(1)
		for _, d := range s.Combat.Dummies {
			if dd := dist(d.X, d.Y, s.Player.X, s.Player.Y); d.Alive() && dd < bd {
				best, bd = d, dd
			}
		}
		if best != nil {
			return input.Normalize(best.X-s.Player.X, best.Y-10-(s.Player.Y-12))
		}
	}
	mx, my := ctx.Viewport.Cursor()
	if mx >= 0 && my >= 0 && mx < ctx.Viewport.Width && my < ctx.Viewport.Height {
		v := ctx.Camera.Visible()
		wx, wy := v.X+float64(mx), v.Y+float64(my)
		dx, dy = wx-s.Player.X, wy-(s.Player.Y-12)
		if math.Hypot(dx, dy) > 4 {
			return input.Normalize(dx, dy)
		}
	}
	return s.Player.Dir()
}

// cast fires the active spell.
func (s *Campus) cast(ctx *engine.Context) {
	sp := s.Book.Active()
	p := s.Player
	if sp == nil {
		s.Combat.Say(p.X, p.Y-36, "no spell bound", ui_dim)
		return
	}
	if !p.Spend(sp.Mana) {
		s.Combat.Say(p.X, p.Y-36, "no mana", ui_dim)
		return
	}
	tint := color.RGBA{sp.Tint[0], sp.Tint[1], sp.Tint[2], 255}
	switch sp.Shape {
	case spells.ShapeProjectile:
		dx, dy := s.aim(ctx)
		n := max(sp.Count, 1)
		spread := 0.18 // radians between shots
		for i := 0; i < n; i++ {
			a := math.Atan2(dy, dx) + (float64(i)-float64(n-1)/2)*spread
			s.Combat.Fire(p.X, p.Y-12, math.Cos(a), math.Sin(a), sp.Speed, sp.Size, sp.Damage, sp.Range, sp.Tint)
		}
	case spells.ShapeSelf:
		before := p.HP
		p.HP = math.Min(p.MaxHP, p.HP+sp.Heal)
		s.Combat.Say(p.X, p.Y-36, fmt.Sprintf("+%.0f", p.HP-before), tint)
	case spells.ShapeBlink:
		dx, dy := s.aim(ctx)
		// Walk the path in small steps and stop before the first wall.
		bx, by := p.X, p.Y
		for d := 4.0; d <= sp.Range; d += 4 {
			nx, ny := p.X+dx*d, p.Y+dy*d
			if s.Map.Collides(nx-p.BoxW/2, ny-p.BoxH, p.BoxW, p.BoxH) || s.blockedByNPC(nx-p.BoxW/2, ny-p.BoxH, p.BoxW, p.BoxH) {
				break
			}
			bx, by = nx, ny
		}
		s.Combat.Say(p.X, p.Y-30, "*", tint)
		p.X, p.Y = bx, by
		s.Combat.Say(p.X, p.Y-30, "*", tint)
	case spells.ShapeShield:
		s.Shield = &combat.Shield{HP: sp.Shield, MaxHP: sp.Shield, Left: sp.Duration, Tint: tint}
	default:
		s.Combat.Say(p.X, p.Y-36, "fizzle", ui_dim)
	}
}

var ui_dim = color.RGBA{200, 190, 180, 255}

func dist(ax, ay, bx, by float64) float64 {
	return math.Hypot(ax-bx, ay-by)
}

// Draw renders ground, the circle, then props and characters interleaved
// by row, then effects, the overhead layer, and the interface.
func (s *Campus) Draw(ctx *engine.Context, canvas *ebiten.Image) {
	ox, oy := ctx.Camera.Offset()
	v := ctx.Camera.Visible()
	s.Map.DrawLayerRows(canvas, layerGround, ox, oy, v.X, v.Y, v.W, v.H, 0, s.Map.Height-1)
	s.drawCircle(canvas, ox, oy, ctx.Tick)

	chars := make([]drawable, 0, len(s.NPCs)+len(s.Combat.Dummies)+1)
	chars = append(chars, drawable{s.Player.Y, func() {
		s.Player.Draw(canvas, ox, oy)
		combat.DrawShield(canvas, s.Shield, s.Player.X+ox, s.Player.Y+oy)
	}})
	for _, n := range s.NPCs {
		if n.Sprite != nil {
			nn := n
			chars = append(chars, drawable{n.Y, func() { nn.Draw(canvas, ox, oy) }})
		}
	}
	for _, d := range s.Combat.Dummies {
		dd := d
		chars = append(chars, drawable{d.Y, func() { dd.Draw(canvas, ox, oy) }})
	}
	sortDrawables(chars)
	row := s.Map.RowOf(v.Y)
	for _, c := range chars {
		cr := s.Map.RowOf(c.y - 1)
		if cr >= row {
			s.Map.DrawLayerRows(canvas, layerProps, ox, oy, v.X, v.Y, v.W, v.H, row, cr)
			row = cr + 1
		}
		c.draw()
	}
	s.Map.DrawLayerRows(canvas, layerProps, ox, oy, v.X, v.Y, v.W, v.H, row, s.Map.Height-1)
	s.Combat.DrawEffects(canvas, ox, oy)
	s.Map.DrawLayerRows(canvas, layerOverhead, ox, oy, v.X, v.Y, v.W, v.H, 0, s.Map.Height-1)

	uiOpen := s.Talk.Active() || s.Panel.Open || s.Circle.Open
	if !uiOpen {
		if s.inCircle {
			hud.Prompt(canvas, "E: Open the circle", s.Player.X+ox, s.Player.Y+oy-8)
		} else if s.nearest != nil {
			label := "E: Talk"
			if s.nearest.Sprite == nil {
				label = "E: Read"
			}
			hud.Prompt(canvas, label, s.nearest.X+ox, s.nearest.Y+oy)
		}
	}
	if !s.Circle.Open {
		hud.Status(canvas, s.Player.HP, s.Player.MaxHP, s.Player.Mana, s.Player.MaxMana, s.Book)
	}
	s.Talk.Draw(canvas)
	s.Panel.Draw(canvas, &s.Inventory)
	s.Circle.Draw(canvas, &s.Inventory)

	if s.Debug {
		ebitenutil.DebugPrint(canvas, fmt.Sprintf("\n\n\ntps %.0f fps %.0f\npos %.0f,%.0f tile %d,%d",
			ebiten.ActualTPS(), ebiten.ActualFPS(), s.Player.X, s.Player.Y, int(s.Player.X)/16, int(s.Player.Y)/16))
	}
}

// drawCircle paints the 12-point star on the ground; it glows when the
// player stands in it.
func (s *Campus) drawCircle(canvas *ebiten.Image, ox, oy float64, tick uint64) {
	cx, cy := float32(s.circleX+ox), float32(s.circleY+oy)
	a := uint8(110)
	if s.inCircle {
		a = uint8(170 + 60*math.Sin(float64(tick)/8))
	}
	ink := color.RGBA{230, 200, 120, a}
	dim := color.RGBA{230, 200, 120, a / 2}
	r := float32(circleRadius)
	vector.StrokeCircle(canvas, cx, cy, r, 1, ink, false)
	vector.StrokeCircle(canvas, cx, cy, r*0.35, 1, dim, false)
	pt := func(i int) (float32, float32) {
		ang := -math.Pi/2 + float64(i)*2*math.Pi/hud.Ring
		return cx + float32(math.Cos(ang))*r, cy + float32(math.Sin(ang))*r
	}
	for i := 0; i < hud.Ring; i++ {
		x0, y0 := pt(i)
		x1, y1 := pt((i + 5) % hud.Ring)
		vector.StrokeLine(canvas, x0, y0, x1, y1, 1, dim, false)
		vector.DrawFilledCircle(canvas, x0, y0, 1.5, ink, false)
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
