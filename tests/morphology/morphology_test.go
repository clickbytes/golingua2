package morphology_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

func pack(t *testing.T, lang string) *compile.Pack {
	t.Helper()
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	def := pl.Definition()
	if lang == "en" {
		def = en.Definition()
	}
	out, err := compile.Compile(def, c)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestProceduralMorphologyAndInverse(t *testing.T) {
	for _, tc := range []struct {
		lang, lexeme, surface string
		f                     features.Spec
	}{
		{"pl", "pl.noun.cat", "kot", features.Spec{Number: "singular", Case: "nominative"}},
		{"pl", "pl.noun.cat", "kota", features.Spec{Number: "singular", Case: "accusative"}},
		{"pl", "pl.noun.dog", "psa", features.Spec{Number: "singular", Case: "genitive"}},
		{"pl", "pl.noun.book", "książki", features.Spec{Number: "singular", Case: "genitive"}},
		{"pl", "pl.noun.book", "książki", features.Spec{Number: "plural", Case: "nominative"}},
		{"pl", "pl.noun.book", "książce", features.Spec{Number: "singular", Case: "dative"}},
		{"pl", "pl.noun.school", "szkoły", features.Spec{Number: "singular", Case: "genitive"}},
		{"pl", "pl.noun.chair", "krzesła", features.Spec{Number: "plural", Case: "nominative"}},
		{"pl", "pl.adjective.black", "czarny", features.Spec{Number: "singular", Case: "nominative", Gender: "masculine", Degree: "positive"}},
		{"pl", "pl.adjective.black", "czarna", features.Spec{Number: "singular", Case: "nominative", Gender: "feminine", Degree: "positive"}},
		{"pl", "pl.adjective.black", "czarne", features.Spec{Number: "singular", Case: "nominative", Gender: "neuter", Degree: "positive"}},
		{"pl", "pl.verb.byc", "jest", features.Spec{Number: "singular", Person: "third", Tense: "present"}},
		{"en", "en.noun.apple", "apple", features.Spec{Number: "singular", Case: "nominative"}},
		{"en", "en.noun.apple", "apples", features.Spec{Number: "plural", Case: "nominative"}},
		{"en", "en.article.a", "an", features.Spec{Onset: "vowel"}},
		{"en", "en.article.a", "a", features.Spec{Onset: "consonant"}},
		{"en", "en.adjective.new", "new", features.Spec{Degree: "positive"}},
		{"en", "en.verb.be", "is", features.Spec{Number: "singular", Person: "third", Tense: "present"}},
	} {
		t.Run(tc.lang+"-"+tc.lexeme+"-"+tc.f.Key(), func(t *testing.T) {
			p := pack(t, tc.lang)
			surface, err := p.Morph.Generate(tc.lexeme, tc.f)
			if err != nil || surface != tc.surface {
				t.Fatalf("got %q, %v, want %q", surface, err, tc.surface)
			}
			found := false
			for _, reading := range p.Morph.Analyze(surface) {
				if reading.LexemeID == tc.lexeme && reading.Features == tc.f {
					found = true
				}
			}
			if !found {
				t.Fatalf("inverse morphology lost licensed reading for %s", tc.lexeme)
			}
		})
	}
}

func TestSyncretismRetainsDistinctReadings(t *testing.T) {
	p := pack(t, "pl")
	seen := map[string]bool{}
	for _, r := range p.Morph.Analyze("książki") {
		if r.LexemeID == "pl.noun.book" {
			seen[r.Features.Key()] = true
		}
	}
	for _, want := range []features.Spec{{Number: "singular", Case: "genitive"}, {Number: "plural", Case: "nominative"}} {
		if !seen[want.Key()] {
			t.Fatalf("missing syncretic reading %+v", want)
		}
	}
	if _, err := p.Morph.Generate("pl.noun.book", features.Spec{Number: "singular", Case: "locative"}); err == nil {
		t.Fatal("undeclared case was guessed")
	}
}
