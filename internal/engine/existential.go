package engine

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"sort"
	"strings"
)

// Existential is a spatial relation with a quantified located entity. Its
// language-specific existential syntax does not change the spatial predicate.
func (e *Engine) analyzeExistentialPattern(p *compile.Pack, pat grammar.Pattern, tokens []string, req context.Request) []Candidate {
	frame, ok := p.LocationFrame(pat.FrameID)
	if !ok {
		return nil
	}
	var results []Candidate
	if p.Language == "en" {
		if len(tokens) < 6 || !markerLicensed(p, tokens[0], p.Specials.ExistentialMarker) {
			return nil
		}
		subjects := e.parseNumberedNP(p, tokens[2:4], "nominative")
		for adpLen := 1; adpLen <= 3 && 4+adpLen < len(tokens); adpLen++ {
			if !markerLicensed(p, strings.Join(tokens[4:4+adpLen], " "), frame.AdpositionID) {
				continue
			}
			sites := e.parseNounPhraseMode(p, tokens[4+adpLen:], frame.SiteCase, req, false, frame.BareSite)
			results = append(results, e.buildExistentialCandidates(p, frame, tokens[1], subjects, sites)...)
		}
	} else {
		if len(tokens) < 5 {
			return nil
		}
		for adpLen := 1; adpLen <= 2; adpLen++ {
			if !markerLicensed(p, strings.Join(tokens[:adpLen], " "), frame.AdpositionID) {
				continue
			}
			for siteLen := 1; siteLen <= 2; siteLen++ {
				index := adpLen + siteLen
				if index+3 != len(tokens) {
					continue
				}
				sites := e.parseNounPhraseMode(p, tokens[adpLen:index], frame.SiteCase, req, false, frame.BareSite)
				subjects := e.parseNumberedNP(p, tokens[index+1:], "nominative")
				results = append(results, e.buildExistentialCandidates(p, frame, tokens[index], subjects, sites)...)
			}
		}
	}
	return results
}

func (e *Engine) buildExistentialCandidates(p *compile.Pack, frame grammar.LocationFrame, copula string, subjects, sites []nounReading) []Candidate {
	var result []Candidate
	for _, sub := range subjects {
		for _, site := range sites {
			if sub.ref.Cardinality < 1 || sub.lex.POS != lexicon.Noun || site.lex.POS != lexicon.Noun || site.ref.Cardinality > 0 || site.ref.Number != "singular" || lexicalSiteClass(site.lex) != frame.SiteClass {
				continue
			}
			finite := false
			for _, reading := range p.Morph.Analyze(copula) {
				if reading.LexemeID == p.Specials.Copula && reading.Features.Person == "third" && reading.Features.Tense == "present" && reading.Features.Number == sub.ref.Number {
					finite = true
					break
				}
			}
			if !finite {
				continue
			}
			located, place := sub.ref, site.ref
			located.ID = "r1"
			place.ID = "r2"
			predicate := sem.Node{ID: "l1", Kind: sem.State, Concept: frame.Predicate, Existential: true, Arguments: []sem.Argument{{Role: frame.LocatedRole, Target: located.ID}, {Role: frame.SiteRole, Target: place.ID}}}
			m := sem.Meaning{Root: predicate.ID, Nodes: []sem.Node{predicate, located, place}}
			if sem.Validate(m, e.catalog) != nil {
				continue
			}
			result = append(result, Candidate{Meaning: m})
		}
	}
	return result
}

func (e *Engine) realizeExistential(p *compile.Pack, m sem.Meaning, root sem.NodeID) Result {
	byID := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	relation := byID[root]
	produced := map[string]Candidate{}
	for _, pat := range p.Constructions {
		if !pat.Location || !pat.Existential || !pat.Realize {
			continue
		}
		frame, ok := p.LocationFrame(pat.FrameID)
		if !ok || frame.Predicate != relation.Concept {
			continue
		}
		args := map[string]sem.Node{}
		for _, a := range relation.Arguments {
			args[a.Role] = byID[a.Target]
		}
		sub, ok := args[frame.LocatedRole]
		if !ok || sub.Cardinality < 1 {
			continue
		}
		site, ok := args[frame.SiteRole]
		if !ok || site.Number != "singular" {
			continue
		}
		subjects := e.renderNumberedNP(p, sub, "nominative")
		sites := e.renderLocationSite(p, site, frame)
		copula, err := p.Morph.Generate(p.Specials.Copula, features.Spec{Tense: "present", Person: "third", Number: sub.Number})
		if err != nil {
			continue
		}
		adp, err := p.Morph.Generate(frame.AdpositionID, features.Spec{})
		if err != nil {
			continue
		}
		marker := ""
		if p.Specials.ExistentialMarker != "" {
			marker, err = p.Morph.Generate(p.Specials.ExistentialMarker, features.Spec{})
			if err != nil {
				continue
			}
		}
		for _, noun := range subjects {
			for _, loc := range sites {
				parts := map[grammar.Slot]string{grammar.SubjectPhrase: noun.text, grammar.Copula: copula, grammar.AdpositionPhrase: adp, grammar.LocationSitePhrase: loc.text, grammar.ExistentialMarker: marker}
				tokens := []string{}
				for _, el := range pat.Sequence {
					tokens = append(tokens, parts[el.Slot])
				}
				surface := capitalize(strings.Join(tokens, " ")) + "."
				produced[surface] = Candidate{Meaning: m.Clone(), Text: surface, Assumptions: append([]Annotation(nil), loc.assumptions...), Losses: append([]Annotation(nil), loc.losses...)}
				if len(produced) > maxCandidates {
					return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "existential alternatives exceeded bound")
				}
			}
		}
	}
	if len(produced) == 0 {
		return fail(status.Unsupported, "EXISTENTIAL_CONSTRUCTION_UNAVAILABLE", "quantified existence lacks an exact licensed construction")
	}
	keys := make([]string, 0, len(produced))
	for k := range produced {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	r := Result{Outcome: status.OK}
	for _, k := range keys {
		r.Candidates = append(r.Candidates, produced[k])
	}
	if len(keys) > 1 {
		r.Outcome = status.Ambiguous
	}
	return r
}
