// Package grammar defines reusable typed valency frames and construction layouts.
// Neither stores sentences or makes lexical permissions by surface spelling.
package grammar

import (
	"fmt"

	"github.com/clickbytes/golingua2/internal/core"
)

type Slot string

const (
	Demonstrative Slot = "demonstrative"
	Copula        Slot = "copula"
	Article       Slot = "article"
	Noun          Slot = "noun"
	Adjective     Slot = "adjective"
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
	ID         string
	FrameID    string
	Predicate  core.ID
	Sequence   []Element
	Analyze    bool
	Realize    bool
	DemoGender string // "fixed_neuter", "agree", "invariant"
}

func Validate(p Pattern) error {
	if p.ID == "" || p.FrameID == "" || len(p.Sequence) < 3 || len(p.Sequence) > 12 || (!p.Analyze && !p.Realize) {
		return fmt.Errorf("invalid construction declaration %q", p.ID)
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
