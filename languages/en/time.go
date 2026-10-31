package en

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// Past cells are a reusable regular or lexical-principal-part algorithm.
// Inflection does not recognize a complete input sentence.
func englishTemporalRules(irregular bool) []morph.Rule {
	result := englishVerbRules([]morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.Append, Arg: "s"}})
	past := []morph.Operation{{Kind: morph.SelectLemma}, {Kind: morph.Append, Arg: "ed"}}
	if irregular {
		past = []morph.Operation{{Kind: morph.SelectStem, Arg: "past"}}
	}
	return append(result, morph.Rule{Features: features.Spec{Tense: "past"}, Steps: past})
}

func addTime(d compile.Definition) compile.Definition {
	d.MorphClasses = append(d.MorphClasses,
		morph.Class{ID: "en.verb.temporal-regular", Language: "en", POS: lexicon.Verb, Rules: englishTemporalRules(false)},
		morph.Class{ID: "en.verb.temporal-irregular", Language: "en", POS: lexicon.Verb, Rules: englishTemporalRules(true)},
		morph.Class{ID: "en.auxiliary.future", Language: "en", POS: lexicon.Auxiliary, Rules: []morph.Rule{lemmaCell(features.Spec{})}},
	)
	d.Specials.TimeMarkers = map[string]string{"past": "en.time.yesterday", "future": "en.time.tomorrow"}
	d.Specials.FutureAuxiliary = "en.auxiliary.will"
	d.Specials.WeatherExpletive = "en.particle.weather-it"
	d.Lexemes = append(d.Lexemes,
		lexicon.Entry{ID: "en.time.yesterday", Lemma: "yesterday", Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant"},
		lexicon.Entry{ID: "en.time.tomorrow", Lemma: "tomorrow", Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant"},
		lexicon.Entry{ID: d.Specials.FutureAuxiliary, Lemma: "will", Language: "en", POS: lexicon.Auxiliary, ClassID: "en.auxiliary.future"},
		lexicon.Entry{ID: d.Specials.WeatherExpletive, Lemma: "it", Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant"},
		lexicon.Entry{ID: "en.noun.rain", Lemma: "rain", Language: "en", POS: lexicon.Noun, ClassID: "en.noun.mass", Countability: "mass"},
		lexicon.Entry{ID: "en.verb.rain", Lemma: "rain", Language: "en", POS: lexicon.Verb, ClassID: "en.verb.temporal-regular"},
	)
	d.Senses = append(d.Senses,
		lexicon.Sense{ID: "en.sense.noun.rain", LexemeID: "en.noun.rain", Concept: core.ID("entity.rain")},
		lexicon.Sense{ID: "en.sense.verb.rain", LexemeID: "en.verb.rain", Concept: core.ID("predicate.rain"), FrameIDs: []string{"en.frame.weather"}},
	)
	for i := range d.Lexemes {
		switch d.Lexemes[i].ID {
		case "en.verb.sleep":
			d.Lexemes[i].ClassID = "en.verb.temporal-irregular"
			d.Lexemes[i].Stems = map[string]string{"past": "slept"}
		case "en.verb.run":
			d.Lexemes[i].ClassID = "en.verb.temporal-irregular"
			d.Lexemes[i].Stems = map[string]string{"past": "ran"}
		case "en.verb.read":
			d.Lexemes[i].ClassID = "en.verb.temporal-irregular"
			d.Lexemes[i].Stems = map[string]string{"past": "read"}
		case "en.verb.work", "en.verb.cook":
			d.Lexemes[i].ClassID = "en.verb.temporal-regular"
		}
	}
	d.EventFrames = append(d.EventFrames, grammar.EventFrame{ID: "en.frame.weather", SubjectRole: "theme", Weather: true, WeatherSubjectConcept: "entity.rain", AllowImplicitSubject: true})
	for _, f := range d.EventFrames {
		if f.ID != "en.frame.intr" && f.ID != "en.frame.acc" && f.ID != "en.frame.weather" {
			continue
		}
		if f.Weather {
			d.Constructions = append(d.Constructions, grammar.Pattern{ID: f.ID + ".past.expletive", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: "past", Weather: true, Analyze: true, Realize: true, Sequence: []grammar.Element{{Slot: grammar.WeatherExpletive}, {Slot: grammar.PredicateVerb}, {Slot: grammar.TemporalAdverb}}})
			continue
		}
		for _, tense := range []string{"past", "future"} {
			middle := []grammar.Element{{Slot: grammar.SubjectPhrase}}
			if tense == "future" {
				middle = append(middle, grammar.Element{Slot: grammar.FutureAuxiliary})
			}
			middle = append(middle, grammar.Element{Slot: grammar.PredicateVerb})
			if f.ObjectRole != "" {
				middle = append(middle, grammar.Element{Slot: grammar.ObjectPhrase})
			}
			prefix := append([]grammar.Element{{Slot: grammar.TemporalAdverb}}, middle...)
			suffix := append(append([]grammar.Element{}, middle...), grammar.Element{Slot: grammar.TemporalAdverb})
			d.Constructions = append(d.Constructions,
				grammar.Pattern{ID: f.ID + "." + tense + ".prefix", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: tense, Analyze: true, Realize: true, Sequence: prefix},
				grammar.Pattern{ID: f.ID + "." + tense + ".suffix", FrameID: f.ID, Event: true, Temporal: true, TemporalTense: tense, Analyze: true, Realize: true, Sequence: suffix},
			)
		}
	}
	return d
}
