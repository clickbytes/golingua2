package compiler_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

func TestCompilerRejectsUnlicensedData(t *testing.T) {
	catalog, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*compile.Definition)
	}{
		{"duplicate lexeme", func(d *compile.Definition) { d.Lexemes = append(d.Lexemes, d.Lexemes[0]) }},
		{"unbound sense", func(d *compile.Definition) {
			d.Senses = append(d.Senses, lexicon.Sense{ID: "bad", LexemeID: "no.such.lexeme", Concept: "entity.cat"})
		}},
		{"wrong semantic type", func(d *compile.Definition) { d.Senses[0].Concept = "property.black" }},
		{"duplicate class", func(d *compile.Definition) { d.MorphClasses = append(d.MorphClasses, d.MorphClasses[0]) }},
		{"missing stem", func(d *compile.Definition) {
			for i := range d.Lexemes {
				if d.Lexemes[i].ID == "pl.noun.cat" {
					d.Lexemes[i].Stems = nil
				}
			}
		}},
		{"wrong morphological class", func(d *compile.Definition) { d.Lexemes[0].ClassID = "en.noun.regular-s" }},
		{"duplicate construction", func(d *compile.Definition) { d.Constructions = append(d.Constructions, d.Constructions[0]) }},
		{"unlicensed slot", func(d *compile.Definition) {
			d.Constructions[0].Sequence = append(d.Constructions[0].Sequence, grammar.Element{Slot: grammar.Adjective})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := pl.Definition()
			tc.mutate(&d)
			if _, err := compile.Compile(d, catalog); err == nil {
				t.Fatal("compiler accepted invalid package")
			}
		})
	}
}

func TestDirectionLicensesCannotBePromoted(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	d := en.Definition()
	for i := range d.Constructions {
		if d.Constructions[i].Predicate == "predicate.has_property" {
			d.Constructions[i].Realize = false
		}
	}
	p, err := compile.Compile(d, c)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range p.Constructions {
		if p.Predicate == "predicate.has_property" && p.Realize {
			t.Fatal("analyze-only pattern became target-capable")
		}
	}
}

func TestCompiledMorphologyIsDetachedFromAuthoredRules(t *testing.T) {
	catalog, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	definition := pl.Definition()
	compiled, err := compile.Compile(definition, catalog)
	if err != nil {
		t.Fatal(err)
	}
	// Mutating user-supplied data after compile must not alter a live pack.
	for c := range definition.MorphClasses {
		if definition.MorphClasses[c].ID == "pl.adjective.hard-y" {
			definition.MorphClasses[c].Rules[0].Steps[1].Arg = "wrong"
		}
	}
	form, err := compiled.Morph.Generate("pl.adjective.black", features.Spec{Number: "singular", Case: "nominative", Gender: "masculine", Degree: "positive"})
	if err != nil || form != "czarny" {
		t.Fatalf("compiled rule mutated through source data: %q %v", form, err)
	}
}
