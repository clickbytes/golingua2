package en

import (
	"fmt"
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

func addQuantity(d compile.Definition) compile.Definition {
	d.MorphClasses = append(d.MorphClasses, morph.Class{ID: "en.numeral.invariant", Language: "en", POS: lexicon.Numeral, Rules: []morph.Rule{lemmaCell(features.Spec{})}})
	for value, spelling := range []string{"", "one", "two", "three", "four", "five"} {
		if value == 0 {
			continue
		}
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: fmt.Sprintf("en.numeral.%d", value), Language: "en", POS: lexicon.Numeral, Lemma: spelling, ClassID: "en.numeral.invariant", NumeralValue: value})
	}
	// Fully regular plural forms are produced by en.noun.regular-s.
	d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: "en.noun.mug", Language: "en", POS: lexicon.Noun, Lemma: "mug", ClassID: "en.noun.regular-s", Onset: "consonant", Countability: "count"})
	d.Senses = append(d.Senses, lexicon.Sense{ID: "en.sense.noun.mug", LexemeID: "en.noun.mug", Concept: core.ID("entity.mug")})
	for i := range d.MorphClasses {
		if d.MorphClasses[i].ID == "en.copula.present" {
			d.MorphClasses[i].Rules = append(d.MorphClasses[i].Rules, morph.Rule{Features: features.Spec{Number: "plural", Person: "third", Tense: "present"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "present_3pl"}}})
		}
	}
	for i := range d.Lexemes {
		if d.Lexemes[i].ID == "en.verb.be" {
			d.Lexemes[i].Stems["present_3pl"] = "are"
		}
	}
	d.Specials.ExistentialMarker = "en.particle.there"
	d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: d.Specials.ExistentialMarker, Lemma: "there", Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant"})
	for _, f := range d.LocationFrames {
		if f.SiteClass != "general" {
			continue
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{
			ID: f.ID + ".existential", FrameID: f.ID, Location: true, Existential: true, Analyze: true, Realize: true,
			Sequence: []grammar.Element{{Slot: grammar.ExistentialMarker}, {Slot: grammar.Copula}, {Slot: grammar.SubjectPhrase}, {Slot: grammar.AdpositionPhrase}, {Slot: grammar.LocationSitePhrase}},
		})
	}
	return d
}
