package en

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

// Adpositions (including multiword adpositions) are licensed lexical units.
// There is no direct Polish-to-English preposition dictionary.
func addLocation(d compile.Definition) compile.Definition {
	d.MorphClasses = append(d.MorphClasses, morph.Class{ID: "en.adposition.invariant", Language: "en", POS: lexicon.Adposition, Rules: []morph.Rule{lemmaCell(features.Spec{})}})
	for i := range d.Lexemes {
		if d.Lexemes[i].ID == "en.noun.school" {
			d.Lexemes[i].SiteClass = "institution"
		}
	}
	for _, x := range []struct{ concept, lemma string }{{"kitchen", "kitchen"}, {"tree", "tree"}, {"room", "room"}} {
		id := "en.noun." + x.concept
		entry := lexicon.Entry{ID: id, Language: "en", Lemma: x.lemma, POS: lexicon.Noun, ClassID: "en.noun.regular-s", Onset: "consonant"}
		if x.concept == "tree" {
			entry.SiteClass = "tree"
		}
		d.Lexemes = append(d.Lexemes, entry)
		d.Senses = append(d.Senses, lexicon.Sense{ID: "en.sense.noun." + x.concept, LexemeID: id, Concept: core.ID("entity." + x.concept)})
	}
	for _, x := range []struct{ key, lemma string }{
		{"in", "in"}, {"on", "on"}, {"under", "under"}, {"at", "at"}, {"in_front_of", "in front of"}, {"behind", "behind"}, {"next_to", "next to"},
	} {
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: "en.adposition." + x.key, Language: "en", Lemma: x.lemma, POS: lexicon.Adposition, ClassID: "en.adposition.invariant"})
	}
	for _, f := range []struct {
		relation, siteClass, adposition string
		bare                            bool
	}{
		{"inside", "general", "in", false},
		{"on_support", "general", "on", false},
		{"below", "general", "under", false},
		{"at_site", "institution", "at", true},
		{"at_site", "tree", "in", false},
		{"in_front", "general", "in_front_of", false},
		{"behind", "general", "behind", false},
		{"adjacent", "general", "next_to", false},
	} {
		frameID := "en.location." + f.relation + "." + f.siteClass
		d.LocationFrames = append(d.LocationFrames, grammar.LocationFrame{
			ID: frameID, Predicate: core.ID("relation." + f.relation), LocatedRole: "located", SiteRole: "site", SiteClass: f.siteClass,
			SiteCase: "nominative", AdpositionID: "en.adposition." + f.adposition, BareSite: f.bare,
		})
		d.Constructions = append(d.Constructions, grammar.Pattern{
			ID: frameID + ".clause", FrameID: frameID, Location: true, Analyze: true, Realize: true,
			Sequence: []grammar.Element{{Slot: grammar.SubjectPhrase}, {Slot: grammar.Copula}, {Slot: grammar.AdpositionPhrase}, {Slot: grammar.LocationSitePhrase}},
		})
	}
	return d
}
