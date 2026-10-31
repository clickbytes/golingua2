package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/status"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

// A new predicate+verb in BOTH languages needs only CORE, lexical and sense
// data. No new parser branch, output template, morphology class or grammar
// declaration is allowed. This example is NOT one of the 100 A1 sentences.
func TestM2NewPredicateViaDataOnly(t *testing.T) {
	original, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	definitions := append(original.Definitions(), core.Definition{ID: "predicate.observe", Kind: core.Predicate, Roles: []core.Role{{Name: "agent", Target: core.Entity}, {Name: "patient", Target: core.Entity}}})
	catalog, err := core.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	polish, english := pl.Definition(), en.Definition()
	polish.Lexemes = append(polish.Lexemes, lexicon.Entry{ID: "pl.verb.observe", Language: "pl", POS: lexicon.Verb, Lemma: "obserwować", ClassID: "pl.verb.present-e", Stems: map[string]string{"root": "obserwuj"}})
	polish.Senses = append(polish.Senses, lexicon.Sense{ID: "pl.sense.verb.observe", LexemeID: "pl.verb.observe", Concept: "predicate.observe", FrameIDs: []string{"pl.frame.acc"}})
	english.Lexemes = append(english.Lexemes, lexicon.Entry{ID: "en.verb.observe", Language: "en", POS: lexicon.Verb, Lemma: "observe", ClassID: "en.verb.present-s"})
	english.Senses = append(english.Senses, lexicon.Sense{ID: "en.sense.verb.observe", LexemeID: "en.verb.observe", Concept: "predicate.observe", FrameIDs: []string{"en.frame.acc"}})
	e, err := engine.New(catalog, polish, english)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ source, target, input, want string }{
		{"pl", "en", "Nauczyciel obserwuje kota.", "A teacher observes a cat."},
		{"en", "pl", "A teacher observes a cat.", "Nauczyciel obserwuje kota."},
	} {
		r := e.Translate(tc.source, tc.target, tc.input, context.Request{})
		if r.Outcome != status.Ambiguous && r.Outcome != status.OK {
			t.Fatalf("new predicate unavailable: %+v", r)
		}
		found := false
		for _, cand := range r.Candidates {
			if cand.Text == tc.want {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing new regular verb output %q: %+v", tc.want, r)
		}
	}
}
