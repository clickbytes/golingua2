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

// Temporal clauses consume declared slots. Only NP slots vary in length; the
// search has at most 3*3 partitions and a bounded set of lexical readings.
type temporalMatch struct {
	tokens map[grammar.Slot][]string
	start  map[grammar.Slot]int
}

func temporalSegments(pat grammar.Pattern, words []string) []temporalMatch {
	uses := map[grammar.Slot]bool{}
	for _, v := range pat.Sequence {
		uses[v.Slot] = true
	}
	out := []temporalMatch{}
	subLens := []int{0}
	if uses[grammar.SubjectPhrase] {
		subLens = []int{1, 2}
	}
	objLens := []int{0}
	if uses[grammar.ObjectPhrase] {
		objLens = []int{1, 2}
	}
	for _, subLen := range subLens {
		for _, objLen := range objLens {
			widths := map[grammar.Slot]int{grammar.SubjectPhrase: subLen, grammar.ObjectPhrase: objLen}
			remaining := len(words)
			for _, v := range pat.Sequence {
				n := widths[v.Slot]
				if n == 0 {
					n = 1
				}
				remaining -= n
			}
			if remaining != 0 {
				continue
			}
			parsed := temporalMatch{tokens: map[grammar.Slot][]string{}, start: map[grammar.Slot]int{}}
			index := 0
			for _, v := range pat.Sequence {
				n := widths[v.Slot]
				if n == 0 {
					n = 1
				}
				parsed.tokens[v.Slot] = words[index : index+n]
				parsed.start[v.Slot] = index
				index += n
			}
			out = append(out, parsed)
		}
	}
	return out
}

func timeRef(tense string) string {
	if tense == "past" {
		return "previous_day"
	}
	return "next_day"
}
func grammaticalSubjectNumber(ref sem.Node) string {
	if ref.Number == "" {
		return "singular"
	}
	return ref.Number
}
func grammaticalSubjectPerson(ref sem.Node) string {
	if ref.Person == "" {
		return "third"
	}
	return ref.Person
}
func temporalVerbForm(p *compile.Pack, lex lexicon.Entry, tense string, implicit bool, ref sem.Node, gender string) (string, error) {
	if p.Language == "en" {
		if tense == "past" {
			return p.Morph.Generate(lex.ID, features.Spec{Tense: "past"})
		}
		return p.Morph.Generate(lex.ID, features.Spec{VerbForm: "base"})
	}
	if tense == "future" && implicit {
		return p.Morph.Generate(lex.ID, features.Spec{VerbForm: "infinitive"})
	}
	return p.Morph.Generate(lex.ID, features.Spec{Person: "third", Number: grammaticalSubjectNumber(ref), Gender: gender, Tense: "past"})
}

func temporalVerbReadingValid(p *compile.Pack, verb lexicon.Entry, f features.Spec, pat grammar.Pattern, implicit bool, subj sem.Node, gender string) bool {
	if p.Language == "en" {
		if pat.TemporalTense == "past" {
			return f.Tense == "past" && f.Person == "" && f.VerbForm == ""
		}
		return f.VerbForm == "base" && f.Tense == ""
	}
	if pat.TemporalTense == "future" && implicit {
		return f.VerbForm == "infinitive" && f.Tense == ""
	}
	return f.VerbForm == "" && f.Tense == "past" && f.Person == "third" && f.Number == grammaticalSubjectNumber(subj) && f.Gender == gender
}

func temporalAuxCompatible(p *compile.Pack, token string, pat grammar.Pattern, subject sem.Node) bool {
	for _, r := range p.Morph.Analyze(token) {
		if r.LexemeID != p.Specials.FutureAuxiliary {
			continue
		}
		if p.Language == "en" && r.Features == (features.Spec{}) {
			return true
		}
		if p.Language == "pl" && r.Features.Tense == "future" && r.Features.Person == grammaticalSubjectPerson(subject) && r.Features.Number == grammaticalSubjectNumber(subject) {
			return true
		}
	}
	return false
}

func (e *Engine) analyzeTemporalPattern(p *compile.Pack, pat grammar.Pattern, tokens []string, req context.Request) []Candidate {
	frame, ok := p.EventFrame(pat.FrameID)
	if !ok || !pat.Temporal {
		return nil
	}
	out := []Candidate{}
	for _, part := range temporalSegments(pat, tokens) {
		w := part.tokens
		if !markerLicensed(p, w[grammar.TemporalAdverb][0], p.Specials.TimeMarkers[pat.TemporalTense]) {
			continue
		}
		if len(w[grammar.WeatherExpletive]) > 0 && !markerLicensed(p, w[grammar.WeatherExpletive][0], p.Specials.WeatherExpletive) {
			continue
		}
		implicit := len(w[grammar.SubjectPhrase]) == 0 && !frame.Weather
		subjects := []nounReading{}
		if frame.Weather && len(w[grammar.SubjectPhrase]) == 0 {
			subjects = []nounReading{{ref: sem.Node{Kind: sem.Referent, Concept: frame.WeatherSubjectConcept}}}
		} else if !implicit {
			subjects = e.parseNounPhrase(p, w[grammar.SubjectPhrase], "nominative", req, true)
		}
		if implicit && p.Language == "pl" {
			subjects = []nounReading{{ref: sem.Node{Kind: sem.Referent, Concept: "entity.person", Person: "first", Number: "singular"}}}
		}
		if len(subjects) == 0 {
			continue
		}
		objects := []nounReading{{}}
		if frame.ObjectRole != "" {
			objCase := frame.ObjectCase
			if p.Language == "en" {
				objCase = "nominative"
			}
			objects = e.parseNounPhrase(p, w[grammar.ObjectPhrase], objCase, req, false)
			if len(objects) == 0 {
				continue
			}
		}
		for _, verbForm := range p.Morph.Analyze(w[grammar.PredicateVerb][0]) {
			verb, ok := p.Lexicon.Lookup(verbForm.LexemeID)
			if !ok || verb.POS != lexicon.Verb {
				continue
			}
			for _, sense := range p.Lexicon.SensesFor(verb.ID) {
				if !licensedFrame(sense, frame.ID) {
					continue
				}
				for _, sub := range subjects {
					if frame.Weather && sub.ref.Concept != frame.WeatherSubjectConcept {
						continue
					}
					gender := sub.lex.Gender
					if frame.Weather {
						gender = "masculine"
					}
					if !temporalVerbReadingValid(p, verb, verbForm.Features, pat, implicit, sub.ref, gender) {
						continue
					}
					if pat.TemporalTense == "future" && !temporalAuxCompatible(p, w[grammar.FutureAuxiliary][0], pat, sub.ref) {
						continue
					}
					for _, obj := range objects {
						pred := sem.Node{ID: "e1", Kind: sem.Event, Concept: sense.Concept, Tense: pat.TemporalTense, TimeReference: timeRef(pat.TemporalTense), Arguments: []sem.Argument{{Role: frame.SubjectRole, Target: "r1"}}}
						subject := sub.ref
						subject.ID = "r1"
						nodes := []sem.Node{pred, subject}
						if frame.ObjectRole != "" {
							object := obj.ref
							object.ID = "r2"
							pred.Arguments = append(pred.Arguments, sem.Argument{Role: frame.ObjectRole, Target: "r2"})
							nodes[0] = pred
							nodes = append(nodes, object)
						}
						m := sem.Meaning{Root: pred.ID, Nodes: nodes}
						if sem.Validate(m, e.catalog) != nil {
							continue
						}
						evidence := []Evidence{{Token: part.start[grammar.PredicateVerb], Lexeme: verb.ID, Sense: sense.ID, Node: pred.ID},
							{Token: part.start[grammar.TemporalAdverb], Lexeme: p.Specials.TimeMarkers[pat.TemporalTense], Node: pred.ID}}
						if len(w[grammar.SubjectPhrase]) > 0 {
							evidence = append(evidence, Evidence{Token: part.start[grammar.SubjectPhrase], Lexeme: sub.lex.ID, Sense: sub.sense.ID, Node: subject.ID})
						}
						if frame.ObjectRole != "" {
							evidence = append(evidence, Evidence{Token: part.start[grammar.ObjectPhrase], Lexeme: obj.lex.ID, Sense: obj.sense.ID, Node: "r2"})
						}
						if pat.TemporalTense == "future" {
							evidence = append(evidence, Evidence{Token: part.start[grammar.FutureAuxiliary], Lexeme: p.Specials.FutureAuxiliary, Node: pred.ID})
						}
						if frame.Weather && len(w[grammar.WeatherExpletive]) > 0 {
							evidence = append(evidence, Evidence{Token: part.start[grammar.WeatherExpletive], Lexeme: p.Specials.WeatherExpletive, Node: "r1"})
						}
						out = append(out, Candidate{Meaning: m, Evidence: evidence})
					}
				}
			}
		}
	}
	return out
}

func (e *Engine) realizeTemporal(p *compile.Pack, m sem.Meaning, eventID sem.NodeID) Result {
	byID := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	event := byID[eventID]
	made := map[string]Candidate{}
	for _, pat := range p.Constructions {
		if !pat.Event || !pat.Temporal || !pat.Realize || pat.TemporalTense != event.Tense {
			continue
		}
		frame, ok := p.EventFrame(pat.FrameID)
		if !ok {
			continue
		}
		refs := map[string]sem.Node{}
		for _, a := range event.Arguments {
			refs[a.Role] = byID[a.Target]
		}
		sub, ok := refs[frame.SubjectRole]
		if !ok {
			continue
		}
		if frame.Weather != (sub.Concept == frame.WeatherSubjectConcept) {
			continue
		}
		object, hasObject := refs[frame.ObjectRole]
		if (frame.ObjectRole != "") != hasObject || len(event.Arguments) != 1+btoiFrame(hasObject) {
			continue
		}
		implicit := !hasSlot(pat, grammar.SubjectPhrase) && !frame.Weather
		if implicit && (p.Language != "pl" || sub.Person != "first" || sub.Number != "singular") {
			continue
		}
		if !implicit && sub.Person != "" && p.Language == "pl" {
			continue
		}
		subjects := []phraseOption{{text: ""}}
		if hasSlot(pat, grammar.SubjectPhrase) {
			subjects = e.renderEventNP(p, sub, "nominative")
		}
		objects := []phraseOption{{text: ""}}
		if hasObject {
			objCase := frame.ObjectCase
			if p.Language == "en" {
				objCase = "nominative"
			}
			objects = e.renderEventNP(p, object, objCase)
		}
		timeText, err := p.Morph.Generate(p.Specials.TimeMarkers[event.Tense], features.Spec{})
		if err != nil {
			continue
		}
		for _, verb := range selectLexemes(p, event.Concept, lexicon.Verb) {
			authorized := false
			for _, sense := range p.Lexicon.SensesFor(verb.ID) {
				if licensedFrame(sense, frame.ID) {
					authorized = true
					break
				}
			}
			if !authorized {
				continue
			}
			for _, subOption := range subjects {
				gender := subOption.grammaticalGender
				if frame.Weather {
					gender = "masculine"
				}
				verbText, err := temporalVerbForm(p, verb, event.Tense, implicit, sub, gender)
				if err != nil {
					continue
				}
				auxText := ""
				if event.Tense == "future" {
					auxFeatures := features.Spec{}
					if p.Language == "pl" {
						auxFeatures = features.Spec{Tense: "future", Number: grammaticalSubjectNumber(sub), Person: grammaticalSubjectPerson(sub)}
					}
					auxText, err = p.Morph.Generate(p.Specials.FutureAuxiliary, auxFeatures)
					if err != nil {
						continue
					}
				}
				for _, objOption := range objects {
					parts := []string{}
					for _, element := range pat.Sequence {
						switch element.Slot {
						case grammar.TemporalAdverb:
							parts = append(parts, timeText)
						case grammar.SubjectPhrase:
							parts = append(parts, subOption.text)
						case grammar.ObjectPhrase:
							parts = append(parts, objOption.text)
						case grammar.PredicateVerb:
							parts = append(parts, verbText)
						case grammar.FutureAuxiliary:
							parts = append(parts, auxText)
						case grammar.WeatherExpletive:
							exp, err := p.Morph.Generate(p.Specials.WeatherExpletive, features.Spec{})
							if err == nil {
								parts = append(parts, exp)
							}
						}
					}
					s := capitalize(strings.Join(parts, " ")) + "."
					if len(made) >= maxCandidates {
						return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "temporal candidate budget exceeded")
					}
					if _, exists := made[s]; exists {
						continue
					}
					candidate := Candidate{Meaning: m.Clone(), Text: s}
					candidate.Assumptions = append(append([]Annotation(nil), subOption.assumptions...), objOption.assumptions...)
					candidate.Losses = append(append([]Annotation(nil), subOption.losses...), objOption.losses...)
					made[s] = candidate
				}
			}
		}
	}
	if len(made) == 0 {
		return fail(status.Unsupported, "TEMPORAL_CONSTRUCTION_UNAVAILABLE", "no time-licensed lexical frame or inflection covers this event")
	}
	keys := make([]string, 0, len(made))
	for k := range made {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := Result{Outcome: status.OK}
	for _, k := range keys {
		result.Candidates = append(result.Candidates, made[k])
	}
	if len(keys) > 1 {
		result.Outcome = status.Ambiguous
	}
	return result
}
func hasSlot(pat grammar.Pattern, slot grammar.Slot) bool {
	for _, x := range pat.Sequence {
		if x.Slot == slot {
			return true
		}
	}
	return false
}
func btoiFrame(b bool) int {
	if b {
		return 1
	}
	return 0
}
