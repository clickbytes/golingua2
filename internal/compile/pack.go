// Package compile checks language packages once and builds immutable runtime indexes.
package compile

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/morph"
)

type Specials struct {
	Demonstrative string
	Copula        string
	Article       string
	// True only when the demonstrative itself encodes proximal deixis.
	DemoDenotesProximal bool
}

type Definition struct {
	Language      string
	Lexemes       []lexicon.Entry
	Senses        []lexicon.Sense
	MorphClasses  []morph.Class
	Frames        []grammar.Frame
	Constructions []grammar.Pattern
	Specials      Specials
}

type Pack struct {
	frames        map[string]grammar.Frame
	Language      string
	Lexicon       *lexicon.Index
	Morph         *morph.Generator
	Constructions []grammar.Pattern
	Specials      Specials
}

func Compile(d Definition, catalog *core.Catalog) (*Pack, error) {
	if d.Language != "pl" && d.Language != "en" {
		return nil, fmt.Errorf("unsupported language %q", d.Language)
	}
	li, err := lexicon.Compile(d.Lexemes, d.Senses, catalog, d.Language)
	if err != nil {
		return nil, err
	}
	mi, err := morph.Compile(d.MorphClasses, li.Entries())
	if err != nil {
		return nil, err
	}
	if d.Specials.Demonstrative == "" || d.Specials.Copula == "" {
		return nil, fmt.Errorf("missing demonstrative or copula")
	}
	demo, ok := li.Lookup(d.Specials.Demonstrative)
	if !ok || demo.POS != lexicon.Determiner {
		return nil, fmt.Errorf("invalid demonstrative")
	}
	copula, ok := li.Lookup(d.Specials.Copula)
	if !ok || copula.POS != lexicon.Verb {
		return nil, fmt.Errorf("invalid copula")
	}
	if d.Specials.Article != "" {
		a, ok := li.Lookup(d.Specials.Article)
		if !ok || a.POS != lexicon.Article {
			return nil, fmt.Errorf("invalid article")
		}
	}
	if len(d.Frames) == 0 || len(d.Frames) > 1024 {
		return nil, fmt.Errorf("invalid valency frame count")
	}
	frames := make(map[string]grammar.Frame, len(d.Frames))
	for _, f := range d.Frames {
		if err := grammar.ValidateFrame(f, catalog); err != nil {
			return nil, err
		}
		if _, exists := frames[f.ID]; exists {
			return nil, fmt.Errorf("duplicate valency frame %q", f.ID)
		}
		frames[f.ID] = f
	}
	if len(d.Constructions) == 0 || len(d.Constructions) > 1000 {
		return nil, fmt.Errorf("invalid construction count")
	}
	constructions := make([]grammar.Pattern, len(d.Constructions))
	seen := map[string]bool{}
	for i, p := range d.Constructions {
		if err := grammar.Validate(p); err != nil {
			return nil, err
		}
		if seen[p.ID] {
			return nil, fmt.Errorf("duplicate construction %q", p.ID)
		}
		seen[p.ID] = true
		hasArticle := false
		for _, e := range p.Sequence {
			if e.Slot == grammar.Article {
				hasArticle = true
			}
		}
		if hasArticle && d.Specials.Article == "" {
			return nil, fmt.Errorf("missing article for construction %q", p.ID)
		}
		frame, ok := frames[p.FrameID]
		if !ok || p.Predicate != frame.Predicate {
			return nil, fmt.Errorf("construction %q has unlicensed predicate/frame", p.ID)
		}
		hasAdjective := false
		for _, el := range p.Sequence {
			if el.Slot == grammar.Adjective {
				hasAdjective = true
			}
		}
		if hasAdjective != (frame.ValueSlot == grammar.Adjective) {
			return nil, fmt.Errorf("construction %q cannot cover frame value", p.ID)
		}
		if hasArticle && frame.ValueSlot != grammar.Noun {
			return nil, fmt.Errorf("article cannot be a clause value")
		}
		if p.DemoGender == "agree" && frame.ValueSlot != grammar.Adjective {
			return nil, fmt.Errorf("demonstrative agreement lacks declared bearer")
		}

		p.Sequence = append([]grammar.Element(nil), p.Sequence...)
		constructions[i] = p
	}
	sort.Slice(constructions, func(i, j int) bool { return constructions[i].ID < constructions[j].ID })
	return &Pack{frames: frames, Language: d.Language, Lexicon: li, Morph: mi, Constructions: constructions, Specials: d.Specials}, nil
}

// Frame returns a copied immutable compiled binding, not mutable declaration data.
func (p *Pack) Frame(id string) (grammar.Frame, bool) {
	if p == nil {
		return grammar.Frame{}, false
	}
	f, ok := p.frames[id]
	return f, ok
}
