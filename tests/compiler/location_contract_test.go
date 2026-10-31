package compiler_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/languages/en"
)

func TestM4SpatialContractsRefuseForgedLicenses(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*compile.Definition)
	}{
		{"wrong-relation", func(d *compile.Definition) { d.LocationFrames[0].Predicate = "predicate.see" }},
		{"wrong-role", func(d *compile.Definition) { d.LocationFrames[0].LocatedRole = "agent" }},
		{"missing-adposition", func(d *compile.Definition) { d.LocationFrames[0].AdpositionID = "en.adposition.missing" }},
		{"unlicensed-adposition-pos", func(d *compile.Definition) { d.LocationFrames[0].AdpositionID = "en.noun.house" }},
		{"wrong-case", func(d *compile.Definition) { d.LocationFrames[0].SiteCase = "inapplicable" }},
		{"fake-site-class", func(d *compile.Definition) { d.LocationFrames[0].SiteClass = "school-by-name" }},
		{"duplicate-location-frame", func(d *compile.Definition) { d.LocationFrames = append(d.LocationFrames, d.LocationFrames[0]) }},
		{"pattern-omits-site", func(d *compile.Definition) {
			for i := range d.Constructions {
				if d.Constructions[i].Location {
					d.Constructions[i].Sequence = d.Constructions[i].Sequence[:3]
					break
				}
			}
		}},
		{"pattern-swaps-site", func(d *compile.Definition) {
			for i := range d.Constructions {
				if d.Constructions[i].Location {
					s := d.Constructions[i].Sequence
					s[2], s[3] = s[3], s[2]
					break
				}
			}
		}},
		{"pattern-unbound-frame", func(d *compile.Definition) {
			for i := range d.Constructions {
				if d.Constructions[i].Location {
					d.Constructions[i].FrameID = "absent"
					break
				}
			}
		}},
		{"invalid-lexical-class", func(d *compile.Definition) {
			for i := range d.Lexemes {
				if d.Lexemes[i].ID == "en.noun.school" {
					d.Lexemes[i].SiteClass = "not-allowed"
					break
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := en.Definition()
			tc.change(&d)
			if _, err := compile.Compile(d, c); err == nil {
				t.Fatalf("forged location contract %q accepted", tc.name)
			}
		})
	}
	d := en.Definition()
	// Changing direction on a location pattern must not create a source license.
	for i := range d.Constructions {
		if d.Constructions[i].Location {
			d.Constructions[i].Analyze = false
		}
	}
	p, err := compile.Compile(d, c)
	if err != nil {
		t.Fatal(err)
	}
	for _, pattern := range p.Constructions {
		if pattern.Location && pattern.Analyze {
			t.Fatalf("realize-only location leaked to analysis: %s", pattern.ID)
		}
	}
	p2, err := compile.Compile(en.Definition(), c)
	if err != nil {
		t.Fatal(err)
	}
	original, ok := p2.LocationFrame("en.location.at_site.institution")
	if !ok {
		t.Fatal("missing compiled location frame")
	}
	d2 := en.Definition()
	for i := range d2.LocationFrames {
		if d2.LocationFrames[i].ID == original.ID {
			d2.LocationFrames[i].AdpositionID = "en.adposition.in"
		}
	}
	stored, ok := p2.LocationFrame(original.ID)
	if !ok || stored.AdpositionID != original.AdpositionID {
		t.Fatal("compiled location frame changed after source mutation")
	}
	// An adposition is structural lexicon, not a semantic event sense.
	d3 := en.Definition()
	d3.Senses = append(d3.Senses, lexicon.Sense{ID: "en.fake.adposition.predicate", LexemeID: "en.adposition.in", Concept: "relation.inside"})
	if _, err := compile.Compile(d3, c); err == nil {
		t.Fatal("adposition falsely licensed as semantic sense")
	}
}
