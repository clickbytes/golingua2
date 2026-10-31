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

// Questions are scopes over an ordinary proposition with an optionally typed
// answer gap. No query is answered from a lexeme, token order, or guessed fact.
func questionMarker(p *compile.Pack, kind string) string {
	id := p.Specials.QuestionWords[kind]
	if id == "" {
		return ""
	}
	surface, err := p.Morph.Generate(id, features.Spec{})
	if err != nil {
		return ""
	}
	return surface
}
func questionStarts(p *compile.Pack, tokens []string, kind string) int {
	marker := strings.Fields(questionMarker(p, kind))
	if len(marker) == 0 || len(tokens) < len(marker) {
		return 0
	}
	if markerLicensed(p, strings.Join(tokens[:len(marker)], " "), p.Specials.QuestionWords[kind]) {
		return len(marker)
	}
	return 0
}
func questionScope(kind, role string, base sem.Meaning) sem.Meaning {
	m := base.Clone()
	m.Nodes = append(m.Nodes, sem.Node{ID: "q1", Kind: sem.Scope, Operator: sem.Ask, Content: m.Root, QuestionKind: kind, QuestionRole: role})
	m.Root = "q1"
	return m
}
func (e *Engine) queriedCandidate(base sem.Meaning, kind, role string) (Candidate, bool) {
	m := questionScope(kind, role, base)
	if sem.Validate(m, e.catalog) != nil {
		return Candidate{}, false
	}
	return Candidate{Meaning: m}, true
}
func questionEventFrames(p *compile.Pack) []grammar.EventFrame {
	seen := map[string]bool{}
	var out []grammar.EventFrame
	for _, pat := range p.Constructions {
		if !pat.Event || pat.Negated || !pat.Analyze || seen[pat.FrameID] {
			continue
		}
		seen[pat.FrameID] = true
		if f, ok := p.EventFrame(pat.FrameID); ok {
			out = append(out, f)
		}
	}
	return out
}
func frameEvent(frame grammar.EventFrame, verb verbReading, subject, object sem.Node) sem.Meaning {
	subject.ID = "r1"
	predicate := sem.Node{ID: "e1", Kind: sem.Event, Concept: verb.sense.Concept, Tense: "present", Arguments: []sem.Argument{{Role: frame.SubjectRole, Target: "r1"}}}
	nodes := []sem.Node{predicate, subject}
	if frame.ObjectRole != "" {
		object.ID = "r2"
		predicate.Arguments = append(predicate.Arguments, sem.Argument{Role: frame.ObjectRole, Target: "r2"})
		nodes[0] = predicate
		nodes = append(nodes, object)
	}
	return sem.Meaning{Root: "e1", Nodes: nodes}
}
func (e *Engine) analyzePolar(p *compile.Pack, tokens []string, req context.Request) []Candidate {
	var out []Candidate
	if p.Language == "pl" {
		if questionStarts(p, tokens, "polar") != 1 {
			return nil
		}
		remainder := strings.Join(tokens[1:], " ") + "."
		parsed := e.Analyze(p.Language, remainder, req)
		for _, cand := range parsed.Candidates {
			if x, ok := e.queriedCandidate(cand.Meaning, "polar", ""); ok {
				out = append(out, x)
			}
		}
		return out
	}
	if len(tokens) < 3 {
		return nil
	}
	if markerLicensed(p, tokens[0], p.Specials.Copula) {
		for i := 1; i < len(tokens); i++ {
			positive := append([]string(nil), tokens[1:i+1]...)
			positive = append(positive, tokens[0])
			positive = append(positive, tokens[i+1:]...)
			result := e.Analyze(p.Language, strings.Join(positive, " ")+".", req)
			for _, cand := range result.Candidates {
				n := map[sem.NodeID]sem.Node{}
				for _, x := range cand.Meaning.Nodes {
					n[x.ID] = x
				}
				if n[cand.Meaning.Root].Kind != sem.State {
					continue
				}
				if x, ok := e.queriedCandidate(cand.Meaning, "polar", ""); ok {
					out = append(out, x)
				}
			}
		}
	}
	for _, f := range questionEventFrames(p) {
		if len(tokens) < 4 {
			continue
		}
		for subLen := 1; subLen <= 2; subLen++ {
			verbIndex := 1 + subLen
			if verbIndex >= len(tokens) {
				continue
			}
			subjects := e.parseNounPhrase(p, tokens[1:verbIndex], "nominative", req, true)
			objects := []nounReading{{}}
			if f.ObjectRole != "" {
				objects = e.parseNounPhrase(p, tokens[verbIndex+1:], "nominative", req, false)
			} else if verbIndex != len(tokens)-1 {
				continue
			}
			for _, sub := range subjects {
				auxiliaryValid := false
				for _, aux := range negativeAuxiliaries(p, tokens[0]) {
					if agreesWithVerb(sub.ref, aux) {
						auxiliaryValid = true
						break
					}
				}
				if !auxiliaryValid {
					continue
				}
				for _, verb := range e.parseEventVerb(p, tokens[verbIndex], f.ID, true) {
					for _, obj := range objects {
						if m, ok := e.queriedCandidate(frameEvent(f, verb, sub.ref, obj.ref), "polar", ""); ok {
							out = append(out, m)
						}
					}
				}
			}
		}
	}
	return out
}
func (e *Engine) analyzeWho(p *compile.Pack, tokens []string, req context.Request) []Candidate {
	if questionStarts(p, tokens, "who") != 1 || len(tokens) < 3 {
		return nil
	}
	var out []Candidate
	for _, f := range questionEventFrames(p) {
		if f.ObjectRole == "" {
			continue
		}
		roleCase := f.ObjectCase
		if p.Language == "en" {
			roleCase = "nominative"
		}
		for _, verb := range e.parseEventVerb(p, tokens[1], f.ID, false) {
			if verb.person != "third" || verb.number != "singular" {
				continue
			}
			for _, obj := range e.parseNounPhrase(p, tokens[2:], roleCase, req, false) {
				unknown := sem.Node{Kind: sem.Referent, Concept: "entity.unspecified", Number: "singular"}
				if c, ok := e.queriedCandidate(frameEvent(f, verb, unknown, obj.ref), "who", f.SubjectRole); ok {
					out = append(out, c)
				}
			}
		}
	}
	return out
}
func (e *Engine) analyzeWhat(p *compile.Pack, tokens []string, req context.Request) []Candidate {
	if questionStarts(p, tokens, "what") != 1 || len(tokens) < 3 {
		return nil
	}
	var out []Candidate
	for _, f := range questionEventFrames(p) {
		if f.ObjectRole == "" {
			continue
		}
		for subLen := 1; subLen <= 2; subLen++ {
			var subjects []nounReading
			var verbs []verbReading
			auxText := ""
			if p.Language == "pl" {
				if 2+subLen != len(tokens) {
					continue
				}
				verbs = e.parseEventVerb(p, tokens[1], f.ID, false)
				subjects = e.parseNounPhrase(p, tokens[2:], "nominative", req, true)
			} else {
				// What + do/does + subject NP + lexical base verb.
				if 3+subLen != len(tokens) {
					continue
				}
				auxText = tokens[1]
				subjects = e.parseNounPhrase(p, tokens[2:2+subLen], "nominative", req, true)
				verbs = e.parseEventVerb(p, tokens[len(tokens)-1], f.ID, true)
			}
			for _, sub := range subjects {
				if auxText != "" {
					valid := false
					for _, aux := range negativeAuxiliaries(p, auxText) {
						if agreesWithVerb(sub.ref, aux) {
							valid = true
						}
					}
					if !valid {
						continue
					}
				}
				for _, verb := range verbs {
					if auxText == "" && !agreesWithVerb(sub.ref, verb) {
						continue
					}
					unknown := sem.Node{Kind: sem.Referent, Concept: "entity.unspecified"}
					if c, ok := e.queriedCandidate(frameEvent(f, verb, sub.ref, unknown), "what", f.ObjectRole); ok {
						out = append(out, c)
					}
				}
			}
		}
	}
	return out
}
func (e *Engine) analyzeWhere(p *compile.Pack, tokens []string, req context.Request) []Candidate {
	if questionStarts(p, tokens, "where") != 1 || len(tokens) < 3 {
		return nil
	}
	validCopula := false
	for _, form := range p.Morph.Analyze(tokens[1]) {
		if form.LexemeID == p.Specials.Copula && form.Features.Number == "singular" && form.Features.Person == "third" && form.Features.Tense == "present" {
			validCopula = true
			break
		}
	}
	if !validCopula {
		return nil
	}
	var out []Candidate
	for _, sub := range e.parseNounPhrase(p, tokens[2:], "nominative", req, true) {
		if sub.ref.Number != "singular" {
			continue
		}
		located := sub.ref
		located.ID = "r1"
		site := sem.Node{ID: "r2", Kind: sem.Referent, Concept: "entity.unspecified"}
		pred := sem.Node{ID: "l1", Kind: sem.State, Concept: "relation.located", Arguments: []sem.Argument{{Role: "located", Target: located.ID}, {Role: "site", Target: site.ID}}}
		if c, ok := e.queriedCandidate(sem.Meaning{Root: pred.ID, Nodes: []sem.Node{pred, located, site}}, "where", "site"); ok {
			out = append(out, c)
		}
	}
	return out
}
func (e *Engine) analyzeHowMany(p *compile.Pack, tokens []string, req context.Request) []Candidate {
	markerLen := questionStarts(p, tokens, "how_many")
	if markerLen == 0 {
		return nil
	}
	var out []Candidate
	for _, f := range questionEventFrames(p) {
		if f.ObjectRole == "" {
			continue
		}
		var subject nounReading
		var verbs []verbReading
		var objects []nounReading
		if p.Language == "pl" {
			if len(tokens) != markerLen+2 || !f.AllowImplicitSubject {
				continue
			}
			verbs = e.parseEventVerb(p, tokens[markerLen], f.ID, false)
			objects = e.parseNounPhrase(p, tokens[markerLen+1:], "genitive", req, false)
		} else {
			if len(tokens) != markerLen+4 {
				continue
			}
			objects = e.parseNounPhrase(p, tokens[markerLen:markerLen+1], "nominative", req, false)
			subjects := e.parseNounPhrase(p, tokens[markerLen+2:markerLen+3], "nominative", req, true)
			if len(subjects) != 1 {
				continue
			}
			subject = subjects[0]
			validAux := false
			for _, aux := range negativeAuxiliaries(p, tokens[markerLen+1]) {
				if agreesWithVerb(subject.ref, aux) {
					validAux = true
					break
				}
			}
			if !validAux {
				continue
			}
			verbs = e.parseEventVerb(p, tokens[markerLen+3], f.ID, true)
		}
		for _, verb := range verbs {
			subjectRef := subject.ref
			if p.Language == "pl" {
				if verb.person != "first" && verb.person != "second" {
					continue
				}
				subjectRef = sem.Node{Kind: sem.Referent, Concept: "entity.person", Person: verb.person, Number: verb.number}
			}
			for _, obj := range objects {
				if obj.ref.Number != "plural" || obj.lex.Countability == "mass" {
					continue
				}
				ref := obj.ref
				ref.Definiteness = ""
				ref.Cardinality = 0
				if c, ok := e.queriedCandidate(frameEvent(f, verb, subjectRef, ref), "how_many", f.ObjectRole); ok {
					out = append(out, c)
				}
			}
		}
	}
	return out
}
func (e *Engine) analyzeQuestion(p *compile.Pack, tokens []string, req context.Request) Result {
	if len(p.Specials.QuestionWords) != 5 {
		return fail(status.Unsupported, "QUESTION_CAPABILITY_UNAVAILABLE", "question operator is not licensed by language pack")
	}
	var choices []Candidate
	choices = append(choices, e.analyzePolar(p, tokens, req)...)
	choices = append(choices, e.analyzeWho(p, tokens, req)...)
	choices = append(choices, e.analyzeWhat(p, tokens, req)...)
	choices = append(choices, e.analyzeWhere(p, tokens, req)...)
	choices = append(choices, e.analyzeHowMany(p, tokens, req)...)
	byKey := map[string]Candidate{}
	for _, c := range choices {
		key, err := sem.CanonicalKey(c.Meaning, e.catalog)
		if err != nil {
			continue
		}
		byKey[key] = c
		if len(byKey) > maxCandidates {
			return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "question analysis alternatives exceeded bound")
		}
	}
	if len(byKey) == 0 {
		return fail(status.Unsupported, "QUESTION_CONSTRUCTION_UNAVAILABLE", "no licensed question construction covers this input")
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	r := Result{Outcome: status.OK}
	for _, k := range keys {
		r.Candidates = append(r.Candidates, byKey[k])
	}
	if len(keys) > 1 {
		r.Outcome = status.Ambiguous
	}
	return r
}

// Render questions from the typed proposition. The interrogative prefix is not
// a translation phrase; it is a target-specific grammatical operator.
func (e *Engine) realizeQuestion(p *compile.Pack, m sem.Meaning, root sem.Node) Result {
	if len(p.Specials.QuestionWords) != 5 {
		return fail(status.Unsupported, "QUESTION_CAPABILITY_UNAVAILABLE", "target question morphology unavailable")
	}
	var predicate sem.Node
	byID := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	predicate = byID[root.Content]
	if root.QuestionKind == "polar" && p.Language == "pl" {
		affirmative := m.Clone()
		affirmative.Root = root.Content
		affirmative.Nodes = excludeScopeNode(affirmative.Nodes, root.ID)
		base := e.Realize(p.Language, affirmative)
		return prefixQuestions(base, questionMarker(p, "polar"), m)
	}
	if root.QuestionKind == "polar" && predicate.Kind == sem.State {
		affirmative := m.Clone()
		affirmative.Root = root.Content
		affirmative.Nodes = excludeScopeNode(affirmative.Nodes, root.ID)
		base := e.Realize(p.Language, affirmative)
		if base.Outcome != status.OK && base.Outcome != status.Ambiguous {
			return base
		}
		copula, err := p.Morph.Generate(p.Specials.Copula, features.Spec{Tense: "present", Person: "third", Number: "singular"})
		if err != nil {
			return fail(status.Unsupported, "QUESTION_COPULA_UNAVAILABLE", err.Error())
		}
		var out []Candidate
		for _, cand := range base.Candidates {
			tokens := strings.Fields(strings.TrimSuffix(cand.Text, "."))
			for i := 1; i < len(tokens); i++ {
				if strings.EqualFold(tokens[i], copula) {
					reordered := append([]string{copula}, tokens[:i]...)
					reordered[1] = strings.ToLower(reordered[1][:1]) + reordered[1][1:]
					reordered = append(reordered, tokens[i+1:]...)
					out = append(out, Candidate{Text: capitalize(strings.Join(reordered, " ")) + "?", Meaning: m.Clone(), Assumptions: cand.Assumptions, Losses: cand.Losses})
					break
				}
			}
		}
		return questionResult(out)
	}
	if predicate.Kind == sem.Event {
		return e.realizeEventQuestion(p, m, root, predicate, byID)
	}
	if root.QuestionKind == "where" {
		located := byID[predicate.Arguments[0].Target]
		subject := e.renderEventNP(p, located, "nominative")
		copula, err := p.Morph.Generate(p.Specials.Copula, features.Spec{Number: "singular", Person: "third", Tense: "present"})
		if err != nil {
			return fail(status.Unsupported, "QUESTION_COPULA_UNAVAILABLE", err.Error())
		}
		marker := questionMarker(p, "where")
		var candidates []Candidate
		for _, s := range subject {
			candidates = append(candidates, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+copula+" "+s.text) + "?", Assumptions: s.assumptions, Losses: s.losses})
		}
		return questionResult(candidates)
	}
	return fail(status.Unsupported, "QUESTION_CONSTRUCTION_UNAVAILABLE", "unsupported typed question")
}
func excludeScopeNode(nodes []sem.Node, id sem.NodeID) []sem.Node {
	out := make([]sem.Node, 0, len(nodes))
	for _, n := range nodes {
		if n.ID != id {
			out = append(out, n)
		}
	}
	return out
}
func prefixQuestions(base Result, marker string, m sem.Meaning) Result {
	if base.Outcome != status.OK && base.Outcome != status.Ambiguous {
		return base
	}
	if marker == "" {
		return fail(status.Unsupported, "QUESTION_MARKER_UNAVAILABLE", "question particle unavailable")
	}
	var out []Candidate
	for _, c := range base.Candidates {
		out = append(out, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+strings.ToLower(c.Text[:1])+c.Text[1:len(c.Text)-1]) + "?", Assumptions: c.Assumptions, Losses: c.Losses})
	}
	return questionResult(out)
}
func questionResult(choices []Candidate) Result {
	if len(choices) == 0 {
		return fail(status.Unsupported, "QUESTION_CONSTRUCTION_UNAVAILABLE", "no target interrogative is licensed")
	}
	byText := map[string]Candidate{}
	for _, c := range choices {
		if c.Text == "" {
			continue
		}
		byText[c.Text] = c
		if len(byText) > maxCandidates {
			return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "question target alternatives exceeded bound")
		}
	}
	keys := make([]string, 0, len(byText))
	for k := range byText {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	r := Result{Outcome: status.OK}
	for _, k := range keys {
		r.Candidates = append(r.Candidates, byText[k])
	}
	if len(keys) > 1 {
		r.Outcome = status.Ambiguous
	}
	return r
}
func questionVerbOptions(p *compile.Pack, predicate sem.Node, frame grammar.EventFrame) []lexicon.Entry {
	out := []lexicon.Entry{}
	for _, lex := range selectLexemes(p, predicate.Concept, lexicon.Verb) {
		for _, sense := range p.Lexicon.SensesFor(lex.ID) {
			if licensedFrame(sense, frame.ID) {
				out = append(out, lex)
				break
			}
		}
	}
	return out
}
func (e *Engine) renderQuestionNPWithoutNumeral(p *compile.Pack, ref sem.Node, gramCase string) []string {
	if ref.Kind != sem.Referent || ref.Person != "" || ref.Number != "plural" || ref.Cardinality != 0 {
		return nil
	}
	var out []string
	for _, noun := range selectLexemes(p, ref.Concept, lexicon.Noun) {
		if noun.Countability == "mass" || noun.Countability == "plural_only" {
			continue
		}
		surface, err := p.Morph.Generate(noun.ID, features.Spec{Number: "plural", Case: gramCase})
		if err == nil {
			out = append(out, surface)
		}
	}
	return out
}
func (e *Engine) realizeEventQuestion(p *compile.Pack, m sem.Meaning, q sem.Node, event sem.Node, nodes map[sem.NodeID]sem.Node) Result {
	var choices []Candidate
	for _, frame := range questionEventFrames(p) {
		args := map[string]sem.Node{}
		for _, a := range event.Arguments {
			args[a.Role] = nodes[a.Target]
		}
		sub, ok := args[frame.SubjectRole]
		if !ok {
			continue
		}
		obj, hasObj := args[frame.ObjectRole]
		if hasObj != (frame.ObjectRole != "") {
			continue
		}
		if q.QuestionKind == "who" && q.QuestionRole != frame.SubjectRole || q.QuestionKind == "what" && q.QuestionRole != frame.ObjectRole || q.QuestionKind == "how_many" && q.QuestionRole != frame.ObjectRole {
			continue
		}
		subjOptions := e.renderEventNP(p, sub, "nominative")
		objCase := frame.ObjectCase
		if p.Language == "en" {
			objCase = "nominative"
		}
		objOptions := []phraseOption{{text: ""}}
		if hasObj && q.QuestionKind != "what" && q.QuestionKind != "how_many" {
			objOptions = e.renderEventNP(p, obj, objCase)
		}
		person := sub.Person
		if person == "" {
			person = "third"
		}
		number := sub.Number
		if number == "" {
			number = "singular"
		}
		for _, verb := range questionVerbOptions(p, event, frame) {
			finite, finiteErr := p.Morph.Generate(verb.ID, features.Spec{Tense: event.Tense, Person: person, Number: number})
			base, baseErr := p.Morph.Generate(verb.ID, features.Spec{VerbForm: "base"})
			aux, auxErr := p.Morph.Generate(p.Specials.NegationAuxiliary, features.Spec{Tense: "present", Person: person, Number: number})
			if q.QuestionKind == "who" {
				if finiteErr != nil || !hasObj {
					continue
				}
				marker := questionMarker(p, "who")
				for _, ob := range objOptions {
					choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+finite+" "+ob.text) + "?", Assumptions: ob.assumptions, Losses: ob.losses})
				}
			} else if q.QuestionKind == "what" {
				if !hasObj {
					continue
				}
				marker := questionMarker(p, "what")
				for _, subject := range subjOptions {
					if p.Language == "pl" && finiteErr == nil {
						choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+finite+" "+subject.text) + "?", Assumptions: subject.assumptions, Losses: subject.losses})
					}
					if p.Language == "en" && auxErr == nil && baseErr == nil {
						choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+aux+" "+subject.text+" "+base) + "?", Assumptions: subject.assumptions, Losses: subject.losses})
					}
				}
			} else if q.QuestionKind == "how_many" {
				if !hasObj {
					continue
				}
				marker := questionMarker(p, "how_many")
				requestedCase := "nominative"
				if p.Language == "pl" {
					requestedCase = "genitive"
				}
				nouns := e.renderQuestionNPWithoutNumeral(p, obj, requestedCase)
				for _, noun := range nouns {
					if p.Language == "pl" && finiteErr == nil && frame.AllowImplicitSubject && sub.Person != "" {
						choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+finite+" "+noun) + "?"})
					}
					if p.Language == "en" && auxErr == nil && baseErr == nil {
						for _, subject := range subjOptions {
							choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(marker+" "+noun+" "+aux+" "+subject.text+" "+base) + "?", Assumptions: subject.assumptions, Losses: subject.losses})
						}
					}
				}
			} else if q.QuestionKind == "polar" && p.Language == "en" {
				if auxErr != nil || baseErr != nil {
					continue
				}
				for _, subject := range subjOptions {
					for _, ob := range objOptions {
						parts := []string{aux, subject.text, base}
						if hasObj {
							parts = append(parts, ob.text)
						}
						choices = append(choices, Candidate{Meaning: m.Clone(), Text: capitalize(strings.Join(parts, " ")) + "?", Assumptions: append(append([]Annotation{}, subject.assumptions...), ob.assumptions...), Losses: append(append([]Annotation{}, subject.losses...), ob.losses...)})
					}
				}
			}
		}
	}
	return questionResult(choices)
}
