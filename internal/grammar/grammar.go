// Package grammar defines reusable typed valency frames and construction layouts.
// Neither stores sentences or makes lexical permissions by surface spelling.
package grammar

import (
	"fmt"

	"github.com/clickbytes/golingua2/internal/core"
)

type Slot string

const (
	Demonstrative     Slot = "demonstrative"
	Copula            Slot = "copula"
	Article           Slot = "article"
	Noun              Slot = "noun"
	Adjective         Slot = "adjective"
	SubjectPhrase     Slot = "subject_phrase"
	PredicateVerb     Slot = "predicate_verb"
	ObjectPhrase      Slot = "object_phrase"
	NegationAuxiliary Slot = "negation_auxiliary"
	NegationMarker    Slot = "negation_marker"
)

type Frame struct {
	ID        string
	Predicate core.ID
	Role      string
	// ValueSlot supplies the predicate's typed value concept.
	ValueSlot Slot
	// An explicit entity concept if the referent is an unspecified class member.
	// Empty means its concept comes from the noun slot (property attribution).
	ReferentConcept core.ID
}

// EventFrame uses neutral role signatures and language-local government. It has
// no per-verb construction list: each verb sense explicitly licenses the frame.
type EventFrame struct {
	ID                   string
	SubjectRole          string
	ObjectRole           string // empty for an intransitive predicate
	ObjectCase           string // target language morphological case
	NegatedObjectCase    string // explicit case government for Not, empty means ObjectCase
	AllowImplicitSubject bool
}

func ValidateEventFrame(f EventFrame) error {
	if f.ID == "" || f.SubjectRole == "" || (f.ObjectRole != "" && f.ObjectRole == f.SubjectRole) || (f.ObjectRole == "" && (f.ObjectCase != "" || f.NegatedObjectCase != "")) {
		return fmt.Errorf("invalid event frame %q", f.ID)
	}
	return nil
}

// ValidateFrame compares the language frame with the neutral concept signature.
func ValidateFrame(f Frame, catalog *core.Catalog) error {
	if f.ID == "" || f.Role == "" || (f.ValueSlot != Noun && f.ValueSlot != Adjective) {
		return fmt.Errorf("invalid valency frame %q", f.ID)
	}
	def, ok := catalog.Lookup(f.Predicate)
	if !ok || def.Kind != core.Predicate {
		return fmt.Errorf("frame %q has unknown predicate", f.ID)
	}
	valueKind := core.Entity
	if f.ValueSlot == Adjective {
		valueKind = core.Property
	}
	if def.ValueKind != valueKind {
		return fmt.Errorf("frame %q has wrong predicate value sort", f.ID)
	}
	if len(def.Roles) != 1 || def.Roles[0].Name != f.Role || def.Roles[0].Target != core.Entity {
		return fmt.Errorf("frame %q has wrong semantic role signature", f.ID)
	}
	if f.ReferentConcept != "" {
		entity, ok := catalog.Lookup(f.ReferentConcept)
		if !ok || entity.Kind != core.Entity {
			return fmt.Errorf("frame %q has invalid referent concept", f.ID)
		}
	}
	if f.ValueSlot == Adjective && f.ReferentConcept != "" {
		return fmt.Errorf("attribute frame must bind the noun referent")
	}
	return nil
}

type Element struct{ Slot Slot }
type Pattern struct {
	Event      bool
	Negated    bool // this construction covers an exact Not scope
	ID         string
	FrameID    string
	Predicate  core.ID
	Sequence   []Element
	Analyze    bool
	Realize    bool
	DemoGender string // "fixed_neuter", "agree", "invariant"
}

func Validate(p Pattern) error {
	if p.ID == "" || p.FrameID == "" || len(p.Sequence) == 0 || len(p.Sequence) > 12 || (!p.Analyze && !p.Realize) {
		return fmt.Errorf("invalid construction declaration %q", p.ID)
	}
	if p.Event {
		if p.DemoGender != "" || p.Predicate != "" {
			return fmt.Errorf("event pattern mixes unrelated predicate/demonstrative policies %q", p.ID)
		}
		counts := map[Slot]int{}
		for _, el := range p.Sequence {
			counts[el.Slot]++
		}
		if counts[PredicateVerb] != 1 || counts[SubjectPhrase] > 1 || counts[ObjectPhrase] > 1 ||
			counts[NegationMarker] > 1 || counts[NegationAuxiliary] > 1 ||
			counts[SubjectPhrase]+counts[ObjectPhrase]+counts[PredicateVerb]+counts[NegationMarker]+counts[NegationAuxiliary] != len(p.Sequence) {
			return fmt.Errorf("invalid event slots %q", p.ID)
		}
		if p.Negated != (counts[NegationMarker] == 1) || (counts[NegationAuxiliary] == 1 && !p.Negated) {
			return fmt.Errorf("negation marker/auxiliary mismatch in %q", p.ID)
		}
		// The layout is shared by both languages; only language packs declare
		// which of the optional auxiliary and explicit subject are licensed.
		expected := []Slot{}
		if counts[SubjectPhrase] != 0 {
			expected = append(expected, SubjectPhrase)
		}
		if counts[NegationAuxiliary] != 0 {
			expected = append(expected, NegationAuxiliary)
		}
		if counts[NegationMarker] != 0 {
			expected = append(expected, NegationMarker)
		}
		expected = append(expected, PredicateVerb)
		if counts[ObjectPhrase] != 0 {
			expected = append(expected, ObjectPhrase)
		}
		if len(expected) != len(p.Sequence) {
			return fmt.Errorf("invalid event word order %q", p.ID)
		}
		for i, slot := range expected {
			if p.Sequence[i].Slot != slot {
				return fmt.Errorf("invalid event word order %q", p.ID)
			}
		}
		return nil
	}
	if p.Negated {
		return fmt.Errorf("negation scope requires an event construction %q", p.ID)
	}
	if p.DemoGender != "fixed_neuter" && p.DemoGender != "agree" && p.DemoGender != "invariant" {
		return fmt.Errorf("construction %q needs demonstrative policy", p.ID)
	}
	counts := map[Slot]int{}
	for _, x := range p.Sequence {
		switch x.Slot {
		case Demonstrative, Copula, Article, Noun, Adjective:
		default:
			return fmt.Errorf("unknown slot %q", x.Slot)
		}
		counts[x.Slot]++
	}
	if counts[Demonstrative] != 1 || counts[Copula] != 1 || counts[Noun] != 1 || counts[Article] > 1 || counts[Adjective] > 1 {
		return fmt.Errorf("invalid construction slots %q", p.ID)
	}
	return nil
}
