// Package engine executes shared compiled language contracts through one Meaning IR.
package engine

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
)

const (
	maxCandidates    = 128
	maxAnalysisSteps = 4096
)

type Engine struct {
	catalog *core.Catalog
	packs   map[string]*compile.Pack
}
type Evidence struct {
	Token  int        `json:"token"`
	Lexeme string     `json:"lexeme"`
	Sense  string     `json:"sense,omitempty"`
	Node   sem.NodeID `json:"node,omitempty"`
}
type Candidate struct {
	Meaning     sem.Meaning  `json:"meaning"`
	Text        string       `json:"text,omitempty"`
	Evidence    []Evidence   `json:"evidence,omitempty"`
	Assumptions []Annotation `json:"assumptions,omitempty"`
	Losses      []Annotation `json:"losses,omitempty"`
}

// Annotation belongs to a specific occurrence, not the shared concept.
type Annotation struct {
	Node  sem.NodeID `json:"node"`
	Field string     `json:"field"`
	Value string     `json:"value,omitempty"`
}
type Result struct {
	Outcome     status.Outcome      `json:"outcome"`
	Candidates  []Candidate         `json:"candidates"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

func New(catalog *core.Catalog, definitions ...compile.Definition) (*Engine, error) {
	if catalog == nil || len(definitions) == 0 {
		return nil, fmt.Errorf("missing ontology or language packages")
	}
	e := &Engine{catalog: catalog, packs: make(map[string]*compile.Pack, len(definitions))}
	for _, d := range definitions {
		if _, ok := e.packs[d.Language]; ok {
			return nil, fmt.Errorf("duplicate language %q", d.Language)
		}
		p, err := compile.Compile(d, catalog)
		if err != nil {
			return nil, fmt.Errorf("compile %s: %w", d.Language, err)
		}
		e.packs[p.Language] = p
	}
	return e, nil
}

func fail(outcome status.Outcome, code, message string) Result {
	return Result{Outcome: outcome, Diagnostics: []status.Diagnostic{{Code: code, Message: message}}}
}
func (e *Engine) PackForms(language string) int {
	if p := e.packs[language]; p != nil {
		return p.Morph.CountForms()
	}
	return 0
}

func words(input string) ([]string, error) {
	s := strings.TrimSpace(input)
	if !strings.HasSuffix(s, ".") {
		return nil, fmt.Errorf("declarative sentence must end with a period")
	}
	s = strings.TrimSuffix(s, ".")
	if strings.ContainsAny(s, ".!?;:\n\t") {
		return nil, fmt.Errorf("unsupported sentence punctuation")
	}
	tokens := strings.Fields(strings.ToLower(s))
	if len(tokens) == 0 || len(tokens) > 16 {
		return nil, fmt.Errorf("invalid token count")
	}
	for _, t := range tokens {
		if t == "" {
			return nil, fmt.Errorf("empty token")
		}
	}
	return tokens, nil
}

type match struct {
	entry lexicon.Entry
	sense lexicon.Sense
	form  morph.Analysis
}

func (e *Engine) possible(p *compile.Pack, slot grammar.Slot, token string) []match {
	matches := []match{}
	for _, a := range p.Morph.Analyze(token) {
		lex, ok := p.Lexicon.Lookup(a.LexemeID)
		if !ok {
			continue
		}
		var required lexicon.POS
		var special string
		switch slot {
		case grammar.Noun:
			required = lexicon.Noun
		case grammar.Adjective:
			required = lexicon.Adjective
		case grammar.Demonstrative:
			required = lexicon.Determiner
			special = p.Specials.Demonstrative
		case grammar.Copula:
			required = lexicon.Verb
			special = p.Specials.Copula
		case grammar.Article:
			required = lexicon.Article
			special = p.Specials.Article
		default:
			continue
		}
		if lex.POS != required || (special != "" && lex.ID != special) {
			continue
		}
		if (slot == grammar.Demonstrative || slot == grammar.Copula || slot == grammar.Article) && special == "" {
			continue
		}
		if slot == grammar.Noun && (a.Features.Number != "singular" || a.Features.Case != "nominative") {
			continue
		}
		if slot == grammar.Adjective && a.Features.Degree != "positive" {
			continue
		}
		if slot == grammar.Copula && (a.Features.Person != "third" || a.Features.Number != "singular" || a.Features.Tense != "present") {
			continue
		}
		if slot == grammar.Demonstrative && a.Features.Number != "singular" {
			continue
		}
		if slot == grammar.Noun || slot == grammar.Adjective {
			for _, sense := range p.Lexicon.SensesFor(lex.ID) {
				matches = append(matches, match{entry: lex, sense: sense, form: a})
			}
		} else {
			matches = append(matches, match{entry: lex, form: a})
		}
	}
	return matches
}

func compatible(p *compile.Pack, pat grammar.Pattern, selected map[grammar.Slot]match, req context.Request) bool {
	noun, ok := selected[grammar.Noun]
	if !ok {
		return false
	}
	demo, ok := selected[grammar.Demonstrative]
	if !ok {
		return false
	}
	_, ok = selected[grammar.Copula]
	if !ok {
		return false
	}
	if p.Specials.DemoDenotesProximal && req.Deixis == context.Distal {
		return false
	}
	switch pat.DemoGender {
	case "fixed_neuter":
		if demo.form.Features.Gender != "neuter" {
			return false
		}
	case "agree":
		if demo.form.Features.Gender != "" && demo.form.Features.Gender != noun.entry.Gender {
			return false
		}
	case "invariant":
		if demo.form.Features.Gender != "" {
			return false
		}
	}
	if adj, ok := selected[grammar.Adjective]; ok {
		if adj.form.Features.Gender != "" && adj.form.Features.Gender != noun.entry.Gender {
			return false
		}
		if adj.form.Features.Number != "" && adj.form.Features.Number != "singular" {
			return false
		}
		if adj.form.Features.Case != "" && adj.form.Features.Case != "nominative" {
			return false
		}
	}
	if article, ok := selected[grammar.Article]; ok {
		if noun.entry.Onset == "" || article.form.Features.Onset != noun.entry.Onset {
			return false
		}
	}
	return true
}

func (e *Engine) makeMeaning(p *compile.Pack, pat grammar.Pattern, selected map[grammar.Slot]match, req context.Request) (sem.Meaning, error) {
	noun := selected[grammar.Noun]
	deixis := req.Deixis
	if deixis == "" {
		deixis = context.Unresolved
	}
	if p.Specials.DemoDenotesProximal {
		deixis = context.Proximal
	}
	frame, licensed := p.Frame(pat.FrameID)
	if !licensed || frame.Predicate != pat.Predicate {
		return sem.Meaning{}, fmt.Errorf("unlicensed target frame")
	}
	root := sem.Node{ID: "s1", Kind: sem.State, Concept: frame.Predicate}
	ref := sem.Node{ID: "r1", Kind: sem.Referent, Reference: deixis, Concept: frame.ReferentConcept}
	if ref.Concept == "" {
		ref.Concept = noun.sense.Concept
	}
	value, ok := selected[frame.ValueSlot]
	if !ok || value.sense.Concept == "" {
		return sem.Meaning{}, fmt.Errorf("unbound predicate value")
	}
	root.Value = value.sense.Concept
	root.Arguments = []sem.Argument{{Role: frame.Role, Target: ref.ID}}
	m := sem.Meaning{Root: root.ID, Nodes: []sem.Node{ref, root}}
	return m, sem.Validate(m, e.catalog)
}

func (e *Engine) Analyze(language, text string, req context.Request) Result {
	p := e.packs[language]
	if p == nil {
		return fail(status.Unsupported, "UNKNOWN_LANGUAGE", "source language package is unavailable")
	}
	if !req.Valid() {
		return fail(status.Invalid, "INVALID_CONTEXT", "unknown deixis")
	}
	tokens, err := words(text)
	if err != nil {
		return fail(status.Unsupported, "UNSUPPORTED_SENTENCE", err.Error())
	}
	found := make(map[string]Candidate)
	for _, pat := range p.Constructions {
		if pat.Event {
			continue
		}
		if !pat.Analyze || len(pat.Sequence) != len(tokens) {
			continue
		}
		choices := make([][]match, len(tokens))
		possible := true
		for i, el := range pat.Sequence {
			choices[i] = e.possible(p, el.Slot, tokens[i])
			if len(choices[i]) == 0 {
				possible = false
				break
			}
		}
		if !possible {
			continue
		}
		selected := map[grammar.Slot]match{}
		steps := 0
		overBudget := false
		var visit func(int) bool
		visit = func(i int) bool {
			steps++
			if steps > maxAnalysisSteps {
				overBudget = true
				return true
			}
			if i == len(choices) {
				if !compatible(p, pat, selected, req) {
					return false
				}
				m, err := e.makeMeaning(p, pat, selected, req)
				if err != nil {
					return false
				}
				key, err := sem.CanonicalKey(m, e.catalog)
				if err != nil {
					return false
				}
				evidence := make([]Evidence, 0, len(pat.Sequence))
				for idx, elem := range pat.Sequence {
					x := selected[elem.Slot]
					node := sem.NodeID("s1")
					if elem.Slot == grammar.Noun || elem.Slot == grammar.Demonstrative {
						node = "r1"
					}
					evidence = append(evidence, Evidence{Token: idx, Lexeme: x.entry.ID, Sense: x.sense.ID, Node: node})
				}
				found[key] = Candidate{Meaning: m, Evidence: evidence}
				return len(found) > maxCandidates
			}
			for _, c := range choices[i] {
				selected[pat.Sequence[i].Slot] = c
				if visit(i + 1) {
					return true
				}
			}
			return false
		}
		if visit(0) {
			if overBudget {
				return fail(status.ResourceLimit, "ANALYSIS_WORK_LIMIT", "analysis search work budget exceeded")
			}
			return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "analysis candidate limit exceeded")
		}
	}
	// Event constructions reuse this Meaning IR and the same result envelope.
	for _, pat := range p.Constructions {
		if !pat.Event || !pat.Analyze {
			continue
		}
		for _, candidate := range e.analyzeEventPattern(p, pat, tokens, req) {
			key, err := sem.CanonicalKey(candidate.Meaning, e.catalog)
			if err != nil {
				continue
			}
			found[key] = candidate
			if len(found) > maxCandidates {
				return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "analysis candidate limit exceeded")
			}
		}
	}
	if len(found) == 0 {
		return fail(status.Unsupported, "NO_LICENSED_PARSE", "no compiled construction accepts the complete input")
	}
	keys := make([]string, 0, len(found))
	for k := range found {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := Result{Outcome: status.OK, Candidates: make([]Candidate, 0, len(keys))}
	for _, k := range keys {
		result.Candidates = append(result.Candidates, found[k])
	}
	if len(result.Candidates) > 1 {
		result.Outcome = status.Ambiguous
	}
	return result
}

func valueOf(m sem.Meaning) (sem.Node, sem.Node, error) {
	nodes := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		nodes[n.ID] = n
	}
	predicate, ok := nodes[m.Root]
	if !ok || len(predicate.Arguments) != 1 {
		return sem.Node{}, sem.Node{}, fmt.Errorf("unsupported predicate layout")
	}
	subject, ok := nodes[predicate.Arguments[0].Target]
	if !ok {
		return sem.Node{}, sem.Node{}, fmt.Errorf("missing subject")
	}
	return predicate, subject, nil
}

func selectLexemes(p *compile.Pack, concept core.ID, pos lexicon.POS) []lexicon.Entry {
	var out []lexicon.Entry
	for _, v := range p.Lexicon.ByConcept(concept) {
		if v.POS == pos {
			out = append(out, v)
		}
	}
	return out
}
func capitalize(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (e *Engine) render(p *compile.Pack, pat grammar.Pattern, ref sem.Node, noun, adj lexicon.Entry) (string, error) {
	if ref.Reference == context.Distal && p.Specials.DemoDenotesProximal {
		return "", fmt.Errorf("source marks distal, target template requires proximal")
	}
	if ref.Reference == context.Unresolved && p.Specials.DemoDenotesProximal {
		return "", fmt.Errorf("target template requires unresolved proximal reference choice")
	}
	parts := make([]string, 0, len(pat.Sequence))
	for _, element := range pat.Sequence {
		var text string
		var err error
		switch element.Slot {
		case grammar.Noun:
			text, err = p.Morph.Generate(noun.ID, features.Spec{Number: "singular", Case: "nominative"})
		case grammar.Adjective:
			text, err = p.Morph.Generate(adj.ID, features.Spec{Number: "singular", Gender: noun.Gender, Case: "nominative", Degree: "positive"})
		case grammar.Demonstrative:
			gender := noun.Gender
			if pat.DemoGender == "fixed_neuter" {
				gender = "neuter"
			}
			if pat.DemoGender == "invariant" {
				gender = ""
			}
			text, err = p.Morph.Generate(p.Specials.Demonstrative, features.Spec{Number: "singular", Gender: gender})
		case grammar.Copula:
			text, err = p.Morph.Generate(p.Specials.Copula, features.Spec{Number: "singular", Person: "third", Tense: "present"})
		case grammar.Article:
			if noun.Onset == "" {
				return "", fmt.Errorf("target noun lacks attested article onset")
			}
			text, err = p.Morph.Generate(p.Specials.Article, features.Spec{Onset: noun.Onset})
		}
		if err != nil {
			return "", err
		}
		parts = append(parts, text)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("empty realization")
	}
	return capitalize(strings.Join(parts, " ")) + ".", nil
}

func (e *Engine) Realize(language string, meaning sem.Meaning) Result {
	p := e.packs[language]
	if p == nil {
		return fail(status.Unsupported, "UNKNOWN_LANGUAGE", "target language package is unavailable")
	}
	if err := sem.Validate(meaning, e.catalog); err != nil {
		return fail(status.Invalid, "INVALID_MEANING", err.Error())
	}
	for _, n := range meaning.Nodes {
		if n.ID != meaning.Root {
			continue
		}
		if n.Kind == sem.Event {
			return e.realizeEvent(p, meaning, n.ID, false)
		}
		if n.Kind == sem.Scope && n.Operator == sem.Not {
			return e.realizeEvent(p, meaning, n.Content, true)
		}
	}
	pred, ref, err := valueOf(meaning)
	if err != nil {
		return fail(status.Unsupported, "UNSUPPORTED_MEANING", err.Error())
	}
	if ref.Reference == context.Unresolved && p.Specials.DemoDenotesProximal {
		return fail(status.Incomplete, "REFERENCE_DEIXIS_UNRESOLVED", "the source meaning does not authorize a proximal demonstrative; context or a declared alternative is needed")
	}
	produced := map[string]Candidate{}
	for _, pat := range p.Constructions {
		if pat.Event || !pat.Realize || pat.Predicate != pred.Concept {
			continue
		}
		frame, licensed := p.Frame(pat.FrameID)
		if !licensed || frame.Predicate != pred.Concept || pred.Arguments[0].Role != frame.Role {
			continue
		}
		if frame.ReferentConcept != "" && ref.Concept != frame.ReferentConcept {
			continue
		}
		nounConcept := ref.Concept
		if frame.ValueSlot == grammar.Noun {
			nounConcept = pred.Value
		}
		nouns := selectLexemes(p, nounConcept, lexicon.Noun)
		adjectives := []lexicon.Entry{{}}
		if frame.ValueSlot == grammar.Adjective {
			adjectives = selectLexemes(p, pred.Value, lexicon.Adjective)
		}
		for _, noun := range nouns {
			for _, adj := range adjectives {
				text, err := e.render(p, pat, ref, noun, adj)
				if err != nil {
					continue
				}
				if len(produced) >= maxCandidates {
					return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "realization candidates exceeded bound")
				}
				produced[text] = Candidate{Meaning: meaning.Clone(), Text: text}
			}
		}
	}
	if len(produced) == 0 {
		return fail(status.Unsupported, "TARGET_CONSTRUCTION_UNAVAILABLE", "no target variant can realize the complete meaning")
	}
	keys := make([]string, 0, len(produced))
	for k := range produced {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := Result{Outcome: status.OK, Candidates: make([]Candidate, 0, len(keys))}
	for _, k := range keys {
		out.Candidates = append(out.Candidates, produced[k])
	}
	if len(out.Candidates) > 1 {
		out.Outcome = status.Ambiguous
	}
	return out
}

func (e *Engine) Translate(source, target, text string, ctx context.Request) Result {
	if source == target {
		return fail(status.Invalid, "SAME_LANGUAGE", "use Analyze and Realize for same-language evaluation")
	}
	analyses := e.Analyze(source, text, ctx)
	if analyses.Outcome != status.OK && analyses.Outcome != status.Ambiguous {
		return analyses
	}
	texts := map[string]Candidate{}
	for _, analysis := range analyses.Candidates {
		realized := e.Realize(target, analysis.Meaning)
		if realized.Outcome != status.OK && realized.Outcome != status.Ambiguous {
			return realized
		}
		for _, cand := range realized.Candidates {
			if len(texts) >= maxCandidates {
				return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "translation candidates exceeded bound")
			}
			cand.Evidence = append([]Evidence(nil), analysis.Evidence...)
			cand.Assumptions = append(append([]Annotation(nil), analysis.Assumptions...), cand.Assumptions...)
			cand.Losses = append(append([]Annotation(nil), analysis.Losses...), cand.Losses...)
			key, err := sem.CanonicalKey(cand.Meaning, e.catalog)
			if err != nil {
				return fail(status.Invalid, "INVALID_TRANSLATION_MEANING", err.Error())
			}
			texts[cand.Text+"\x00"+key] = cand
		}
	}
	if len(texts) == 0 {
		return fail(status.Unsupported, "NO_TRANSLATION", "no translation was produced")
	}
	keys := make([]string, 0, len(texts))
	for k := range texts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := Result{Outcome: status.OK, Candidates: make([]Candidate, 0, len(keys))}
	for _, k := range keys {
		out.Candidates = append(out.Candidates, texts[k])
	}
	if len(out.Candidates) > 1 {
		out.Outcome = status.Ambiguous
	}
	return out
}
