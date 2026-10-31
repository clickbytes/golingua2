// Package lexicon separates lexical identity and sense from neutral concepts.
package lexicon

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
)

type POS string

const (
	Noun       POS = "noun"
	Adjective  POS = "adjective"
	Verb       POS = "verb"
	Determiner POS = "determiner"
	Article    POS = "article"
)

type Entry struct {
	ID       string
	Lemma    string
	Language string
	POS      POS
	ClassID  string
	Stems    map[string]string
	Gender   string
	// Onset is lexical phonology, not an English spelling heuristic.
	Onset string
	// Overrides are attested exceptional cells; keys are features.Spec.Key().
	Overrides map[string]string
}

type Sense struct {
	ID       string
	LexemeID string
	Concept  core.ID
}

type Index struct {
	entries   map[string]Entry
	senses    map[string]Sense
	byConcept map[core.ID][]Entry
	byLexeme  map[string][]Sense
}

func Compile(entries []Entry, senses []Sense, catalog *core.Catalog, language string) (*Index, error) {
	if catalog == nil || language == "" || len(entries) == 0 || len(entries) > 10000 {
		return nil, fmt.Errorf("invalid lexical package")
	}
	i := &Index{entries: make(map[string]Entry, len(entries)), senses: make(map[string]Sense, len(senses)), byConcept: make(map[core.ID][]Entry), byLexeme: make(map[string][]Sense)}
	for _, e := range entries {
		if e.ID == "" || e.Lemma == "" || strings.TrimSpace(e.Lemma) != e.Lemma || e.Language != language || e.ClassID == "" || (e.POS != Noun && e.POS != Adjective && e.POS != Verb && e.POS != Determiner && e.POS != Article) {
			return nil, fmt.Errorf("invalid lexeme %q", e.ID)
		}
		if _, ok := i.entries[e.ID]; ok {
			return nil, fmt.Errorf("duplicate lexeme %q", e.ID)
		}
		if e.POS == Noun && e.Gender == "" && language == "pl" {
			return nil, fmt.Errorf("Polish noun %q requires gender", e.ID)
		}
		if e.Onset != "" && e.Onset != "vowel" && e.Onset != "consonant" {
			return nil, fmt.Errorf("invalid lexical onset %q", e.ID)
		}
		e.Stems = copyMap(e.Stems)
		e.Overrides = copyMap(e.Overrides)
		i.entries[e.ID] = e
	}
	for _, s := range senses {
		if s.ID == "" || s.LexemeID == "" || s.Concept == "" {
			return nil, fmt.Errorf("invalid sense")
		}
		if _, ok := i.senses[s.ID]; ok {
			return nil, fmt.Errorf("duplicate sense %q", s.ID)
		}
		e, ok := i.entries[s.LexemeID]
		if !ok {
			return nil, fmt.Errorf("unbound sense %q", s.ID)
		}
		d, ok := catalog.Lookup(s.Concept)
		if !ok {
			return nil, fmt.Errorf("unknown concept for %q", s.ID)
		}
		if (e.POS == Noun && d.Kind != core.Entity) || (e.POS == Adjective && d.Kind != core.Property) {
			return nil, fmt.Errorf("lexical concept/POS mismatch at %q", s.ID)
		}
		i.senses[s.ID] = s
		i.byLexeme[s.LexemeID] = append(i.byLexeme[s.LexemeID], s)
		i.byConcept[s.Concept] = append(i.byConcept[s.Concept], e)
	}
	for k, v := range i.byConcept {
		sort.Slice(v, func(a, b int) bool { return v[a].ID < v[b].ID })
		i.byConcept[k] = v
	}
	for k, v := range i.byLexeme {
		sort.Slice(v, func(a, b int) bool { return v[a].ID < v[b].ID })
		i.byLexeme[k] = v
	}
	return i, nil
}

func (i *Index) Lookup(id string) (Entry, bool) {
	if i == nil {
		return Entry{}, false
	}
	e, ok := i.entries[id]
	e.Stems = copyMap(e.Stems)
	e.Overrides = copyMap(e.Overrides)
	return e, ok
}
func (i *Index) ByConcept(id core.ID) []Entry {
	if i == nil {
		return nil
	}
	v := i.byConcept[id]
	out := make([]Entry, len(v))
	for k, x := range v {
		out[k] = x
		out[k].Stems = copyMap(x.Stems)
		out[k].Overrides = copyMap(x.Overrides)
	}
	return out
}
func (i *Index) SenseFor(lexemeID string, concept core.ID) (Sense, bool) {
	if i == nil {
		return Sense{}, false
	}
	for _, s := range i.senses {
		if s.LexemeID == lexemeID && s.Concept == concept {
			return s, true
		}
	}
	return Sense{}, false
}

// SensesFor never guesses semantic identity from a lexeme's spelling or ID.
func (i *Index) SensesFor(lexemeID string) []Sense {
	if i == nil {
		return nil
	}
	return append([]Sense(nil), i.byLexeme[lexemeID]...)
}

func (i *Index) Entries() []Entry {
	if i == nil {
		return nil
	}
	ids := make([]string, 0, len(i.entries))
	for id := range i.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	v := make([]Entry, 0, len(ids))
	for _, id := range ids {
		e, _ := i.Lookup(id)
		v = append(v, e)
	}
	return v
}
func copyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	o := make(map[K]V, len(m))
	for k, v := range m {
		o[k] = v
	}
	return o
}

// OverrideKey is the stable key for one exceptional morphological form.
func OverrideKey(f features.Spec) string { return f.Key() }
