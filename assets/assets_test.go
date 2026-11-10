package assets

import (
	"testing"

	"github.com/streamlinedesign/wizard-engine/engine/sprite"
)

// The art contract from docs/ASEPRITE.md, checked against the files that
// ship: every tag the game looks up exists and every frame has the cell size.
func TestArtContract(t *testing.T) {
	cases := []struct {
		json string
		w, h int
		tags []string
	}{
		{"sprites/wizard.json", 16, 40, []string{
			"idle_down", "idle_up", "idle_side", "walk_down", "walk_up", "walk_side",
			"cast_down", "cast_up", "cast_side", "hurt", "meditate"}},
		{"effects/fireball.json", 32, 32, []string{"fly", "hit"}},
		{"effects/heal.json", 32, 32, []string{"self"}},
		{"effects/thunder.json", 32, 32, []string{"strike"}},
		{"effects/gust.json", 32, 32, []string{"self"}},
		{"effects/barrier.json", 48, 48, []string{"appear", "hold", "break"}},
		{"effects/teleport.json", 32, 32, []string{"out", "in"}},
		{"effects/hit.json", 32, 32, []string{"hit"}},
		{"effects/puff.json", 16, 16, []string{"puff"}},
	}
	for _, c := range cases {
		data, err := Data(c.json)
		if err != nil {
			t.Fatal(err)
		}
		sheet, err := sprite.ParseAseprite(data)
		if err != nil {
			t.Fatalf("%s: %v", c.json, err)
		}
		if _, err := FS.ReadFile(pathNext(c.json, sheet.Image)); err != nil {
			t.Errorf("%s: image %q: %v", c.json, sheet.Image, err)
		}
		for _, tag := range c.tags {
			if _, ok := sheet.Tags[tag]; !ok {
				t.Errorf("%s: missing tag %q", c.json, tag)
			}
		}
		for i, f := range sheet.Frames {
			if f.Size.X != c.w || f.Size.Y != c.h {
				t.Errorf("%s frame %d: %dx%d, want %dx%d", c.json, i, f.Size.X, f.Size.Y, c.w, c.h)
			}
		}
	}
}
