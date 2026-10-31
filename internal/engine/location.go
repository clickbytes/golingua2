package engine

import (
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
)

// An unspecified site class is the general lexical class. Institution and
// tree are declared language-local selection classes, not surface exceptions.
func lexicalSiteClass(lex lexicon.Entry) string {
	if lex.SiteClass == "" {
		return "general"
	}
	return lex.SiteClass
}

// analyzeLocationPattern consumes independently recognized subject NP, finite
// copula, possibly multiword adposition, and site NP. The frame (not the
// adposition spelling) provides the neutral relation and argument roles.
func (e *Engine) analyzeLocationPattern(p *compile.Pack, pat grammar.Pattern, tokens []string, req context.Request) []Candidate {
	frame, ok := p.LocationFrame(pat.FrameID)
	if !ok {
		return nil
	}
	var out []Candidate
	for subjLen := 1; subjLen <= 2; subjLen++ {
		if subjLen >= len(tokens) {
			continue
		}
		copulaIndex := subjLen
		finiteCopula := false
		for _, reading := range p.Morph.Analyze(tokens[copulaIndex]) {
			if reading.LexemeID == p.Specials.Copula && reading.Features.Tense == "present" && reading.Features.Person == "third" && reading.Features.Number == "singular" {
				finiteCopula = true
				break
			}
		}
		if !finiteCopula {
			continue
		}
		subjects := e.parseNounPhrase(p, tokens[:subjLen], "nominative", req, true)
		if len(subjects) == 0 {
			continue
		}
		for adpLen := 1; adpLen <= 3; adpLen++ {
			adpIndex := copulaIndex + 1
			siteIndex := adpIndex + adpLen
			if siteIndex >= len(tokens) {
				continue
			}
			if !markerLicensed(p, strings.Join(tokens[adpIndex:siteIndex], " "), frame.AdpositionID) {
				continue
			}
			if frame.BareSite && len(tokens[siteIndex:]) != 1 {
				continue
			}
			sites := e.parseNounPhraseMode(p, tokens[siteIndex:], frame.SiteCase, req, false, frame.BareSite)
			for _, located := range subjects {
				// This first location slice deliberately licenses nominal,
				// singular third-person statements, not arbitrary pronouns.
				if located.lex.POS != lexicon.Noun || located.ref.Number != "singular" {
					continue
				}
				for _, site := range sites {
					if site.lex.POS != lexicon.Noun || lexicalSiteClass(site.lex) != frame.SiteClass || site.ref.Number != "singular" {
						continue
					}
					relation := sem.Node{ID: "l1", Kind: sem.State, Concept: frame.Predicate, Arguments: []sem.Argument{
						{Role: frame.LocatedRole, Target: "r1"}, {Role: frame.SiteRole, Target: "r2"},
					}}
					a, b := located.ref, site.ref
					a.ID = "r1"
					b.ID = "r2"
					meaning := sem.Meaning{Root: relation.ID, Nodes: []sem.Node{relation, a, b}}
					if err := sem.Validate(meaning, e.catalog); err != nil {
						continue
					}
					out = append(out, Candidate{Meaning: meaning, Evidence: []Evidence{
						{Token: 0, Lexeme: located.lex.ID, Sense: located.sense.ID, Node: a.ID},
						{Token: copulaIndex, Lexeme: p.Specials.Copula, Node: relation.ID},
						{Token: adpIndex, Lexeme: frame.AdpositionID, Node: relation.ID},
						{Token: siteIndex, Lexeme: site.lex.ID, Sense: site.sense.ID, Node: b.ID},
					}})
					if len(out) > maxCandidates {
						return nil
					}
				}
			}
		}
	}
	return out
}

func (e *Engine) renderLocationSite(p *compile.Pack, ref sem.Node, frame grammar.LocationFrame) []phraseOption {
	if !frame.BareSite {
		return e.renderEventNPClass(p, ref, frame.SiteCase, frame.SiteClass)
	}
	if ref.Person != "" || ref.Definiteness != "" || ref.Number != "singular" {
		return nil
	}
	var out []phraseOption
	for _, lex := range selectLexemes(p, ref.Concept, lexicon.Noun) {
		if lexicalSiteClass(lex) != frame.SiteClass || lex.Countability == "mass" || lex.Countability == "plural_only" {
			continue
		}
		text, err := p.Morph.Generate(lex.ID, features.Spec{Number: "singular", Case: frame.SiteCase})
		if err == nil {
			out = append(out, phraseOption{text: text, grammaticalNumber: "singular"})
		}
	}
	return out
}

func (e *Engine) realizeLocation(p *compile.Pack, m sem.Meaning, root sem.NodeID) Result {
	byID := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	relation := byID[root]
	produced := map[string]Candidate{}
	for _, pat := range p.Constructions {
		if !pat.Location || pat.Existential || !pat.Realize {
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
		subject, hasSubject := args[frame.LocatedRole]
		site, hasSite := args[frame.SiteRole]
		if !hasSubject || !hasSite || len(args) != 2 || subject.Kind != sem.Referent || site.Kind != sem.Referent || subject.Person != "" || site.Person != "" || subject.Number != "singular" || site.Number != "singular" {
			continue
		}
		subjects := e.renderEventNP(p, subject, "nominative")
		sites := e.renderLocationSite(p, site, frame)
		adposition, err := p.Morph.Generate(frame.AdpositionID, features.Spec{})
		if err != nil {
			continue
		}
		copula, err := p.Morph.Generate(p.Specials.Copula, features.Spec{Tense: "present", Person: "third", Number: "singular"})
		if err != nil {
			continue
		}
		for _, sub := range subjects {
			for _, loc := range sites {
				if sub.text == "" || loc.text == "" {
					continue
				}
				parts := map[grammar.Slot]string{grammar.SubjectPhrase: sub.text, grammar.Copula: copula, grammar.AdpositionPhrase: adposition, grammar.LocationSitePhrase: loc.text}
				tokens := make([]string, 0, len(pat.Sequence))
				for _, el := range pat.Sequence {
					tokens = append(tokens, parts[el.Slot])
				}
				sentence := capitalize(strings.Join(tokens, " ")) + "."
				candidate := Candidate{Meaning: m.Clone(), Text: sentence}
				candidate.Assumptions = append(append([]Annotation{}, sub.assumptions...), loc.assumptions...)
				candidate.Losses = append(append([]Annotation{}, sub.losses...), loc.losses...)
				if _, exists := produced[sentence]; !exists {
					produced[sentence] = candidate
				}
				if len(produced) > maxCandidates {
					return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "location realization candidate bound exceeded")
				}
			}
		}
	}
	if len(produced) == 0 {
		return fail(status.Unsupported, "LOCATION_CONSTRUCTION_UNAVAILABLE", "no approved spatial construction covers the complete meaning")
	}
	keys := make([]string, 0, len(produced))
	for key := range produced {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := Result{Outcome: status.OK}
	for _, key := range keys {
		result.Candidates = append(result.Candidates, produced[key])
	}
	if len(keys) > 1 {
		result.Outcome = status.Ambiguous
	}
	return result
}
