package a1_test

import (
	"strings"
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
)

func TestM4SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[70:80] {
		s := m4Requirements[i]
		if s.id != tc.ID {
			t.Fatalf("mismatched requirement %s", tc.ID)
		}
		pl := authoredM4Meaning(s, "pl")
		en := authoredM4Meaning(s, "en")
		req := context.Request{}
		t.Run(tc.ID, func(t *testing.T) {
			t.Run("PL_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("pl", tc.PL, req), pl, c) })
			t.Run("EN_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("en", tc.EN, req), en, c) })
			t.Run("IR_to_PL", func(t *testing.T) { m2Contains(t, e.Realize("pl", pl), tc.PL, pl, c) })
			t.Run("IR_to_EN", func(t *testing.T) { m2Contains(t, e.Realize("en", en), tc.EN, en, c) })
			t.Run("PL_to_EN", func(t *testing.T) { m2Contains(t, e.Translate("pl", "en", tc.PL, req), tc.EN, pl, c) })
			t.Run("EN_to_PL", func(t *testing.T) { m2Contains(t, e.Translate("en", "pl", tc.EN, req), tc.PL, en, c) })
		})
	}
}

func TestM4LocationContrastsAndUnseenCompositions(t *testing.T) {
	e, c := newEngine(t)
	invalid := []struct{ language, text string }{
		{"pl", "Kot jest w dom."},
		{"pl", "Książka jest na stół."},
		{"pl", "Pies jest pod stole."},
		{"pl", "Uczeń jest w szkoła."},
		{"pl", "Dziecko jest obok mamą."},
		{"pl", "Rower jest przed domu."},
		{"pl", "Samochód jest za domu."},
		{"pl", "Ptak jest na drzewo."},
		{"pl", "Stół jest w pokoja."},
		{"en", "The cat is in house."},
		{"en", "The dog is under table."},
		{"en", "The student is at the school."}, // bare institutional frame only
		{"en", "The bird is on the tree."},      // distinct on_support reading not licensed for tree in M4
		{"en", "The cat is under a dog in the house."},
		{"en", "The cat is beside the house."},
		{"en", "The cat is in front the house."},
		{"en", "The dog are under the table."},
	}
	for _, tc := range invalid {
		t.Run(tc.language+"/"+tc.text, func(t *testing.T) {
			got := e.Analyze(tc.language, tc.text, context.Request{})
			if len(got.Candidates) != 0 || got.Outcome == status.OK || got.Outcome == status.Ambiguous {
				t.Fatalf("invalid relation accepted: %+v", got)
			}
		})
	}
	unseen := []struct{ source, target, original, want string }{
		{"pl", "en", "Kot jest pod stołem.", "A cat is under a table."},
		{"pl", "en", "Pies jest w domu.", "A dog is in a house."},
		{"pl", "en", "Książka jest za domem.", "A book is behind a house."},
		{"en", "pl", "A cat is next to a dog.", "Kot jest obok psa."},
	}
	for _, tc := range unseen {
		t.Run(tc.original, func(t *testing.T) {
			result := e.Translate(tc.source, tc.target, tc.original, context.Request{})
			seen := false
			for _, candidate := range result.Candidates {
				if candidate.Text == tc.want {
					seen = true
				}
			}
			if !seen {
				t.Fatalf("unseen composition %q missing: %+v", tc.want, result)
			}
		})
	}
	a := authoredM4Meaning(m4Requirements[1], "pl")
	b := authoredM4Meaning(m4Requirements[2], "pl")
	equal, err := sem.Equivalent(a, b, c)
	if err != nil || equal {
		t.Fatalf("on_support and below were conflated: %v", err)
	}
	broken := authoredM4Meaning(m4Requirements[1], "en")
	for i := range broken.Nodes {
		if broken.Nodes[i].Kind == sem.State {
			broken.Nodes[i].Arguments = broken.Nodes[i].Arguments[:1]
		}
	}
	if sem.Validate(broken, c) == nil {
		t.Fatal("missing site argument accepted")
	}
	if r := e.Realize("pl", broken); r.Outcome != status.Invalid || len(r.Candidates) != 0 {
		t.Fatalf("partial spatial relation produced text: %+v", r)
	}
	broken = authoredM4Meaning(m4Requirements[1], "en")
	for i := range broken.Nodes {
		if broken.Nodes[i].Kind == sem.State {
			broken.Nodes[i].Arguments[1].Target = "independent.located"
		}
	}
	if sem.Validate(broken, c) != nil { /* same referent in both roles can be a valid claim; it must not fabricate a different entity */
	}
	wrong := e.Realize("en", authoredM4Meaning(m4Requirements[1], "pl"))
	for _, candidate := range wrong.Candidates {
		if strings.Contains(candidate.Text, "on table.") {
			t.Fatal("English bare singular site generated")
		}
	}
}

func TestM4ReferentScopedDeterminersAndInstitutionalBareSite(t *testing.T) {
	e, _ := newEngine(t)
	fromPolish := e.Translate("pl", "en", "Kot jest w domu.", context.Request{})
	if len(fromPolish.Candidates) != 4 || fromPolish.Outcome != status.Ambiguous {
		t.Fatalf("want four locally scoped English article choices: %+v", fromPolish)
	}
	for _, cand := range fromPolish.Candidates {
		if len(cand.Assumptions) != 2 || cand.Assumptions[0].Field != "determiner" || cand.Assumptions[1].Field != "determiner" || cand.Assumptions[0].Node == cand.Assumptions[1].Node {
			t.Fatalf("missing two independently anchored article decisions: %+v", cand)
		}
	}
	fromEnglish := e.Translate("en", "pl", "The cat is in the house.", context.Request{})
	if len(fromEnglish.Candidates) != 1 || len(fromEnglish.Candidates[0].Losses) != 2 {
		t.Fatalf("expected two local article losses: %+v", fromEnglish)
	}
	seen := map[sem.NodeID]bool{}
	for _, l := range fromEnglish.Candidates[0].Losses {
		if l.Field != "determiner" || l.Value != "definite" || seen[l.Node] {
			t.Fatalf("unexpected article loss: %+v", fromEnglish.Candidates[0].Losses)
		}
		seen[l.Node] = true
	}
	institutional := e.Translate("pl", "en", "Uczeń jest w szkole.", context.Request{})
	if len(institutional.Candidates) != 2 {
		t.Fatalf("English bare institution must not invent an article: %+v", institutional)
	}
	for _, cand := range institutional.Candidates {
		if !strings.HasSuffix(cand.Text, "at school.") || len(cand.Assumptions) != 1 || cand.Assumptions[0].Node != "r1" {
			t.Fatalf("institutional bare site acquired invented determiner: %+v", cand)
		}
	}
	school := e.Translate("en", "pl", "The student is at school.", context.Request{})
	if len(school.Candidates) != 1 || len(school.Candidates[0].Losses) != 1 || school.Candidates[0].Losses[0].Node != "r1" {
		t.Fatalf("institutional bare site acquired source article loss: %+v", school)
	}
}
