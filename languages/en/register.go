// Package en declares English lexical and grammatical data; no translation pairs.
package en

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
func suffixCell(f features.Spec, suffix string) morph.Rule {
	return morph.Rule{Features: f, Steps: []morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.Append, Arg: suffix}}}
}

func Definition() compile.Definition {
	sg := features.Spec{Number: "singular", Case: "nominative"}
	pl := features.Spec{Number: "plural", Case: "nominative"}
	classes := []morph.Class{
		{ID: "en.noun.regular-s", Language: "en", POS: lexicon.Noun, Rules: []morph.Rule{lemmaCell(sg), suffixCell(pl, "s")}},
		{ID: "en.noun.consonant-y", Language: "en", POS: lexicon.Noun, Rules: []morph.Rule{lemmaCell(sg), {Features: pl, Steps: []morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.ReplaceSuffix, Arg: "y", With: "ies"}}}}},
		{ID: "en.noun.es", Language: "en", POS: lexicon.Noun, Rules: []morph.Rule{lemmaCell(sg), suffixCell(pl, "es")}},
		{ID: "en.adjective.invariant", Language: "en", POS: lexicon.Adjective, Rules: []morph.Rule{lemmaCell(features.Spec{Degree: "positive"})}},
		{ID: "en.demonstrative.proximal", Language: "en", POS: lexicon.Determiner, Rules: []morph.Rule{lemmaCell(features.Spec{Number: "singular"})}},
		{ID: "en.copula.present", Language: "en", POS: lexicon.Verb, Rules: []morph.Rule{{Features: features.Spec{Number: "singular", Person: "third", Tense: "present"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "present_3sg"}}}}},
		{ID: "en.article.indefinite", Language: "en", POS: lexicon.Article, Rules: []morph.Rule{
			{Features: features.Spec{Onset: "consonant"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "consonant"}}},
			{Features: features.Spec{Onset: "vowel"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "vowel"}}},
		}},
	}
	data := []struct{ concept, lemma, onset string }{
		{"cat", "cat", "consonant"}, {"dog", "dog", "consonant"}, {"book", "book", "consonant"},
		{"house", "house", "consonant"}, {"table", "table", "consonant"}, {"chair", "chair", "consonant"},
		{"teacher", "teacher", "consonant"}, {"student", "student", "consonant"}, {"apple", "apple", "vowel"},
		{"window", "window", "consonant"}, {"bag", "bag", "consonant"}, {"coffee", "coffee", "consonant"},
		{"school", "school", "consonant"},
	}
	entries := make([]lexicon.Entry, 0, len(data)+10)
	senses := make([]lexicon.Sense, 0, len(data)+7)
	for _, x := range data {
		id := "en.noun." + x.concept
		entries = append(entries, lexicon.Entry{ID: id, Language: "en", Lemma: x.lemma, POS: lexicon.Noun, ClassID: "en.noun.regular-s", Onset: x.onset})
		senses = append(senses, lexicon.Sense{ID: "en.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for _, x := range []struct{ concept, lemma string }{{"black", "black"}, {"small", "small"}, {"new", "new"}, {"big", "big"}, {"open", "open"}, {"red", "red"}, {"hot", "hot"}} {
		id := "en.adjective." + x.concept
		entries = append(entries, lexicon.Entry{ID: id, Language: "en", Lemma: x.lemma, POS: lexicon.Adjective, ClassID: "en.adjective.invariant"})
		senses = append(senses, lexicon.Sense{ID: "en.sense.adjective." + x.concept, LexemeID: id, Concept: core.ID("property." + x.concept)})
	}
	entries = append(entries,
		lexicon.Entry{ID: "en.demo.this", Language: "en", Lemma: "this", POS: lexicon.Determiner, ClassID: "en.demonstrative.proximal"},
		lexicon.Entry{ID: "en.verb.be", Language: "en", Lemma: "be", POS: lexicon.Verb, ClassID: "en.copula.present", Stems: map[string]string{"present_3sg": "is"}},
		lexicon.Entry{ID: "en.article.a", Language: "en", Lemma: "a", POS: lexicon.Article, ClassID: "en.article.indefinite", Stems: map[string]string{"consonant": "a", "vowel": "an"}},
	)
	return compile.Definition{Language: "en", Lexemes: entries, Senses: senses, MorphClasses: classes,
		Frames: []grammar.Frame{
			{ID: "en.frame.classification", Predicate: "predicate.is_a", Role: "member", ValueSlot: grammar.Noun, ReferentConcept: "entity.unspecified"},
			{ID: "en.frame.property", Predicate: "predicate.has_property", Role: "bearer", ValueSlot: grammar.Adjective},
		},
		Specials: compile.Specials{Demonstrative: "en.demo.this", Copula: "en.verb.be", Article: "en.article.a", DemoDenotesProximal: true},
		Constructions: []grammar.Pattern{
			{ID: "en.classification.proximal", FrameID: "en.frame.classification", Predicate: "predicate.is_a", Analyze: true, Realize: true, DemoGender: "invariant", Sequence: []grammar.Element{{Slot: grammar.Demonstrative}, {Slot: grammar.Copula}, {Slot: grammar.Article}, {Slot: grammar.Noun}}},
			{ID: "en.property.proximal", FrameID: "en.frame.property", Predicate: "predicate.has_property", Analyze: true, Realize: true, DemoGender: "invariant", Sequence: []grammar.Element{{Slot: grammar.Demonstrative}, {Slot: grammar.Noun}, {Slot: grammar.Copula}, {Slot: grammar.Adjective}}},
		},
	}
}
