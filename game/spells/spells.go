// Package spells turns keywords into a castable spell. Items carry keywords;
// binding puts a set of items on a scroll; this package resolves the
// resulting keyword bag into stats and a name, Isaac-style: every keyword
// contributes to shared stats, and a short synergy table adds hand-written
// names and tweaks for the combinations worth celebrating.
//
// The stats follow the design notes (docs/DESIGN.md): mana cost, cast
// time, cooldown, damage, spell amp, radius, fail rate, misfire,
// continuous, free fire.
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
	ShapeProjectile Shape = "projectile" // flies from the caster
	ShapeSelf       Shape = "self"       // heals the caster
	ShapeBlink      Shape = "blink"      // moves the caster
	ShapeShield     Shape = "shield"     // bubble on the caster
	ShapeStrike     Shape = "strike"     // hits a point at range (thunder)
	ShapeBuff       Shape = "buff"       // timed stat change on the caster
)

// Keyword is one entry of the keyword table. Every numeric field adds to
// the spell's matching stat; the Mult fields multiply at the end.
type Keyword struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Order sorts keywords in the spell's name; lower comes first.
	Order   int    `json:"order"`
	Element string `json:"element"`
	Shape   Shape  `json:"shape"`

	Damage   float64 `json:"damage"`
	Heal     float64 `json:"heal"`
	Shield   float64 `json:"shield"`
	Count    int     `json:"count"`
	Speed    float64 `json:"speed"`
	Size     float64 `json:"size"`
	Range    float64 `json:"range"`
	Radius   float64 `json:"radius"`
	Duration float64 `json:"duration"`
	Haste    float64 `json:"haste"` // +move speed fraction while buffed
	Dodge    float64 `json:"dodge"` // dodge chance while buffed

	Mana         float64 `json:"mana"`
	ManaMult     float64 `json:"manaMult"`
	CastTime     float64 `json:"castTime"`
	Cooldown     float64 `json:"cooldown"`
	CooldownMult float64 `json:"cooldownMult"`
	Amp          float64 `json:"amp"`     // spell amp, additive fraction
	Fail         float64 `json:"fail"`    // fail rate, additive
	Misfire      float64 `json:"misfire"` // misfire chance, additive
	Continuous   bool    `json:"continuous"`
	FreeFire     bool    `json:"freeFire"`

	Tint []int `json:"tint"`
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

	// The ten stats from the notes.
	Mana       float64
	CastTime   float64
	Cooldown   float64
	Damage     float64 // after amp and synergy
	Amp        float64
	Radius     float64
	Fail       float64
	Misfire    float64
	Continuous bool
	FreeFire   bool

	// Shape-specific.
	Heal     float64
	Shield   float64
	Count    int
	Speed    float64
	Size     float64
	Range    float64
	Duration float64
	Haste    float64
	Dodge    float64

	Tint [3]uint8
	// Synergy names the matched synergy, if any.
	Synergy string
}

// Resolve folds a bag of keywords into a Spell. Unknown keywords are
// ignored (they appear in the name as themselves, so a typo is visible).
func (t *Table) Resolve(ids []string) Spell {
	sp := Spell{Keywords: append([]string(nil), ids...), Tint: [3]uint8{230, 230, 240}}
	manaMult, cdMult := 1.0, 1.0
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
		sp.Radius += k.Radius
		sp.Duration += k.Duration
		sp.Haste += k.Haste
		sp.Dodge += k.Dodge
		sp.Mana += k.Mana
		sp.CastTime += k.CastTime
		sp.Cooldown += k.Cooldown
		sp.Amp += k.Amp
		sp.Fail += k.Fail
		sp.Misfire += k.Misfire
		sp.Continuous = sp.Continuous || k.Continuous
		sp.FreeFire = sp.FreeFire || k.FreeFire
		if k.ManaMult != 0 {
			manaMult *= k.ManaMult
		}
		if k.CooldownMult != 0 {
			cdMult *= k.CooldownMult
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
	sp.Cooldown = math.Max(0, sp.Cooldown*cdMult)
	sp.Fail = clamp01(sp.Fail)
	sp.Misfire = clamp01(sp.Misfire)
	sp.Damage *= 1 + sp.Amp
	sp.Heal *= 1 + sp.Amp
	sp.Shield *= 1 + sp.Amp
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
	sp.Damage = math.Round(sp.Damage*10) / 10
	return sp
}

func clamp01(v float64) float64 { return math.Min(1, math.Max(0, v)) }

// defaultName joins labels in Order, showing repeated keywords as a count.
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

// Describe is a one-line summary of what the spell does, for UI.
func (sp Spell) Describe() string {
	switch sp.Shape {
	case ShapeProjectile:
		return fmt.Sprintf("%d shot, %.0f dmg", max(sp.Count, 1), sp.Damage)
	case ShapeSelf:
		return fmt.Sprintf("heals %.0f", sp.Heal)
	case ShapeBlink:
		return fmt.Sprintf("blink %.0f px", sp.Range)
	case ShapeShield:
		return fmt.Sprintf("shield %.0f for %.0fs", sp.Shield, sp.Duration)
	case ShapeStrike:
		return fmt.Sprintf("strike %.0f dmg, radius %.0f", sp.Damage, sp.Radius)
	case ShapeBuff:
		return fmt.Sprintf("+%.0f%% speed, %.0f%% dodge, %.0fs", sp.Haste*100, sp.Dodge*100, sp.Duration)
	}
	return "no shape yet"
}

// StatLines lists the ten stats for the grimoire page.
func (sp Spell) StatLines() []string {
	yn := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	return []string{
		fmt.Sprintf("mana %.0f", sp.Mana),
		fmt.Sprintf("cast %.1fs", sp.CastTime),
		fmt.Sprintf("cooldown %.1fs", sp.Cooldown),
		fmt.Sprintf("damage %.0f", sp.Damage),
		fmt.Sprintf("amp %+.0f%%", sp.Amp*100),
		fmt.Sprintf("radius %.0f", sp.Radius),
		fmt.Sprintf("fail %.0f%%", sp.Fail*100),
		fmt.Sprintf("misfire %.0f%%", sp.Misfire*100),
		"continuous " + yn(sp.Continuous),
		"free fire " + yn(sp.FreeFire),
	}
}
