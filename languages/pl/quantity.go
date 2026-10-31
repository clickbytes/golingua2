package pl

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
	d.MorphClasses = append(d.MorphClasses,
		morph.Class{ID: "pl.numeral.invariant", Language: "pl", POS: lexicon.Numeral, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "pl.noun.mug", Language: "pl", POS: lexicon.Noun, Rules: []morph.Rule{
			lemmaCell(features.Spec{Number: "singular", Case: "nominative"}), stemCell(features.Spec{Number: "singular", Case: "locative"}, "root", "u"), stemCell(features.Spec{Number: "plural", Case: "nominative"}, "root", "i"),
		}},
		morph.Class{ID: "pl.verb.see-personal", Language: "pl", POS: lexicon.Verb, Rules: []morph.Rule{
			stemCell(features.Spec{Number: "singular", Person: "first", Tense: "present"}, "root", "ę"),
			stemCell(features.Spec{Number: "singular", Person: "third", Tense: "present"}, "root", "i"),
		}},
	)
	for _, n := range []struct {
		value        int
		form, gender string
	}{
		{1, "jeden", "masculine"}, {1, "jedna", "feminine"}, {1, "jedno", "neuter"},
		{2, "dwa", "masculine"}, {2, "dwie", "feminine"}, {2, "dwa", "neuter"},
		{3, "trzy", ""}, {4, "cztery", ""}, {5, "pięć", ""},
	} {
		id := fmt.Sprintf("pl.numeral.%d.%s", n.value, n.gender)
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", POS: lexicon.Numeral, Lemma: n.form, ClassID: "pl.numeral.invariant", NumeralValue: n.value, Gender: n.gender})
	}
	for i := range d.MorphClasses {
		cl := &d.MorphClasses[i]
		switch cl.ID {
		case "pl.noun.feminine.a-k":
			cl.Rules = append(cl.Rules, morph.Rule{Features: features.Spec{Number: "plural", Case: "genitive"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "root"}, {Kind: morph.ReplaceSuffix, Arg: "k", With: "ek"}}})
		case "pl.noun.masculine.animal":
			cl.Rules = append(cl.Rules, stemCell(features.Spec{Number: "plural", Case: "accusative"}, "root", "y"))
		case "pl.noun.masculine.inanimate":
			cl.Rules = append(cl.Rules, stemCell(features.Spec{Number: "plural", Case: "nominative"}, "root", "y"))
		case "pl.copula.present":
			cl.Rules = append(cl.Rules, stemCell(features.Spec{Number: "plural", Person: "third", Tense: "present"}, "present_3pl", ""))
		}
	}
	for i := range d.Lexemes {
		switch d.Lexemes[i].ID {
		case "pl.verb.byc":
			d.Lexemes[i].Stems["present_3pl"] = "są"
		case "pl.verb.see":
			d.Lexemes[i].ClassID = "pl.verb.see-personal"
		}
	}
	d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: "pl.noun.mug", Language: "pl", POS: lexicon.Noun, Lemma: "kubek", ClassID: "pl.noun.mug", Gender: "masculine", Countability: "count", Stems: map[string]string{"root": "kubk"}})
	d.Senses = append(d.Senses, lexicon.Sense{ID: "pl.sense.noun.mug", LexemeID: "pl.noun.mug", Concept: core.ID("entity.mug")})
	// Numeral constructions reuse approved object frames: new implicit subject
	// is licensed for the first person on the existing accusative frame.
	for i := range d.EventFrames {
		if d.EventFrames[i].ID == "pl.frame.acc" {
			d.EventFrames[i].AllowImplicitSubject = true
		}
	}
	d.Constructions = append(d.Constructions,
		grammar.Pattern{ID: "pl.frame.acc.implicit", FrameID: "pl.frame.acc", Event: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.PredicateVerb}, {Slot: grammar.ObjectPhrase}}},
	)
	for _, f := range d.LocationFrames {
		if f.SiteClass != "general" {
			continue
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{
			ID: f.ID + ".existential", FrameID: f.ID, Location: true, Existential: true, Analyze: true, Realize: true,
			Sequence: []grammar.Element{{Slot: grammar.AdpositionPhrase}, {Slot: grammar.LocationSitePhrase}, {Slot: grammar.Copula}, {Slot: grammar.SubjectPhrase}},
		})
	}
	return d
}
