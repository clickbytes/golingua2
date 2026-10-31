// Package pl declares Polish language-local data and procedural inflection classes.
package pl

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

func lemmaCell(f features.Spec) morph.Rule {
	return morph.Rule{Features: f, Steps: []morph.Operation{{Kind: morph.SelectLemma}}}
}
func stemCell(f features.Spec, stem, suffix string) morph.Rule {
	steps := []morph.Operation{{Kind: morph.SelectStem, Arg: stem}}
	if suffix != "" {
		steps = append(steps, morph.Operation{Kind: morph.Append, Arg: suffix})
	}
	return morph.Rule{Features: f, Steps: steps}
}
func nounRules(gen, acc, pl string) []morph.Rule {
	result := []morph.Rule{lemmaCell(features.Spec{Number: "singular", Case: "nominative"})}
	if gen != "" {
		result = append(result, stemCell(features.Spec{Number: "singular", Case: "genitive"}, "root", gen))
	}
	if acc == "lemma" {
		result = append(result, lemmaCell(features.Spec{Number: "singular", Case: "accusative"}))
	} else if acc != "" {
		result = append(result, stemCell(features.Spec{Number: "singular", Case: "accusative"}, "root", acc))
	}
	if pl != "" {
		result = append(result, stemCell(features.Spec{Number: "plural", Case: "nominative"}, "root", pl))
	}
	return result
}

func Definition() compile.Definition {
	classes := []morph.Class{
		{ID: "pl.noun.masculine.animal", Language: "pl", POS: lexicon.Noun, Rules: nounRules("a", "a", "y")},
		{ID: "pl.noun.masculine.person", Language: "pl", POS: lexicon.Noun, Rules: nounRules("a", "a", "")},
		{ID: "pl.noun.masculine.inanimate", Language: "pl", POS: lexicon.Noun, Rules: nounRules("u", "lemma", "")},
		{ID: "pl.noun.feminine.a", Language: "pl", POS: lexicon.Noun, Rules: nounRules("y", "ę", "y")},
		{ID: "pl.noun.feminine.a-k", Language: "pl", POS: lexicon.Noun, Rules: append(nounRules("i", "ę", "i"), morph.Rule{Features: features.Spec{Number: "singular", Case: "dative"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "root"}, {Kind: morph.ReplaceSuffix, Arg: "k", With: "ce"}}})},
		{ID: "pl.noun.neuter.o", Language: "pl", POS: lexicon.Noun, Rules: nounRules("a", "lemma", "a")},
		{ID: "pl.adjective.hard-y", Language: "pl", POS: lexicon.Adjective, Rules: []morph.Rule{
			stemCell(features.Spec{Number: "singular", Case: "nominative", Gender: "masculine", Degree: "positive"}, "root", "y"),
			stemCell(features.Spec{Number: "singular", Case: "nominative", Gender: "feminine", Degree: "positive"}, "root", "a"),
			stemCell(features.Spec{Number: "singular", Case: "nominative", Gender: "neuter", Degree: "positive"}, "root", "e"),
		}},
		{ID: "pl.demonstrative.proximal", Language: "pl", POS: lexicon.Determiner, Rules: []morph.Rule{
			stemCell(features.Spec{Number: "singular", Gender: "masculine"}, "masculine", ""),
			stemCell(features.Spec{Number: "singular", Gender: "feminine"}, "feminine", ""),
			stemCell(features.Spec{Number: "singular", Gender: "neuter"}, "neuter", ""),
		}},
		{ID: "pl.copula.present", Language: "pl", POS: lexicon.Verb, Rules: []morph.Rule{stemCell(features.Spec{Number: "singular", Person: "third", Tense: "present"}, "present_3sg", "")}},
	}
	data := []struct{ concept, lemma, stem, gender, class string }{
		{"cat", "kot", "kot", "masculine", "pl.noun.masculine.animal"},
		{"dog", "pies", "ps", "masculine", "pl.noun.masculine.animal"},
		{"book", "książka", "książk", "feminine", "pl.noun.feminine.a-k"},
		{"house", "dom", "dom", "masculine", "pl.noun.masculine.inanimate"},
		{"table", "stół", "stoł", "masculine", "pl.noun.masculine.inanimate"},
		{"chair", "krzesło", "krzesł", "neuter", "pl.noun.neuter.o"},
		{"teacher", "nauczyciel", "nauczyciel", "masculine", "pl.noun.masculine.person"},
		{"student", "uczeń", "uczni", "masculine", "pl.noun.masculine.person"},
		{"apple", "jabłko", "jabłk", "neuter", "pl.noun.neuter.o"},
		{"window", "okno", "okn", "neuter", "pl.noun.neuter.o"},
		{"bag", "torba", "torb", "feminine", "pl.noun.feminine.a"},
		{"coffee", "kawa", "kaw", "feminine", "pl.noun.feminine.a"},
		{"school", "szkoła", "szkoł", "feminine", "pl.noun.feminine.a"},
	}
	entries := make([]lexicon.Entry, 0, len(data)+10)
	senses := make([]lexicon.Sense, 0, len(data)+7)
	for _, x := range data {
		id := "pl.noun." + x.concept
		entries = append(entries, lexicon.Entry{ID: id, Language: "pl", Lemma: x.lemma, POS: lexicon.Noun, ClassID: x.class, Stems: map[string]string{"root": x.stem}, Gender: x.gender})
		senses = append(senses, lexicon.Sense{ID: "pl.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for _, x := range []struct{ concept, lemma, root string }{
		{"black", "czarny", "czarn"}, {"small", "mały", "mał"}, {"new", "nowy", "now"},
		{"big", "duży", "duż"}, {"open", "otwarty", "otwart"}, {"red", "czerwony", "czerwon"},
		{"hot", "gorący", "gorąc"},
	} {
		id := "pl.adjective." + x.concept
		entries = append(entries, lexicon.Entry{ID: id, Language: "pl", Lemma: x.lemma, POS: lexicon.Adjective, ClassID: "pl.adjective.hard-y", Stems: map[string]string{"root": x.root}})
		senses = append(senses, lexicon.Sense{ID: "pl.sense.adjective." + x.concept, LexemeID: id, Concept: core.ID("property." + x.concept)})
	}
	entries = append(entries,
		lexicon.Entry{ID: "pl.demo.ten", Language: "pl", Lemma: "ten", POS: lexicon.Determiner, ClassID: "pl.demonstrative.proximal", Stems: map[string]string{"masculine": "ten", "feminine": "ta", "neuter": "to"}},
		lexicon.Entry{ID: "pl.verb.byc", Language: "pl", Lemma: "być", POS: lexicon.Verb, ClassID: "pl.copula.present", Stems: map[string]string{"present_3sg": "jest"}},
	)
	return addLocation(addEvents(compile.Definition{Language: "pl", Lexemes: entries, Senses: senses, MorphClasses: classes,
		Frames: []grammar.Frame{
			{ID: "pl.frame.classification", Predicate: "predicate.is_a", Role: "member", ValueSlot: grammar.Noun, ReferentConcept: "entity.unspecified"},
			{ID: "pl.frame.property", Predicate: "predicate.has_property", Role: "bearer", ValueSlot: grammar.Adjective},
		},
		Specials: compile.Specials{Demonstrative: "pl.demo.ten", Copula: "pl.verb.byc"},
		Constructions: []grammar.Pattern{
			{ID: "pl.classification.proximal", FrameID: "pl.frame.classification", Predicate: "predicate.is_a", Analyze: true, Realize: true, DemoGender: "fixed_neuter", Sequence: []grammar.Element{{Slot: grammar.Demonstrative}, {Slot: grammar.Copula}, {Slot: grammar.Noun}}},
			{ID: "pl.property.proximal", FrameID: "pl.frame.property", Predicate: "predicate.has_property", Analyze: true, Realize: true, DemoGender: "agree", Sequence: []grammar.Element{{Slot: grammar.Demonstrative}, {Slot: grammar.Noun}, {Slot: grammar.Copula}, {Slot: grammar.Adjective}}},
		},
	}))
}
