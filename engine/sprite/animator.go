package sprite

// Animator plays one Tag of a Sheet at a time, honouring per-frame durations
// and the tag's direction. It is pure data: Advance takes elapsed time, and
// Frame tells you what to draw. No Ebitengine types, so it is unit-testable.
type Animator struct {
	Sheet *Sheet

	tag      Tag
	tagName  string
	index    int // current frame index into Sheet.Frames
	elapsed  float64
	forward  bool // for pingpong
	plays    int
	finished bool
}

// NewAnimator starts an animator on the given tag (or the first frame if the
// tag is unknown).
func NewAnimator(s *Sheet, tag string) *Animator {
	a := &Animator{Sheet: s}
	a.Play(tag)
	return a
}

// Play switches to a tag from its first frame. Playing the tag that is
// already running is a no-op, so callers can set state every tick.
func (a *Animator) Play(tag string) {
	if a.tagName == tag && !a.finished {
		return
	}
	a.Restart(tag)
}

// Restart switches to a tag and rewinds even if it was already playing.
func (a *Animator) Restart(tag string) {
	a.tagName = tag
	a.elapsed = 0
	a.plays = 0
	a.finished = false
	a.forward = true
	t, ok := a.Sheet.Tags[tag]
	if !ok {
		// Unknown tag: freeze on frame 0 so a typo is visible, not a crash.
		t = Tag{Name: tag, From: 0, To: 0, Direction: Forward}
	}
	a.tag = t
	if t.Direction == Reverse {
		a.index = t.To
	} else {
		a.index = t.From
	}
}

// Tag is the name of the running animation.
func (a *Animator) Tag() string { return a.tagName }

// Finished is true once a non-looping tag has played out.
func (a *Animator) Finished() bool { return a.finished }

// Frame returns the frame to draw right now.
func (a *Animator) Frame() Frame { return a.Sheet.Frames[a.index] }

// Index returns the absolute frame index on the sheet.
func (a *Animator) Index() int { return a.index }

// Advance moves time forward by dt seconds.
func (a *Animator) Advance(dt float64) {
	if a.finished {
		return
	}
	a.elapsed += dt * 1000
	for {
		d := float64(a.Sheet.Frames[a.index].DurationMs)
		if a.elapsed < d {
			return
		}
		a.elapsed -= d
		if !a.step() {
			a.finished = true
			a.elapsed = 0
			return
		}
	}
}

// step moves to the next frame; it returns false when a finite animation
// has completed.
func (a *Animator) step() bool {
	t := a.tag
	if t.From == t.To {
		return a.completePlay()
	}
	switch t.Direction {
	case Reverse:
		if a.index > t.From {
			a.index--
			return true
		}
		if !a.completePlay() {
			return false
		}
		a.index = t.To
	case PingPong:
		if a.forward {
			if a.index < t.To {
				a.index++
				return true
			}
			a.forward = false
			a.index--
			return true
		}
		if a.index > t.From {
			a.index--
			if a.index == t.From {
				if !a.completePlay() {
					return false
				}
				a.forward = true
			}
			return true
		}
		a.forward = true
		a.index++
	default:
		if a.index < t.To {
			a.index++
			return true
		}
		if !a.completePlay() {
			return false
		}
		a.index = t.From
	}
	return true
}

// completePlay counts a finished loop and reports whether to keep going.
func (a *Animator) completePlay() bool {
	a.plays++
	if a.tag.Repeat > 0 && a.plays >= a.tag.Repeat {
		return false
	}
	return true
}
