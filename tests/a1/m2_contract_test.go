package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"testing"
)

func m2Request(c ContextFlags) context.Request {
	req := context.Request{Deixis: context.Unresolved}
	if c&SingularAddressee != 0 {
		req.AddresseeNumber = "singular"
	}
	if c&PluralAddressee != 0 {
		req.AddresseeNumber = "plural"
	}
	return req
}
func m2Contains(t *testing.T, r engine.Result, wantText string, want sem.Meaning, c *core.Catalog) {
	t.Helper()
	if r.Outcome != status.OK && r.Outcome != status.Ambiguous {
		t.Fatalf("target result is not supported: %+v", r)
	}
	found := false
	for _, candidate := range r.Candidates {
		same, err := sem.Equivalent(candidate.Meaning, want, c)
		if err != nil || !same {
			t.Fatalf("target changed meaning in %q: %v %+v", candidate.Text, err, candidate.Meaning)
		}
		if candidate.Text == wantText {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %q among %d target candidates: %+v", wantText, len(r.Candidates), r)
	}
}
func m2SingleMeaning(t *testing.T, r engine.Result, want sem.Meaning, c *core.Catalog) {
	t.Helper()
	if r.Outcome != status.OK || len(r.Candidates) != 1 {
		t.Fatalf("expected one parsed meaning, got %+v", r)
	}
	equal, err := sem.Equivalent(r.Candidates[0].Meaning, want, c)
	if err != nil || !equal {
		t.Fatalf("wrong parsed event meaning: %v\ngot: %+v\nwant: %+v", err, r.Candidates[0].Meaning, want)
	}
}

// Exactly 30 additional A1 situations run all six paths through the REAL
// production engine. Their two source languages have independently authored
// referring details: English a/an, Polish no article, and gender-neutral they.
func TestM2SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[20:50] {
		spec := m2Requirements[i]
		if tc.ID != spec.id {
			t.Fatalf("case mismatch: %q", tc.ID)
		}
		req := m2Request(tc.Context)
		plMeaning := authoredM2Meaning(spec, "pl")
		enMeaning := authoredM2Meaning(spec, "en")
		t.Run(tc.ID, func(t *testing.T) {
			t.Run("PL_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("pl", tc.PL, req), plMeaning, c) })
			t.Run("EN_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("en", tc.EN, req), enMeaning, c) })
			t.Run("IR_to_PL", func(t *testing.T) { m2Contains(t, e.Realize("pl", plMeaning), tc.PL, plMeaning, c) })
			t.Run("IR_to_EN", func(t *testing.T) { m2Contains(t, e.Realize("en", enMeaning), tc.EN, enMeaning, c) })
			t.Run("PL_to_EN", func(t *testing.T) {
				m2Contains(t, e.Translate("pl", "en", tc.PL, req), tc.EN, plMeaning, c)
			})
			t.Run("EN_to_PL", func(t *testing.T) {
				m2Contains(t, e.Translate("en", "pl", tc.EN, req), tc.PL, enMeaning, c)
			})
		})
	}
}

func TestM2AgreementGovernmentAndArticleContrasts(t *testing.T) {
	e, _ := newEngine(t)
	for _, tc := range []struct{ language, text string }{
		{"en", "A teacher read a book."}, {"en", "She have a bag."}, {"en", "He have a cat."}, {"en", "I has a book."},
		{"en", "A father closes door."}, {"en", "A boy eats an apple and drinks water."}, {"en", "A dog sees an cat."},
		{"en", "A girl drinks a water."}, {"en", "A teacher reads book."},
		{"pl", "Pies widzą kota."}, {"pl", "Nauczyciel czyta książki."}, {"pl", "Nauczyciel czyta książka."},
		{"pl", "Nauczyciel pomaga dziecko."}, {"pl", "Oni ma samochód."}, {"pl", "Mam książka."},
		{"pl", "Kot śpi psa."}, {"pl", "Pies widzi ptak."},
	} {
		t.Run(tc.language+"-"+tc.text, func(t *testing.T) {
			r := e.Analyze(tc.language, tc.text, context.Request{})
			if len(r.Candidates) > 0 || r.Outcome == status.OK || r.Outcome == status.Ambiguous {
				t.Fatalf("illegal construction accepted: %+v", r)
			}
		})
	}
}

func TestM2ComposesUnseenActionsAndKeepsTwoCatsDistinct(t *testing.T) {
	e, c := newEngine(t)
	for _, tc := range []struct{ pl, en string }{
		{"Pies widzi ptaka.", "A dog sees a bird."},
		{"Dziewczynka widzi psa.", "A girl sees a dog."},
		{"Nauczyciel pisze książkę.", "A teacher writes a book."},
		{"Kot widzi kota.", "A cat sees a cat."},
	} {
		t.Run(tc.pl, func(t *testing.T) {
			r := e.Translate("pl", "en", tc.pl, context.Request{})
			seen := false
			for _, x := range r.Candidates {
				if x.Text == tc.en {
					seen = true
				}
			}
			if !seen {
				t.Fatalf("missing composed translation %q: %+v", tc.en, r)
			}
		})
	}
	r := e.Analyze("pl", "Kot widzi kota.", context.Request{})
	if len(r.Candidates) != 1 {
		t.Fatalf("expected one repeated-concept analysis: %+v", r)
	}
	m := r.Candidates[0].Meaning
	if len(m.Nodes) != 3 || m.Nodes[1].ID == m.Nodes[2].ID {
		t.Fatalf("reused reference identity: %+v", m)
	}
	if _, err := sem.CanonicalKey(m, c); err != nil {
		t.Fatal(err)
	}
}

func TestM2UnresolvedYouAndArticleAssumptions(t *testing.T) {
	e, _ := newEngine(t)
	parsed := e.Analyze("en", "You have a dog.", context.Request{})
	if parsed.Outcome != status.Ambiguous || len(parsed.Candidates) != 2 {
		t.Fatalf("missing neutral English you number alternatives: %+v", parsed)
	}
	for _, candidate := range parsed.Candidates {
		if len(candidate.Assumptions) != 1 || candidate.Assumptions[0].Node != "r1" || candidate.Assumptions[0].Field != "addressee_number" || candidate.Assumptions[0].Value == "" {
			t.Fatalf("English you silently acquired a number: %+v", candidate)
		}
	}
	for _, tc := range []struct{ number, want string }{{"singular", "Masz psa."}, {"plural", "Macie psa."}} {
		req := context.Request{AddresseeNumber: tc.number}
		r := e.Translate("en", "pl", "You have a dog.", req)
		if len(r.Candidates) != 1 || r.Candidates[0].Text != tc.want {
			t.Fatalf("wrong addressee number %s: %+v", tc.number, r)
		}
	}
	pl := e.Translate("pl", "en", "Mam książkę.", context.Request{})
	if len(pl.Candidates) != 2 || pl.Outcome != status.Ambiguous {
		t.Fatalf("Polish article uncertainty suppressed: %+v", pl)
	}
	for _, cand := range pl.Candidates {
		if len(cand.Assumptions) != 1 || cand.Assumptions[0].Node != "r2" || cand.Assumptions[0].Field != "determiner" {
			t.Fatalf("unscoped article choice: %+v", cand)
		}
	}
	en := e.Translate("en", "pl", "I have a book.", context.Request{})
	if len(en.Candidates) != 1 || len(en.Candidates[0].Losses) != 1 || en.Candidates[0].Losses[0].Node != "r2" {
		t.Fatalf("English article evidence vanished: %+v", en)
	}
	doors := e.Translate("pl", "en", "Tata zamyka drzwi.", context.Request{})
	if len(doors.Candidates) != 8 {
		t.Fatalf("expected 2*2*2 article/count alternatives, got %+v", doors)
	}
	for _, candidate := range doors.Candidates {
		if candidate.Text == "A father closes door." || candidate.Text == "The father closes door." {
			t.Fatal("bare singular count noun")
		}
	}
}
