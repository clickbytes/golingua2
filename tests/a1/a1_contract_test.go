package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

func newEngine(t *testing.T) (*engine.Engine, *core.Catalog) {
	t.Helper()
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	e, err := engine.New(c, pl.Definition(), en.Definition())
	if err != nil {
		t.Fatal(err)
	}
	return e, c
}

func requireMeaning(t *testing.T, r engine.Result, want sem.Meaning, c *core.Catalog) {
	t.Helper()
	if r.Outcome != status.OK || len(r.Candidates) != 1 {
		t.Fatalf("unexpected result: %+v", r)
	}
	equal, err := sem.Equivalent(r.Candidates[0].Meaning, want, c)
	if err != nil || !equal {
		t.Fatalf("meaning differs: %+v vs %+v, err %v", r.Candidates[0].Meaning, want, err)
	}
}
func requireText(t *testing.T, r engine.Result, example string, c *core.Catalog, want sem.Meaning) {
	t.Helper()
	if r.Outcome != status.OK || len(r.Candidates) != 1 {
		t.Fatalf("expected one licensed realization: %+v", r)
	}
	if r.Candidates[0].Text != example {
		t.Fatalf("expected %q, got %q", example, r.Candidates[0].Text)
	}
	equal, err := sem.Equivalent(r.Candidates[0].Meaning, want, c)
	if err != nil || !equal {
		t.Fatalf("target changed semantic meaning: %+v (%v)", r.Candidates[0].Meaning, err)
	}
}

// Exactly 20 cases are currently implemented. The remaining 80 have inventory
// checks but NO synthetic PASS and NO placeholder translation assertions.
func TestM1SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	req := context.Request{Deixis: context.Proximal}
	for i, tc := range cases[:20] {
		spec := m1MeaningRequirements[i]
		if tc.ID != spec.id {
			t.Fatalf("fixture/index mismatch: %s vs %s", tc.ID, spec.id)
		}
		want := authoredM1Meaning(spec.kind, spec.entity, spec.property)
		t.Run(tc.ID, func(t *testing.T) {
			t.Run("PL_to_IR", func(t *testing.T) { requireMeaning(t, e.Analyze("pl", tc.PL, req), want, c) })
			t.Run("EN_to_IR", func(t *testing.T) { requireMeaning(t, e.Analyze("en", tc.EN, req), want, c) })
			t.Run("IR_to_PL", func(t *testing.T) { requireText(t, e.Realize("pl", want), tc.PL, c, want) })
			t.Run("IR_to_EN", func(t *testing.T) { requireText(t, e.Realize("en", want), tc.EN, c, want) })
			t.Run("PL_to_EN", func(t *testing.T) { requireText(t, e.Translate("pl", "en", tc.PL, req), tc.EN, c, want) })
			t.Run("EN_to_PL", func(t *testing.T) { requireText(t, e.Translate("en", "pl", tc.EN, req), tc.PL, c, want) })
		})
	}
}

func TestM1CompositionNotSentenceLookup(t *testing.T) {
	e, c := newEngine(t)
	req := context.Request{Deixis: context.Proximal}
	// These combinations are not in the authored 100-pair catalog.
	for _, tc := range []struct{ pl, en, entity, property string }{
		{"Ten pies jest czarny.", "This dog is black.", "dog", "black"},
		{"Ta książka jest czerwona.", "This book is red.", "book", "red"},
		{"To krzesło jest duże.", "This chair is big.", "chair", "big"},
	} {
		want := authoredM1Meaning("property", tc.entity, tc.property)
		t.Run(tc.entity+"-"+tc.property, func(t *testing.T) {
			requireMeaning(t, e.Analyze("pl", tc.pl, req), want, c)
			requireMeaning(t, e.Analyze("en", tc.en, req), want, c)
			requireText(t, e.Translate("pl", "en", tc.pl, req), tc.en, c, want)
		})
	}
}

func TestM1RejectsUnlicensedOrIncorrectAgreement(t *testing.T) {
	e, _ := newEngine(t)
	req := context.Request{Deixis: context.Proximal}
	for _, tc := range []struct{ lang, text string }{
		{"en", "This is a apple."}, {"en", "This is an cat."}, {"en", "This dog are black."},
		{"en", "This dog is blacks."}, {"en", "This cat is black and small."},
		{"pl", "Ta kot jest czarny."}, {"pl", "Ten kot jest czarna."},
		{"pl", "To kot jest czarny."}, {"pl", "To jest kota."},
		{"pl", "Ten pies jest mała."},
	} {
		t.Run(tc.lang+"-"+tc.text, func(t *testing.T) {
			r := e.Analyze(tc.lang, tc.text, req)
			if len(r.Candidates) != 0 || r.Outcome == status.OK {
				t.Fatalf("unlicensed sentence accepted: %+v", r)
			}
		})
	}
}
