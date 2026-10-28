// Package spells turns keywords into a castable spell. Items carry keywords;
// binding puts a set of items on a scroll; this package resolves the
// resulting keyword bag into properties and a name, Isaac-style: every
// keyword contributes to shared properties, and a short synergy table adds
// hand-written names and tweaks for the combinations worth celebrating.
package spells

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Shape is how a spell is delivered.
type Shape string

const (
	ShapeNone       Shape = ""
	ShapeProjectile Shape = "projectile"
	ShapeSelf       Shape = "self"
	ShapeBlink      Shape = "blink"
	ShapeShield     Shape = "shield"
)

// Keyword is one entry of the keyword table.
type Keyword struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Order sorts keywords in the spell's name; lower comes first.
	Order    int     `json:"order"`
	Element  string  `json:"element"`
	Shape    Shape   `json:"shape"`
	Damage   float64 `json:"damage"`
	Heal     float64 `json:"heal"`
	Shield   float64 `json:"shield"`
	Count    int     `json:"count"`
	Speed    float64 `json:"speed"`
	Size     float64 `json:"size"`
	Range    float64 `json:"range"`
	Duration float64 `json:"duration"`
	Mana     float64 `json:"mana"`
	ManaMult float64 `json:"manaMult"`
	Tint     []int   `json:"tint"`
}

// Synergy is a hand-written combination.
type Synergy struct {
	Keywords   []string `json:"keywords"`
	Name       string   `json:"name"`
	DamageMult float64  `json:"damageMult"`
	Note       string   `json:"note"`
}

// Table is the loaded keyword and synergy data.
type Table struct {
	keywords  map[string]*Keyword
	synergies []Synergy
}

// ParseTable decodes spells.json.
func ParseTable(data []byte) (*Table, error) {
	var file struct {
		Keywords  []*Keyword `json:"keywords"`
		Synergies []Synergy  `json:"synergies"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("spells.json: %w", err)
	}
	t := &Table{keywords: map[string]*Keyword{}}
	for _, k := range file.Keywords {
		if k.ID == "" {
			return nil, fmt.Errorf("spells.json: keyword with empty id")
		}
		if k.Label == "" {
			k.Label = strings.ToUpper(k.ID[:1]) + k.ID[1:]
		}
		t.keywords[k.ID] = k
	}
	for _, s := range file.Synergies {
		for _, id := range s.Keywords {
			if _, ok := t.keywords[id]; !ok {
				return nil, fmt.Errorf("spells.json: synergy %q uses unknown keyword %q", s.Name, id)
			}
		}
		t.synergies = append(t.synergies, s)
	}
	// Longest synergies first, so the most specific match wins.
	sort.SliceStable(t.synergies, func(i, j int) bool {
		return len(t.synergies[i].Keywords) > len(t.synergies[j].Keywords)
	})
	return t, nil
}

// Keyword looks up one keyword.
func (t *Table) Keyword(id string) (*Keyword, bool) {
	k, ok := t.keywords[id]
	return k, ok
}

// Spell is a resolved, castable thing.
type Spell struct {
	Name     string
	Keywords []string // as given, in binding order
	Shape    Shape
	Element  string
	Damage   float64
	Heal     float64
	Shield   float64
	Count    int
	Speed    float64
	Size     float64
	Range    float64
	Duration float64
	Mana     float64
	Tint     [3]uint8
	// Synergy names the matched synergy, if any.
	Synergy string
}

// Resolve folds a bag of keywords into a Spell. Unknown keywords are
// ignored (they appear in the name as themselves, so a typo is visible).
func (t *Table) Resolve(ids []string) Spell {
	sp := Spell{Keywords: append([]string(nil), ids...), Tint: [3]uint8{230, 230, 240}}
	manaMult := 1.0
	var tintSum [3]float64
	tints := 0
	for _, id := range ids {
		k, ok := t.keywords[id]
		if !ok {
			continue
		}
		if k.Element != "" && sp.Element == "" {
			sp.Element = k.Element
		}
		if k.Shape != "" {
			sp.Shape = k.Shape
		}
		sp.Damage += k.Damage
		sp.Heal += k.Heal
		sp.Shield += k.Shield
		sp.Count += k.Count
		sp.Speed = math.Max(sp.Speed, k.Speed)
		sp.Size += k.Size
		sp.Range = math.Max(sp.Range, k.Range)
		sp.Duration += k.Duration
		sp.Mana += k.Mana
		if k.ManaMult != 0 {
			manaMult *= k.ManaMult
		}
		if len(k.Tint) == 3 {
			for i := range tintSum {
				tintSum[i] += float64(k.Tint[i])
			}
			tints++
		}
	}
	if tints > 0 {
		for i := range sp.Tint {
			sp.Tint[i] = uint8(tintSum[i] / float64(tints))
		}
	}
	sp.Mana = math.Max(1, math.Round(sp.Mana*manaMult))
	sp.Name = t.defaultName(ids)
	if s := t.matchSynergy(ids); s != nil {
		sp.Synergy = s.Name
		if s.Name != "" {
			sp.Name = s.Name
		}
		if s.DamageMult != 0 {
			sp.Damage *= s.DamageMult
		}
	}
	return sp
}

// defaultName joins labels in Order, dropping exact duplicates so two fire
// keywords read "Fire Ball" with the duplicate shown as a count.
func (t *Table) defaultName(ids []string) string {
	type part struct {
		order int
		label string
		n     int
	}
	var parts []*part
	index := map[string]*part{}
	for _, id := range ids {
		if p, ok := index[id]; ok {
			p.n++
			continue
		}
		p := &part{order: 50, label: id, n: 1}
		if k, ok := t.keywords[id]; ok {
			p.order, p.label = k.Order, k.Label
		}
		index[id] = p
		parts = append(parts, p)
	}
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].order < parts[j].order })
	var words []string
	for _, p := range parts {
		w := p.label
		if p.n > 1 {
			w = fmt.Sprintf("%s x%d", w, p.n)
		}
		words = append(words, w)
	}
	return strings.Join(words, " ")
}

// matchSynergy finds the most specific synergy whose keywords (with
// multiplicity) are all present in the bag.
func (t *Table) matchSynergy(ids []string) *Synergy {
	have := map[string]int{}
	for _, id := range ids {
		have[id]++
	}
	for i := range t.synergies {
		s := &t.synergies[i]
		need := map[string]int{}
		for _, id := range s.Keywords {
			need[id]++
		}
		ok := true
		for id, n := range need {
			if have[id] < n {
				ok = false
				break
			}
		}
		if ok {
			return s
		}
	}
	return nil
}
