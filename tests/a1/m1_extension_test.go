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

// New semantic concept and nonstandard lexeme identifiers, but NO new
// construction, morphology rule, parser branch or full-sentence rule.
func TestNewLexemeViaDataOnly(t *testing.T) {
	base, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	defs := []core.Definition{}
	for _, id := range base.SortedIDs() {
		d, _ := base.Lookup(id)
		defs = append(defs, d)
	}
	defs = append(defs, core.Definition{ID: "entity.lamp", Kind: core.Entity})
	catalog, err := core.NewCatalog(defs)
	if err != nil {
		t.Fatal(err)
	}
	p := pl.Definition()
	p.Lexemes = append(p.Lexemes, lexicon.Entry{ID: "example.pl.object-009", Language: "pl", Lemma: "lampa", POS: lexicon.Noun, ClassID: "pl.noun.feminine.a", Gender: "feminine", Stems: map[string]string{"root": "lamp"}})
	p.Senses = append(p.Senses, lexicon.Sense{ID: "example.pl.sense-009", LexemeID: "example.pl.object-009", Concept: "entity.lamp"})
	eng := en.Definition()
	eng.Lexemes = append(eng.Lexemes, lexicon.Entry{ID: "example.en.object-009", Language: "en", Lemma: "lamp", POS: lexicon.Noun, ClassID: "en.noun.regular-s", Onset: "consonant"})
	eng.Senses = append(eng.Senses, lexicon.Sense{ID: "example.en.sense-009", LexemeID: "example.en.object-009", Concept: "entity.lamp"})
	runtime, err := engine.New(catalog, p, eng)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ source, target, input, want string }{
		{"pl", "en", "To jest lampa.", "This is a lamp."},
		{"en", "pl", "This is a lamp.", "To jest lampa."},
		{"pl", "en", "Ta lampa jest nowa.", "This lamp is new."},
		{"en", "pl", "This lamp is new.", "Ta lampa jest nowa."},
	} {
		result := runtime.Translate(tc.source, tc.target, tc.input, context.Request{Deixis: context.Proximal})
		if result.Outcome != status.OK || len(result.Candidates) != 1 || result.Candidates[0].Text != tc.want {
			t.Fatalf("failed data-only extension: %+v (want %q)", result, tc.want)
		}
	}
}

func TestUnresolvedPolishDemonstrativeDoesNotBecomeCertainEnglishThis(t *testing.T) {
	e, _ := newEngine(t)
	actual := e.Translate("pl", "en", "To jest pies.", context.Request{Deixis: context.Unresolved})
	if len(actual.Candidates) != 0 || actual.Outcome == status.OK {
		t.Fatalf("invented proximal reference: %+v", actual)
	}
}
