// Package sprite loads Aseprite sprite-sheet exports and plays their
// animations. The contract with the artist is the JSON file Aseprite writes
// next to the PNG; see docs/ASEPRITE.md for the export settings.
package sprite

import (
	"encoding/json"
	"fmt"
	"image"
	"sort"
)

// Frame is one cell of the sheet.
type Frame struct {
	// Rect is where the frame lives on the sheet, in pixels.
	Rect image.Rectangle
	// Offset is where the trimmed frame sits inside the untrimmed source
	// cell, so trimmed exports still line up. Zero when not trimmed.
	Offset image.Point
	// Size is the untrimmed source cell size.
	Size image.Point
	// DurationMs is how long Aseprite says to show this frame.
	DurationMs int
}

// Direction is Aseprite's tag playback direction.
type Direction string

const (
	Forward  Direction = "forward"
	Reverse  Direction = "reverse"
	PingPong Direction = "pingpong"
)

// Tag is a named frame range (an animation) from Aseprite's timeline.
type Tag struct {
	Name      string
	From, To  int
	Direction Direction
	// Repeat is 0 for "loop forever", otherwise the number of plays.
	Repeat int
}

// Sheet is the parsed description of one exported sprite sheet.
type Sheet struct {
	// Image is the PNG filename Aseprite recorded, relative to the JSON.
	Image  string
	Frames []Frame
	Tags   map[string]Tag
}

// Aseprite writes "frames" either as an array (--format json-array) or as an
// object keyed by frame filename (json-hash). We accept both.
type asepriteFile struct {
	Frames json.RawMessage `json:"frames"`
	Meta   struct {
		Image     string `json:"image"`
		FrameTags []struct {
			Name      string `json:"name"`
			From      int    `json:"from"`
			To        int    `json:"to"`
			Direction string `json:"direction"`
			Repeat    string `json:"repeat"`
		} `json:"frameTags"`
	} `json:"meta"`
}

type asepriteFrame struct {
	Filename string `json:"filename"`
	Frame    struct {
		X, Y, W, H int
	} `json:"frame"`
	SpriteSourceSize struct {
		X, Y, W, H int
	} `json:"spriteSourceSize"`
	SourceSize struct {
		W, H int
	} `json:"sourceSize"`
	Duration int `json:"duration"`
}

// ParseAseprite decodes an Aseprite JSON export.
func ParseAseprite(data []byte) (*Sheet, error) {
	var f asepriteFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("aseprite json: %w", err)
	}

	var raw []asepriteFrame
	if len(f.Frames) > 0 && f.Frames[0] == '[' {
		if err := json.Unmarshal(f.Frames, &raw); err != nil {
			return nil, fmt.Errorf("aseprite frames array: %w", err)
		}
	} else {
		var m map[string]asepriteFrame
		if err := json.Unmarshal(f.Frames, &m); err != nil {
			return nil, fmt.Errorf("aseprite frames hash: %w", err)
		}
		// Hash form loses ordering, so sort by the filename Aseprite wrote,
		// which carries the frame number with {frame} in the default template.
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return naturalLess(keys[i], keys[j]) })
		for _, k := range keys {
			fr := m[k]
			fr.Filename = k
			raw = append(raw, fr)
		}
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("aseprite json: no frames")
	}

	s := &Sheet{Image: f.Meta.Image, Tags: map[string]Tag{}}
	for _, r := range raw {
		fr := Frame{
			Rect:       image.Rect(r.Frame.X, r.Frame.Y, r.Frame.X+r.Frame.W, r.Frame.Y+r.Frame.H),
			Offset:     image.Pt(r.SpriteSourceSize.X, r.SpriteSourceSize.Y),
			Size:       image.Pt(r.SourceSize.W, r.SourceSize.H),
			DurationMs: r.Duration,
		}
		if fr.Size == (image.Point{}) {
			fr.Size = image.Pt(r.Frame.W, r.Frame.H)
		}
		if fr.DurationMs <= 0 {
			fr.DurationMs = 100
		}
		s.Frames = append(s.Frames, fr)
	}

	for _, t := range f.Meta.FrameTags {
		if t.From < 0 || t.To >= len(s.Frames) || t.From > t.To {
			return nil, fmt.Errorf("aseprite tag %q has range %d-%d outside %d frames", t.Name, t.From, t.To, len(s.Frames))
		}
		dir := Direction(t.Direction)
		if dir == "" {
			dir = Forward
		}
		repeat := 0
		if t.Repeat != "" {
			fmt.Sscanf(t.Repeat, "%d", &repeat)
		}
		s.Tags[t.Name] = Tag{Name: t.Name, From: t.From, To: t.To, Direction: dir, Repeat: repeat}
	}
	return s, nil
}

// naturalLess compares strings so that "wizard 2.ase" < "wizard 10.ase".
func naturalLess(a, b string) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ca, cb := a[i], b[j]
		if isDigit(ca) && isDigit(cb) {
			si := i
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			sj := j
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			na, nb := trimZeros(a[si:i]), trimZeros(b[sj:j])
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			continue
		}
		if ca != cb {
			return ca < cb
		}
		i++
		j++
	}
	return len(a)-i < len(b)-j
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func trimZeros(s string) string {
	for len(s) > 1 && s[0] == '0' {
		s = s[1:]
	}
	return s
}
