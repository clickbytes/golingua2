package pl

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// Location paradigms share ordinary inflection cells. Irregular stem changes
// are explicit in a class or in an attested lexeme, not in sentence processing.
func addLocation(d compile.Definition) compile.Definition {
	for i := range d.MorphClasses {
		switch d.MorphClasses[i].ID {
		case "pl.noun.masculine.inanimate":
			d.MorphClasses[i].Rules = append(d.MorphClasses[i].Rules,
				stemCell(features.Spec{Number: "singular", Case: "locative"}, "root", "u"),
				stemCell(features.Spec{Number: "singular", Case: "instrumental"}, "root", "em"))
		}
	}
	d.MorphClasses = append(d.MorphClasses,
		morph.Class{ID: "pl.noun.masculine.locative-l", Language: "pl", POS: lexicon.Noun,
			Rules: append(nounRules("u", "lemma", ""),
				morph.Rule{Features: features.Spec{Number: "singular", Case: "locative"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "root"}, {Kind: morph.ReplaceSuffix, Arg: "ł", With: "le"}}},
				stemCell(features.Spec{Number: "singular", Case: "instrumental"}, "root", "em"))},
		morph.Class{ID: "pl.noun.feminine.locative-l", Language: "pl", POS: lexicon.Noun,
			Rules: append(nounRules("y", "ę", "y"),
				morph.Rule{Features: features.Spec{Number: "singular", Case: "locative"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "root"}, {Kind: morph.ReplaceSuffix, Arg: "ł", With: "le"}}})},
		morph.Class{ID: "pl.noun.feminine.ia", Language: "pl", POS: lexicon.Noun,
			Rules: []morph.Rule{lemmaCell(features.Spec{Number: "singular", Case: "nominative"}), stemCell(features.Spec{Number: "singular", Case: "genitive"}, "root", "i"), stemCell(features.Spec{Number: "singular", Case: "locative"}, "root", "i")}},
		morph.Class{ID: "pl.noun.neuter.locative-ie", Language: "pl", POS: lexicon.Noun,
			Rules: append(nounRules("a", "lemma", "a"), stemCell(features.Spec{Number: "singular", Case: "locative"}, "root", "ie"))},
		morph.Class{ID: "pl.adposition.invariant", Language: "pl", POS: lexicon.Adposition, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
	)
	for i := range d.Lexemes {
		switch d.Lexemes[i].ID {
		case "pl.noun.table":
			d.Lexemes[i].ClassID = "pl.noun.masculine.locative-l"
		case "pl.noun.school":
			d.Lexemes[i].ClassID = "pl.noun.feminine.locative-l"
			d.Lexemes[i].SiteClass = "institution"
		}
	}
	for _, x := range []struct{ concept, lemma, root, gender, class, siteClass string }{
		{"kitchen", "kuchnia", "kuchn", "feminine", "pl.noun.feminine.ia", ""},
		{"tree", "drzewo", "drzew", "neuter", "pl.noun.neuter.locative-ie", "tree"},
		{"room", "pokój", "pokoj", "masculine", "pl.noun.masculine.inanimate", ""},
	} {
		id := "pl.noun." + x.concept
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", Lemma: x.lemma, POS: lexicon.Noun, ClassID: x.class, Stems: map[string]string{"root": x.root}, Gender: x.gender, SiteClass: x.siteClass})
		d.Senses = append(d.Senses, lexicon.Sense{ID: "pl.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for _, x := range []struct{ key, lemma string }{{"w", "w"}, {"na", "na"}, {"pod", "pod"}, {"przed", "przed"}, {"za", "za"}, {"obok", "obok"}} {
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: "pl.adposition." + x.key, Language: "pl", Lemma: x.lemma, POS: lexicon.Adposition, ClassID: "pl.adposition.invariant"})
	}
	for _, f := range []struct{ relation, siteClass, siteCase, adposition string }{
		{"inside", "general", "locative", "w"},
		{"on_support", "general", "locative", "na"},
		{"below", "general", "instrumental", "pod"},
		{"at_site", "institution", "locative", "w"},
		{"at_site", "tree", "locative", "na"},
		{"in_front", "general", "instrumental", "przed"},
		{"behind", "general", "instrumental", "za"},
		{"adjacent", "general", "genitive", "obok"},
	} {
		frameID := "pl.location." + f.relation + "." + f.siteClass
		d.LocationFrames = append(d.LocationFrames, grammar.LocationFrame{
			ID: frameID, Predicate: core.ID("relation." + f.relation), LocatedRole: "located", SiteRole: "site",
			SiteClass: f.siteClass, SiteCase: f.siteCase, AdpositionID: "pl.adposition." + f.adposition,
		})
		d.Constructions = append(d.Constructions, grammar.Pattern{
			ID: frameID + ".clause", FrameID: frameID, Location: true, Analyze: true, Realize: true,
			Sequence: []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.Copula}, {Slot: grammar.AdpositionPhrase}, {Slot: grammar.LocationSitePhrase}},
		})
	}
	return d
}
