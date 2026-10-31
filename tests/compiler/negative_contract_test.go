package compiler_test

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
	"testing"
)

func TestM3NegativePatternLicensesCannotBeForged(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*compile.Definition)
	}{
		{"no-marker", func(d *compile.Definition) { d.Specials.NegationMarker = "missing" }},
		{"missing-aux", func(d *compile.Definition) { d.Specials.NegationAuxiliary = "" }},
		{"swapped-aux-marker", func(d *compile.Definition) {
			for i := range d.Constructions {
				p := &d.Constructions[i]
				if !p.Negated {
					continue
				}
				p.Sequence[1], p.Sequence[2] = p.Sequence[2], p.Sequence[1]
				break
			}
		}},
		{"no-operator-slot", func(d *compile.Definition) {
			for i := range d.Constructions {
				p := &d.Constructions[i]
				if !p.Negated {
					continue
				}
				p.Negated = false
				break
			}
		}},
		{"aux-sense-forged", func(d *compile.Definition) {
			d.Senses = append(d.Senses, lexicon.Sense{ID: "fake.aux", LexemeID: d.Specials.NegationAuxiliary, Concept: "predicate.sleep", FrameIDs: []string{"en.frame.intr"}})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := en.Definition()
			tc.change(&d)
			if _, err := compile.Compile(d, c); err == nil {
				t.Fatalf("accepted false licensing %s", tc.name)
			}
		})
	}
	d := pl.Definition()
	for i := range d.Constructions {
		if d.Constructions[i].Negated {
			d.Constructions[i].Analyze = false
		}
	}
	p, err := compile.Compile(d, c)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range p.Constructions {
		if x.Negated && x.Analyze {
			t.Fatal("realize-only negative pattern leaked into source")
		}
	}
	d = en.Definition()
	d.Constructions = append(d.Constructions, grammar.Pattern{ID: "bad.neg.classification", FrameID: "en.frame.classification", Predicate: "predicate.is_a", Negated: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.Demonstrative}}})
	if _, err := compile.Compile(d, c); err == nil {
		t.Fatal("state non-event Not was silently accepted")
	}
}
