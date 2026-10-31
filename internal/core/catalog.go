// Package core owns language-neutral concept definitions and typed predicate signatures.
package core

import (
	"fmt"
	"sort"
	"strings"
)

type ID string
type Kind string

const (
	Entity    Kind = "entity"
	Property  Kind = "property"
	Predicate Kind = "predicate"
)

type Role struct {
	Name   string
	Target Kind
}

type Definition struct {
	ID        ID
	Kind      Kind
	Roles     []Role
	ValueKind Kind
}

type Catalog struct{ byID map[ID]Definition }

func NewCatalog(defs []Definition) (*Catalog, error) {
	if len(defs) == 0 || len(defs) > 1024 {
		return nil, fmt.Errorf("invalid concept count: %d", len(defs))
	}
	c := &Catalog{byID: make(map[ID]Definition, len(defs))}
	for _, d := range defs {
		if !validID(string(d.ID)) || (d.Kind != Entity && d.Kind != Property && d.Kind != Predicate) {
			return nil, fmt.Errorf("invalid concept: %q", d.ID)
		}
		if _, ok := c.byID[d.ID]; ok {
			return nil, fmt.Errorf("duplicate concept: %q", d.ID)
		}
		if d.Kind != Predicate && (len(d.Roles) != 0 || d.ValueKind != "") {
			return nil, fmt.Errorf("non-predicate has signature: %q", d.ID)
		}
		if d.Kind == Predicate && (len(d.Roles) == 0 || len(d.Roles) > 8 || (d.ValueKind != "" && d.ValueKind != Entity && d.ValueKind != Property)) {
			return nil, fmt.Errorf("missing predicate signature: %q", d.ID)
		}
		seen := map[string]bool{}
		for _, r := range d.Roles {
			if !validID(r.Name) || seen[r.Name] || (r.Target != Entity && r.Target != Property && r.Target != Predicate) {
				return nil, fmt.Errorf("invalid predicate role: %q", d.ID)
			}
			seen[r.Name] = true
		}
		d.Roles = append([]Role(nil), d.Roles...)
		c.byID[d.ID] = d
	}
	return c, nil
}

func (c *Catalog) Lookup(id ID) (Definition, bool) {
	if c == nil {
		return Definition{}, false
	}
	d, ok := c.byID[id]
	d.Roles = append([]Role(nil), d.Roles...)
	return d, ok
}

func (c *Catalog) SortedIDs() []ID {
	if c == nil {
		return nil
	}
	out := make([]ID, 0, len(c.byID))
	for id := range c.byID {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func validID(s string) bool {
	if s == "" || len(s) > 128 || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// Definitions exposes detached declarations for explicit construction of an
// extended catalog. Compiled predicate signatures are not mutated in place.
func (c *Catalog) Definitions() []Definition {
	ids := c.SortedIDs()
	out := make([]Definition, 0, len(ids))
	for _, id := range ids {
		def, _ := c.Lookup(id)
		out = append(out, def)
	}
	return out
}
