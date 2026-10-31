package pl

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// The same suffix class handles every compatible verb with an attested stem;
// lexical identity and valency are attached to sense/frame records, not a
// switch on the lemma or a full-sentence string.
func polishPresent3SG(ending string) morph.Class {
	return morph.Class{ID: "pl.verb.present-" + ending, Language: "pl", POS: lexicon.Verb, Rules: []morph.Rule{stemCell(features.Spec{Number: "singular", Person: "third", Tense: "present"}, "root", ending)}}
}
func polishNounPLCase(gramCase string) morph.Rule {
	return lemmaCell(features.Spec{Number: "plural", Case: gramCase})
}
func addEvents(d compile.Definition) compile.Definition {
	d.MorphClasses = append(d.MorphClasses,
		polishPresent3SG("a"), polishPresent3SG("e"), polishPresent3SG("i"),
		morph.Class{ID: "pl.verb.have-present", Language: "pl", POS: lexicon.Verb, Rules: []morph.Rule{
			stemCell(features.Spec{Number: "singular", Person: "first", Tense: "present"}, "root", "am"),
			stemCell(features.Spec{Number: "singular", Person: "second", Tense: "present"}, "root", "asz"),
			stemCell(features.Spec{Number: "singular", Person: "third", Tense: "present"}, "root", "a"),
			stemCell(features.Spec{Number: "plural", Person: "first", Tense: "present"}, "root", "amy"),
			stemCell(features.Spec{Number: "plural", Person: "second", Tense: "present"}, "root", "acie"),
			stemCell(features.Spec{Number: "plural", Person: "third", Tense: "present"}, "root", "ają"),
		}},
		morph.Class{ID: "pl.particle.invariant", Language: "pl", POS: lexicon.Particle, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "pl.pronoun.invariant", Language: "pl", POS: lexicon.Pronoun, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
		morph.Class{ID: "pl.noun.plural-only", Language: "pl", POS: lexicon.Noun, Rules: []morph.Rule{polishNounPLCase("nominative"), polishNounPLCase("accusative")}},
		morph.Class{ID: "pl.noun.masculine.a", Language: "pl", POS: lexicon.Noun, Rules: []morph.Rule{lemmaCell(features.Spec{Number: "singular", Case: "nominative"}), stemCell(features.Spec{Number: "singular", Case: "accusative"}, "root", "ę")}},
	)
	d.Specials.NegationMarker = "pl.particle.nie"
	d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: d.Specials.NegationMarker, Language: "pl", Lemma: "nie", POS: lexicon.Particle, ClassID: "pl.particle.invariant"})
	// The ordinary neuter -o class provides regular dative -u as required by
	// pomaga dziecku; the object-case license lives on the frame, not the noun.
	for i := range d.MorphClasses {
		if d.MorphClasses[i].ID == "pl.noun.neuter.o" {
			d.MorphClasses[i].Rules = append(d.MorphClasses[i].Rules, stemCell(features.Spec{Number: "singular", Case: "dative"}, "root", "u"))
		}
	}
	nounData := []struct{ concept, lemma, root, gender, class, countability string }{
		{"bicycle", "rower", "rower", "masculine", "pl.noun.masculine.inanimate", "count"},
		{"car", "samochód", "samochod", "masculine", "pl.noun.masculine.inanimate", "count"},
		{"phone", "telefon", "telefon", "masculine", "pl.noun.masculine.inanimate", "count"},
		{"water", "woda", "wod", "feminine", "pl.noun.feminine.a", "mass"},
		{"letter", "list", "list", "masculine", "pl.noun.masculine.inanimate", "count"},
		{"bird", "ptak", "ptak", "masculine", "pl.noun.masculine.animal", "count"},
		{"girl", "dziewczynka", "dziewczynk", "feminine", "pl.noun.feminine.a-k", "count"},
		{"boy", "chłopiec", "chłopc", "masculine", "pl.noun.masculine.person", "count"},
		{"mother", "mama", "mam", "feminine", "pl.noun.feminine.a", "count"},
		{"father", "tata", "tat", "masculine", "pl.noun.masculine.a", "count"},
		{"door", "drzwi", "drzwi", "nonmasculine", "pl.noun.plural-only", "plural_only"},
		{"doctor", "lekarz", "lekarz", "masculine", "pl.noun.masculine.person", "count"},
		{"child", "dziecko", "dzieck", "neuter", "pl.noun.neuter.o", "count"},
	}
	for _, x := range nounData {
		id := "pl.noun." + x.concept
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", Lemma: x.lemma, POS: lexicon.Noun, ClassID: x.class, Stems: map[string]string{"root": x.root}, Gender: x.gender, Countability: x.countability})
		d.Senses = append(d.Senses, lexicon.Sense{ID: "pl.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for _, x := range []struct{ symbol, person, number, gender string }{
		{"on", "third", "singular", "masculine"}, {"ona", "third", "singular", "feminine"},
		{"oni", "third", "plural", "masculine_personal"}, {"one", "third", "plural", "nonmasculine_personal"},
	} {
		id := "pl.pronoun." + x.symbol
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", Lemma: x.symbol, POS: lexicon.Pronoun, ClassID: "pl.pronoun.invariant", PronounPerson: x.person, PronounNumber: x.number, PronounGender: x.gender})
		d.Senses = append(d.Senses, lexicon.Sense{ID: "pl.sense.pronoun." + x.symbol, LexemeID: id, Concept: "entity.person"})
	}
	verbData := []struct{ concept, lemma, root, frame, class string }{
		{"have", "mieć", "m", "have", "pl.verb.have-present"},
		{"read", "czytać", "czyt", "acc", "pl.verb.present-a"},
		{"see", "widzieć", "widz", "acc", "pl.verb.present-i"},
		{"drink", "pić", "pij", "acc", "pl.verb.present-e"},
		{"eat", "jeść", "j", "acc", "pl.verb.present-e"},
		{"open", "otwierać", "otwier", "acc", "pl.verb.present-a"},
		{"close", "zamykać", "zamyk", "acc", "pl.verb.present-a"},
		{"help", "pomagać", "pomag", "dat", "pl.verb.present-a"},
		{"write", "pisać", "pisz", "acc", "pl.verb.present-e"},
		{"sleep", "spać", "śp", "intr", "pl.verb.present-i"},
		{"run", "biegać", "bieg", "intr", "pl.verb.present-a"},
		{"fly", "latać", "lat", "intr", "pl.verb.present-a"},
		{"sit", "siedzieć", "siedz", "intr", "pl.verb.present-i"},
		{"work", "pracować", "pracuj", "intr", "pl.verb.present-e"},
		{"cook", "gotować", "gotuj", "intr", "pl.verb.present-e"},
		{"walk", "chodzić", "chodz", "intr", "pl.verb.present-i"},
		{"sing", "śpiewać", "śpiew", "intr", "pl.verb.present-a"},
		{"bark", "szczekać", "szczek", "intr", "pl.verb.present-a"},
		{"cry", "płakać", "płacz", "intr", "pl.verb.present-e"},
	}
	for _, x := range verbData {
		id := "pl.verb." + x.concept
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", Lemma: x.lemma, POS: lexicon.Verb, ClassID: x.class, Stems: map[string]string{"root": x.root}})
		d.Senses = append(d.Senses, lexicon.Sense{ID: "pl.sense.verb." + x.concept, LexemeID: id, Concept: core.ID("predicate." + x.concept), FrameIDs: []string{"pl.frame." + x.frame}})
	}
	d.EventFrames = append(d.EventFrames,
		grammar.EventFrame{ID: "pl.frame.have", SubjectRole: "holder", ObjectRole: "theme", ObjectCase: "accusative", NegatedObjectCase: "genitive", AllowImplicitSubject: true},
		grammar.EventFrame{ID: "pl.frame.acc", SubjectRole: "agent", ObjectRole: "patient", ObjectCase: "accusative", NegatedObjectCase: "genitive"},
		grammar.EventFrame{ID: "pl.frame.dat", SubjectRole: "agent", ObjectRole: "beneficiary", ObjectCase: "dative"},
		grammar.EventFrame{ID: "pl.frame.intr", SubjectRole: "agent"},
	)
	for _, f := range d.EventFrames {
		seq := []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.PredicateVerb}}
		if f.ObjectRole != "" {
			seq = append(seq, grammar.Element{Slot: grammar.ObjectPhrase})
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".explicit", FrameID: f.ID, Event: true, Analyze: true, Realize: true, Sequence: seq})
		negative := []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.NegationMarker}, {Slot: grammar.PredicateVerb}}
		if f.ObjectRole != "" {
			negative = append(negative, grammar.Element{Slot: grammar.ObjectPhrase})
		}
		d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".negative", FrameID: f.ID, Event: true, Negated: true, Analyze: true, Realize: true, Sequence: negative})
	}
	// The language allows omitted 1st/2nd person subjects only where the frame
	// licenses it; a definite participant anchor is still present in Meaning.
	d.Constructions = append(d.Constructions, grammar.Pattern{ID: "pl.frame.have.implicit", FrameID: "pl.frame.have", Event: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.PredicateVerb}, {Slot: grammar.ObjectPhrase}}})
	d.Constructions = append(d.Constructions, grammar.Pattern{ID: "pl.frame.have.negative-implicit", FrameID: "pl.frame.have", Event: true, Negated: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.NegationMarker}, {Slot: grammar.PredicateVerb}, {Slot: grammar.ObjectPhrase}}})
	return d
}
