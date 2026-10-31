package pl

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// One past-participle suffix program serves all licensed verbs with a past stem.
// Form gender describes grammatical agreement, not an inferred person gender.
func polishTemporalRules(presentEnding string, infinitive bool) []morph.Rule {
	out := []morph.Rule{stemCell(features.Spec{Number: "singular", Person: "third", Tense: "present"}, "root", presentEnding)}
	cells := []struct{ number, gender, suffix string }{
		{"singular", "masculine", "ł"}, {"singular", "feminine", "ła"}, {"singular", "neuter", "ło"},
		{"plural", "masculine_personal", "li"}, {"plural", "nonmasculine_personal", "ły"},
	}
	for _, c := range cells {
		out = append(out, stemCell(features.Spec{Tense: "past", Person: "third", Number: c.number, Gender: c.gender}, "past", c.suffix))
	}
	if infinitive {
		out = append(out, lemmaCell(features.Spec{VerbForm: "infinitive"}))
	}
	return out
}

func addTime(d compile.Definition) compile.Definition {
	for _, c := range []struct {
		id, ending string
		infinitive bool
	}{
		{"pl.verb.temporal-a", "a", false}, {"pl.verb.temporal-e", "e", false}, {"pl.verb.temporal-i", "i", false}, {"pl.verb.temporal-e-inf", "e", true},
	} {
		d.MorphClasses = append(d.MorphClasses, morph.Class{ID: c.id, Language: "pl", POS: lexicon.Verb, Rules: polishTemporalRules(c.ending, c.infinitive)})
	}
	d.MorphClasses = append(d.MorphClasses,
		morph.Class{ID: "pl.verb.weather-past", Language: "pl", POS: lexicon.Verb, Rules: []morph.Rule{stemCell(features.Spec{Tense: "past", Person: "third", Number: "singular", Gender: "masculine"}, "past", "ł")}},
		morph.Class{ID: "pl.auxiliary.future", Language: "pl", POS: lexicon.Auxiliary, Rules: []morph.Rule{
			{Features: features.Spec{Tense: "future", Person: "third", Number: "singular"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "third_singular"}}},
			{Features: features.Spec{Tense: "future", Person: "first", Number: "singular"}, Steps: []morph.Operation{{Kind: morph.SelectStem, Arg: "first_singular"}}},
		}},
	)
	d.Specials.TimeMarkers = map[string]string{"past": "pl.time.wczoraj", "future": "pl.time.jutro"}
	d.Specials.FutureAuxiliary = "pl.auxiliary.byc_future"
	d.Lexemes = append(d.Lexemes,
		lexicon.Entry{ID: "pl.time.wczoraj", Lemma: "wczoraj", Language: "pl", POS: lexicon.Particle, ClassID: "pl.particle.invariant"},
		lexicon.Entry{ID: "pl.time.jutro", Lemma: "jutro", Language: "pl", POS: lexicon.Particle, ClassID: "pl.particle.invariant"},
		lexicon.Entry{ID: d.Specials.FutureAuxiliary, Lemma: "być", Language: "pl", POS: lexicon.Auxiliary, ClassID: "pl.auxiliary.future", Stems: map[string]string{"third_singular": "będzie", "first_singular": "będę"}},
		lexicon.Entry{ID: "pl.noun.rain", Lemma: "deszcz", Language: "pl", POS: lexicon.Noun, ClassID: "pl.noun.masculine.inanimate", Countability: "mass", Gender: "masculine", Stems: map[string]string{"root": "deszcz"}},
		lexicon.Entry{ID: "pl.verb.rain", Lemma: "padać", Language: "pl", POS: lexicon.Verb, ClassID: "pl.verb.weather-past", Stems: map[string]string{"past": "pada"}},
	)
	d.Senses = append(d.Senses,
		lexicon.Sense{ID: "pl.sense.noun.rain", LexemeID: "pl.noun.rain", Concept: core.ID("entity.rain")},
		lexicon.Sense{ID: "pl.sense.verb.rain", LexemeID: "pl.verb.rain", Concept: core.ID("predicate.rain"), FrameIDs: []string{"pl.frame.weather"}},
	)
	for i := range d.Lexemes {
		switch d.Lexemes[i].ID {
		case "pl.verb.sleep":
			d.Lexemes[i].ClassID = "pl.verb.temporal-i"
			d.Lexemes[i].Stems["past"] = "spa"
		case "pl.verb.run":
			d.Lexemes[i].ClassID = "pl.verb.temporal-a"
			d.Lexemes[i].Stems["past"] = "biega"
		case "pl.verb.read":
			d.Lexemes[i].ClassID = "pl.verb.temporal-a"
			d.Lexemes[i].Stems["past"] = "czyta"
		case "pl.verb.cook":
			d.Lexemes[i].ClassID = "pl.verb.temporal-e"
			d.Lexemes[i].Stems["past"] = "gotowa"
		case "pl.verb.work":
			d.Lexemes[i].ClassID = "pl.verb.temporal-e-inf"
			d.Lexemes[i].Stems["past"] = "pracowa"
		}
	}
	d.EventFrames = append(d.EventFrames, grammar.EventFrame{ID: "pl.frame.weather", SubjectRole: "theme", Weather: true, WeatherSubjectConcept: "entity.rain"})
	for i := range d.EventFrames {
		if d.EventFrames[i].ID == "pl.frame.intr" {
			d.EventFrames[i].AllowImplicitSubject = true
		}
	}
	for _, f := range d.EventFrames {
		if f.ID != "pl.frame.intr" && f.ID != "pl.frame.acc" && f.ID != "pl.frame.weather" {
			continue
		}
		if f.Weather {
			d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".past.nominal", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: "past", Weather: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.TemporalAdverb}, {Slot: grammar.PredicateVerb}, {Slot: grammar.SubjectPhrase}}})
			continue
		}
		for _, tense := range []string{"past", "future"} {
			seq := []grammar.Element{{Slot: grammar.TemporalAdverb}, {Slot: grammar.SubjectPhrase}}
			if tense == "future" {
				seq = append(seq, grammar.Element{Slot: grammar.FutureAuxiliary})
			}
			seq = append(seq, grammar.Element{Slot: grammar.PredicateVerb})
			if f.ObjectRole != "" {
				seq = append(seq, grammar.Element{Slot: grammar.ObjectPhrase})
			}
			d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + "." + tense + ".explicit", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: tense, Analyze: true, Realize: true, Sequence: seq})
			if tense == "future" && f.ObjectRole == "" {
				d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".future.implicit", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: "future", Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.TemporalAdverb}, {Slot: grammar.FutureAuxiliary}, {Slot: grammar.PredicateVerb}}})
			}
		}
	}
	return d
}
