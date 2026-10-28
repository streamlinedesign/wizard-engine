package sprite

import (
	"image"
	"testing"
)

const arrayJSON = `{
 "frames": [
  {"filename":"w 0","frame":{"x":0,"y":0,"w":16,"h":32},"spriteSourceSize":{"x":0,"y":0,"w":16,"h":32},"sourceSize":{"w":16,"h":32},"duration":100},
  {"filename":"w 1","frame":{"x":16,"y":0,"w":16,"h":32},"spriteSourceSize":{"x":0,"y":0,"w":16,"h":32},"sourceSize":{"w":16,"h":32},"duration":150},
  {"filename":"w 2","frame":{"x":32,"y":0,"w":12,"h":30},"spriteSourceSize":{"x":2,"y":2,"w":12,"h":30},"sourceSize":{"w":16,"h":32},"duration":100}
 ],
 "meta": {"image":"w.png","frameTags":[
  {"name":"walk","from":0,"to":2,"direction":"forward"},
  {"name":"bounce","from":0,"to":2,"direction":"pingpong"},
  {"name":"once","from":0,"to":1,"direction":"forward","repeat":"1"}
 ]}
}`

const hashJSON = `{
 "frames": {
  "w 10.ase": {"frame":{"x":32,"y":0,"w":16,"h":32},"duration":100},
  "w 2.ase":  {"frame":{"x":16,"y":0,"w":16,"h":32},"duration":100},
  "w 1.ase":  {"frame":{"x":0,"y":0,"w":16,"h":32},"duration":100}
 },
 "meta": {"image":"w.png","frameTags":[]}
}`

func TestParseArray(t *testing.T) {
	s, err := ParseAseprite([]byte(arrayJSON))
	if err != nil {
		t.Fatal(err)
	}
	if s.Image != "w.png" || len(s.Frames) != 3 {
		t.Fatalf("bad sheet: %+v", s)
	}
	if got := s.Frames[2]; got.Rect != image.Rect(32, 0, 44, 30) || got.Offset != image.Pt(2, 2) || got.Size != image.Pt(16, 32) {
		t.Fatalf("trimmed frame parsed wrong: %+v", got)
	}
	if tag := s.Tags["once"]; tag.Repeat != 1 {
		t.Fatalf("repeat not parsed: %+v", tag)
	}
}

func TestParseHashSortsNaturally(t *testing.T) {
	s, err := ParseAseprite([]byte(hashJSON))
	if err != nil {
		t.Fatal(err)
	}
	want := []int{0, 16, 32}
	for i, fr := range s.Frames {
		if fr.Rect.Min.X != want[i] {
			t.Fatalf("frame %d at x=%d, want %d", i, fr.Rect.Min.X, want[i])
		}
	}
}

func TestParseRejectsBadTag(t *testing.T) {
	bad := `{"frames":[{"frame":{"x":0,"y":0,"w":1,"h":1},"duration":1}],"meta":{"frameTags":[{"name":"x","from":0,"to":5}]}}`
	if _, err := ParseAseprite([]byte(bad)); err == nil {
		t.Fatal("expected error for out-of-range tag")
	}
}

func frames(a *Animator, steps int, dt float64) []int {
	var out []int
	for i := 0; i < steps; i++ {
		out = append(out, a.Index())
		a.Advance(dt)
	}
	return out
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAnimatorForwardHonoursDurations(t *testing.T) {
	s, _ := ParseAseprite([]byte(arrayJSON))
	a := NewAnimator(s, "walk")
	// Frame 1 lasts 150ms, so at 100ms ticks it holds for two ticks the
	// first time round; the 50ms remainder carries over, so the second lap
	// runs on time instead of drifting.
	got := frames(a, 7, 0.1)
	want := []int{0, 1, 1, 2, 0, 1, 2}
	if !equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAnimatorPingPong(t *testing.T) {
	s, _ := ParseAseprite([]byte(arrayJSON))
	for i := range s.Frames {
		s.Frames[i].DurationMs = 100
	}
	a := NewAnimator(s, "bounce")
	got := frames(a, 8, 0.1)
	want := []int{0, 1, 2, 1, 0, 1, 2, 1}
	if !equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAnimatorOnceFinishes(t *testing.T) {
	s, _ := ParseAseprite([]byte(arrayJSON))
	a := NewAnimator(s, "once")
	a.Advance(0.1) // -> 1
	a.Advance(0.2) // past the end
	if !a.Finished() || a.Index() != 1 {
		t.Fatalf("expected to finish on last frame, finished=%v index=%d", a.Finished(), a.Index())
	}
	a.Play("once") // Play on a finished tag restarts it
	if a.Finished() || a.Index() != 0 {
		t.Fatal("Play should restart a finished animation")
	}
}

func TestAnimatorUnknownTagFreezes(t *testing.T) {
	s, _ := ParseAseprite([]byte(arrayJSON))
	a := NewAnimator(s, "nope")
	a.Advance(5)
	if a.Index() != 0 {
		t.Fatalf("unknown tag should hold frame 0, got %d", a.Index())
	}
}

func TestAnimatorPlayIsIdempotent(t *testing.T) {
	s, _ := ParseAseprite([]byte(arrayJSON))
	a := NewAnimator(s, "walk")
	a.Advance(0.1)
	a.Play("walk")
	if a.Index() != 1 {
		t.Fatal("Play on the running tag must not rewind")
	}
}
