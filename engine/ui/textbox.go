package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// TextBox shows one page of text with a typewriter reveal. The caller
// decides when to advance; the box only knows whether it has finished
// revealing the current page.
type TextBox struct {
	Theme Theme
	// CharsPerSecond is the reveal speed. 0 shows everything at once.
	CharsPerSecond float64

	speaker string
	lines   []string
	total   int
	shown   float64
}

// NewTextBox makes a box with the parchment theme and a brisk reveal.
func NewTextBox() *TextBox {
	return &TextBox{Theme: Parchment, CharsPerSecond: 60}
}

// Set loads a page of text, wrapped to width pixels.
func (t *TextBox) Set(speaker, body string, width float64) {
	t.speaker = speaker
	t.lines = Wrap(body, width)
	t.total = 0
	for _, l := range t.lines {
		t.total += len([]rune(l))
	}
	t.shown = 0
	if t.CharsPerSecond == 0 {
		t.shown = float64(t.total)
	}
}

// Advance reveals more characters.
func (t *TextBox) Advance(dt float64) {
	t.shown += t.CharsPerSecond * dt
	if t.shown > float64(t.total) {
		t.shown = float64(t.total)
	}
}

// Done reports whether the whole page is visible.
func (t *TextBox) Done() bool { return int(t.shown) >= t.total }

// Finish reveals the whole page immediately.
func (t *TextBox) Finish() { t.shown = float64(t.total) }

// Draw renders the panel at (x, y) with the given size, the speaker name
// on the first line and the revealed text below it.
func (t *TextBox) Draw(dst *ebiten.Image, x, y, w, h float64) {
	Panel(dst, x, y, w, h, t.Theme)
	const pad = 6
	cy := y + pad - 2
	if t.speaker != "" {
		Print(dst, t.speaker, x+pad, cy, t.Theme.Accent)
		cy += LineHeight
	}
	left := int(t.shown)
	for _, l := range t.lines {
		if left <= 0 {
			break
		}
		r := []rune(l)
		if len(r) > left {
			r = r[:left]
		}
		Print(dst, string(r), x+pad, cy, t.Theme.Text)
		left -= len(r)
		cy += LineHeight
	}
	if t.Done() {
		Print(dst, "v", x+w-pad-6, y+h-LineHeight-2, t.Theme.TextDim)
	}
}
