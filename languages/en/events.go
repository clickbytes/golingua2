package en

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// English third-singular morphology uses reusable classes; has is the one
// explicit exception in this first slice, not an extra sentence rule.
func englishPresentRules(thirdSteps []morph.Operation) []morph.Rule {
	rules := make([]morph.Rule, 0, 6)
	for _, number := range []string{"singular", "plural"} {
		for _, person := range []string{"first", "second", "third"} {
			steps := []morph.Operation{{Kind: morph.SelectLemma}}
			if number == "singular" && person == "third" {
				steps = thirdSteps
			}
			rules = append(rules, morph.Rule{Features: features.Spec{Number: number, Person: person, Tense: "present"}, Steps: steps})
		}
	}
	return rules
}

// The lexical base form is distinct from finite present: do-support must not
// borrow a third-plural agreement cell as a shortcut.
func englishVerbRules(thirdSteps []morph.Operation) []morph.Rule {
	rules := englishPresentRules(thirdSteps)
	return append(rules, morph.Rule{Features: features.Spec{VerbForm: "base"}, Steps: []morph.Operation{{Kind: morph.SelectLemma}}})
}

func addEvents(d compile.Definition) compile.Definition {
	d.MorphClasses = append(d.MorphClasses,
		morph.Class{ID: "en.verb.present-s", Language: "en", POS: lexicon.Verb, Rules: englishVerbRules([]morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.Append, Arg: "s"}})},
		morph.Class{ID: "en.verb.present-y", Language: "en", POS: lexicon.Verb, Rules: englishVerbRules([]morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.ReplaceSuffix, Arg: "y", With: "ies"}})},
		morph.Class{ID: "en.pronoun.invariant", Language: "en", POS: lexicon.Pronoun, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "en.article.definite", Language: "en", POS: lexicon.Article, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "en.auxiliary.do-present", Language: "en", POS: lexicon.Auxiliary, Rules: englishPresentRules([]morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.Append, Arg: "es"}})},
		morph.Class{ID: "en.particle.invariant", Language: "en", POS: lexicon.Particle, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "en.noun.mass", Language: "en", POS: lexicon.Noun, Rules: []morph.Rule{lemmaCell(features.Spec{Number: "singular", Case: "nominative"})}},
	)
	d.Specials.DefiniteArticle = "en.article.the"
	d.Specials.NegationMarker = "en.particle.not"
	d.Specials.NegationAuxiliary = "en.auxiliary.do"
	d.Lexemes = append(d.Lexemes,
		lexicon.Entry{ID: d.Specials.NegationMarker, Lemma: "not", Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant"},
		lexicon.Entry{ID: d.Specials.NegationAuxiliary, Lemma: "do", Language: "en", POS: lexicon.Auxiliary, ClassID: "en.auxiliary.do-present"},
	)
	d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: "en.article.the", Lemma: "the", Language: "en", POS: lexicon.Article, ClassID: "en.article.definite"})
	nounData := []struct{ concept, lemma, onset, class, countability string }{
		{"bicycle", "bicycle", "consonant", "en.noun.regular-s", "count"},
		{"car", "car", "consonant", "en.noun.regular-s", "count"},
		{"phone", "phone", "consonant", "en.noun.regular-s", "count"},
		{"water", "water", "consonant", "en.noun.mass", "mass"},
		{"letter", "letter", "consonant", "en.noun.regular-s", "count"},
		{"bird", "bird", "consonant", "en.noun.regular-s", "count"},
		{"girl", "girl", "consonant", "en.noun.regular-s", "count"},
		{"boy", "boy", "consonant", "en.noun.regular-s", "count"},
		{"mother", "mother", "consonant", "en.noun.regular-s", "count"},
		{"father", "father", "consonant", "en.noun.regular-s", "count"},
		{"door", "door", "consonant", "en.noun.regular-s", "count"},
		{"doctor", "doctor", "consonant", "en.noun.regular-s", "count"},
		{"child", "child", "consonant", "en.noun.regular-s", "count"},
	}
	for _, x := range nounData {
		id := "en.noun." + x.concept
		entry := lexicon.Entry{ID: id, Language: "en", Lemma: x.lemma, POS: lexicon.Noun, ClassID: x.class, Onset: x.onset, Countability: x.countability}
		if x.concept == "child" {
			entry.Overrides = map[string]string{(features.Spec{Number: "plural", Case: "nominative"}).Key(): "children"}
		}
		d.Lexemes = append(d.Lexemes, entry)
		d.Senses = append(d.Senses, lexicon.Sense{ID: "en.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for i := range d.Lexemes {
		if d.Lexemes[i].ID == "en.noun.coffee" {
			d.Lexemes[i].Countability = "mass"
			d.Lexemes[i].ClassID = "en.noun.mass"
		}
	}
	for _, x := range []struct{ symbol, person, number, gender string }{
		{"i", "first", "singular", ""}, {"you", "second", "", ""},
		{"he", "third", "singular", "masculine"}, {"she", "third", "singular", "feminine"},
		{"we", "first", "plural", ""}, {"they", "third", "plural", ""},
	} {
		id := "en.pronoun." + x.symbol
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "en", Lemma: x.symbol, POS: lexicon.Pronoun, ClassID: "en.pronoun.invariant", PronounPerson: x.person, PronounNumber: x.number, PronounGender: x.gender})
		d.Senses = append(d.Senses, lexicon.Sense{ID: "en.sense.pronoun." + x.symbol, LexemeID: id, Concept: "entity.person"})
	}
	verbData := []struct{ concept, lemma, frame, class string }{
		{"have", "have", "have", "en.verb.present-s"},
		{"read", "read", "acc", "en.verb.present-s"},
		{"see", "see", "acc", "en.verb.present-s"},
		{"drink", "drink", "acc", "en.verb.present-s"},
		{"eat", "eat", "acc", "en.verb.present-s"},
		{"open", "open", "acc", "en.verb.present-s"},
		{"close", "close", "acc", "en.verb.present-s"},
		{"help", "help", "dat", "en.verb.present-s"},
		{"write", "write", "acc", "en.verb.present-s"},
		{"sleep", "sleep", "intr", "en.verb.present-s"},
		{"run", "run", "intr", "en.verb.present-s"},
		{"fly", "fly", "intr", "en.verb.present-y"},
		{"sit", "sit", "intr", "en.verb.present-s"},
		{"work", "work", "intr", "en.verb.present-s"},
		{"cook", "cook", "intr", "en.verb.present-s"},
		{"walk", "walk", "intr", "en.verb.present-s"},
		{"sing", "sing", "intr", "en.verb.present-s"},
		{"bark", "bark", "intr", "en.verb.present-s"},
		{"cry", "cry", "intr", "en.verb.present-y"},
	}
	for _, x := range verbData {
		id := "en.verb." + x.concept
		entry := lexicon.Entry{ID: id, Language: "en", Lemma: x.lemma, POS: lexicon.Verb, ClassID: x.class}
		if x.concept == "have" {
			entry.Overrides = map[string]string{(features.Spec{Number: "singular", Person: "third", Tense: "present"}).Key(): "has"}
		}
		d.Lexemes = append(d.Lexemes, entry)
		d.Senses = append(d.Senses, lexicon.Sense{ID: "en.sense.verb." + x.concept, LexemeID: id, Concept: core.ID("predicate." + x.concept), FrameIDs: []string{"en.frame." + x.frame}})
	}
	d.EventFrames = append(d.EventFrames,
		grammar.EventFrame{ID: "en.frame.have", SubjectRole: "holder", ObjectRole: "theme", ObjectCase: "nominative"},
		grammar.EventFrame{ID: "en.frame.acc", SubjectRole: "agent", ObjectRole: "patient", ObjectCase: "nominative"},
		grammar.EventFrame{ID: "en.frame.dat", SubjectRole: "agent", ObjectRole: "beneficiary", ObjectCase: "nominative"},
		grammar.EventFrame{ID: "en.frame.intr", SubjectRole: "agent"},
	)
	for _, f := range d.EventFrames {
		seq := []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.PredicateVerb}}
		if f.ObjectRole != "" {
			seq = append(seq, grammar.Element{Slot: grammar.ObjectPhrase})
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".clause", FrameID: f.ID, Event: true, Analyze: true, Realize: true, Sequence: seq})
		negative := []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.NegationAuxiliary}, {Slot: grammar.NegationMarker}, {Slot: grammar.PredicateVerb}}
		if f.ObjectRole != "" {
			negative = append(negative, grammar.Element{Slot: grammar.ObjectPhrase})
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".negative", FrameID: f.ID, Event: true, Negated: true, Analyze: true, Realize: true, Sequence: negative})
	}
	return d
}
