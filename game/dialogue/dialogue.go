// Package dialogue holds conversations as data and runs one at a time.
// A dialogue is a speaker, a list of pages, optional items to give when
// it finishes, and an optional "then" that replaces it for next time.
package dialogue

import (
	"encoding/json"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/engine/input"
	"github.com/streamlinedesign/wizard-engine/engine/ui"
	"github.com/streamlinedesign/wizard-engine/game/items"
)

// Gift is an item handed over when a dialogue ends.
type Gift struct {
	Item  string `json:"item"`
	Count int    `json:"count"`
}

// Def is one conversation.
type Def struct {
	Speaker string   `json:"speaker"`
	Pages   []string `json:"pages"`
	Give    []Gift   `json:"give"`
	// Then names the dialogue the speaker uses after this one has played.
	Then string `json:"then"`
}

// Registry is every dialogue by id.
type Registry struct {
	defs map[string]*Def
}

// ParseRegistry decodes dialogue.json: a top-level "dialogues" object.
func ParseRegistry(data []byte) (*Registry, error) {
	var file struct {
		Dialogues map[string]*Def `json:"dialogues"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("dialogue.json: %w", err)
	}
	for id, d := range file.Dialogues {
		if len(d.Pages) == 0 {
			return nil, fmt.Errorf("dialogue.json: %q has no pages", id)
		}
		if d.Then != "" {
			if _, ok := file.Dialogues[d.Then]; !ok {
				return nil, fmt.Errorf("dialogue.json: %q then -> unknown %q", id, d.Then)
			}
		}
	}
	return &Registry{defs: file.Dialogues}, nil
}

// Get looks up a dialogue.
func (r *Registry) Get(id string) (*Def, bool) {
	d, ok := r.defs[id]
	return d, ok
}

// Runner plays one dialogue through a TextBox. Scenes own one Runner and
// route input to it while Active.
type Runner struct {
	Box *ui.TextBox

	def  *Def
	page int
	// OnGive is called for each gift when the last page is dismissed.
	OnGive func(g Gift)
	// OnEnd is called with the dialogue that just finished.
	OnEnd func(d *Def)

	width float64
}

// NewRunner makes an idle runner whose box wraps text to width pixels.
func NewRunner(width float64) *Runner {
	return &Runner{Box: ui.NewTextBox(), width: width}
}

// Active reports whether a dialogue is showing.
func (r *Runner) Active() bool { return r.def != nil }

// Start begins a dialogue from its first page.
func (r *Runner) Start(d *Def) {
	r.def = d
	r.page = 0
	r.Box.Set(d.Speaker, d.Pages[0], r.width)
}

// Update advances the reveal and routes the confirm key to Confirm.
func (r *Runner) Update(ctx *engine.Context) {
	if r.def == nil {
		return
	}
	r.Box.Advance(ctx.DT)
	if ctx.Input.JustPressed(input.Interact) || ctx.Input.JustPressed(input.Cast) {
		r.Confirm()
	}
}

// Confirm is one press of the confirm key: the first finishes the page's
// reveal, the next turns the page, and the one after the last page ends
// the dialogue, handing out gifts.
func (r *Runner) Confirm() {
	if r.def == nil {
		return
	}
	if !r.Box.Done() {
		r.Box.Finish()
		return
	}
	r.page++
	if r.page < len(r.def.Pages) {
		r.Box.Set(r.def.Speaker, r.def.Pages[r.page], r.width)
		return
	}
	d := r.def
	r.def = nil
	if r.OnGive != nil {
		for _, g := range d.Give {
			r.OnGive(g)
		}
	}
	if r.OnEnd != nil {
		r.OnEnd(d)
	}
}

// Draw renders the box along the bottom of the canvas.
func (r *Runner) Draw(canvas *ebiten.Image) {
	if r.def == nil {
		return
	}
	b := canvas.Bounds()
	const h = 56
	r.Box.Draw(canvas, 4, float64(b.Dy()-h-4), float64(b.Dx()-8), h)
}

// GiftsTo applies a dialogue's gifts to an inventory; used as OnGive.
func GiftsTo(reg *items.Registry, inv *items.Inventory) func(Gift) {
	return func(g Gift) {
		if d, ok := reg.Get(g.Item); ok {
			inv.Add(d, g.Count)
		}
	}
}
