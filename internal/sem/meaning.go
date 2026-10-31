// Package sem owns the single authoritative, language-neutral Meaning IR.
package sem

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
)

type NodeID string
type NodeKind string

const (
	Referent NodeKind = "referent"
	State    NodeKind = "state"
)

type Argument struct {
	Role   string
	Target NodeID
}

type Node struct {
	ID      NodeID
	Kind    NodeKind
	Concept core.ID
	// Value is a typed concept argument of a predicate, not a surface form.
	Value     core.ID
	Arguments []Argument
	Reference context.Deixis
}

type Meaning struct {
	Root  NodeID
	Nodes []Node
}

func (m Meaning) Clone() Meaning {
	out := Meaning{Root: m.Root, Nodes: make([]Node, len(m.Nodes))}
	for i, n := range m.Nodes {
		out.Nodes[i] = n
		out.Nodes[i].Arguments = append([]Argument(nil), n.Arguments...)
	}
	return out
}

func Validate(m Meaning, catalog *core.Catalog) error {
	if catalog == nil || len(m.Nodes) == 0 || len(m.Nodes) > 64 {
		return fmt.Errorf("missing catalog or invalid meaning size")
	}
	byID := make(map[NodeID]Node, len(m.Nodes))
	for _, n := range m.Nodes {
		if n.ID == "" || len(n.ID) > 80 {
			return fmt.Errorf("invalid occurrence ID")
		}
		if _, exists := byID[n.ID]; exists {
			return fmt.Errorf("duplicate occurrence %q", n.ID)
		}
		def, ok := catalog.Lookup(n.Concept)
		if !ok {
			return fmt.Errorf("unknown concept %q", n.Concept)
		}
		switch n.Kind {
		case Referent:
			if def.Kind != core.Entity || n.Value != "" || len(n.Arguments) != 0 || (n.Reference != context.Unresolved && n.Reference != context.Proximal && n.Reference != context.Distal) {
				return fmt.Errorf("invalid referent %q", n.ID)
			}
		case State:
			if def.Kind != core.Predicate || n.Reference != "" {
				return fmt.Errorf("invalid state %q", n.ID)
			}
			v, ok := catalog.Lookup(n.Value)
			if !ok || v.Kind != def.ValueKind {
				return fmt.Errorf("invalid predicate value at %q", n.ID)
			}
			if len(n.Arguments) != len(def.Roles) {
				return fmt.Errorf("predicate %q has incomplete roles", n.ID)
			}
			seen := map[string]bool{}
			for _, arg := range n.Arguments {
				roleOk := false
				for _, r := range def.Roles {
					if r.Name == arg.Role {
						roleOk = true
						break
					}
				}
				if !roleOk || seen[arg.Role] {
					return fmt.Errorf("unlicensed/duplicate role %q", arg.Role)
				}
				seen[arg.Role] = true
			}
		default:
			return fmt.Errorf("unsupported node kind %q", n.Kind)
		}
		byID[n.ID] = n
	}
	root, ok := byID[m.Root]
	if !ok || root.Kind != State {
		return fmt.Errorf("invalid meaning root")
	}
	reachable := map[NodeID]bool{m.Root: true}
	for _, n := range m.Nodes {
		if n.Kind != State {
			continue
		}
		def, _ := catalog.Lookup(n.Concept)
		for _, a := range n.Arguments {
			target, ok := byID[a.Target]
			if !ok {
				return fmt.Errorf("dangling role %q", a.Target)
			}
			for _, r := range def.Roles {
				if r.Name == a.Role {
					if r.Target == core.Entity && target.Kind != Referent {
						return fmt.Errorf("invalid target kind for %q", a.Role)
					}
					if r.Target == core.Predicate && target.Kind != State {
						return fmt.Errorf("invalid target kind for %q", a.Role)
					}
				}
			}
			reachable[a.Target] = true
		}
	}
	if len(reachable) != len(m.Nodes) {
		return fmt.Errorf("orphaned occurrence")
	}
	return nil
}

// CanonicalKey retains role bindings and repeated-occurrence identity while
// disregarding arbitrary local occurrence IDs and node ordering.
func CanonicalKey(m Meaning, catalog *core.Catalog) (string, error) {
	if err := Validate(m, catalog); err != nil {
		return "", err
	}
	byID := map[NodeID]Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	names := map[NodeID]int{}
	var visit func(NodeID) string
	visit = func(id NodeID) string {
		if v, ok := names[id]; ok {
			return fmt.Sprintf("@%d", v)
		}
		names[id] = len(names) + 1
		n := byID[id]
		b := strings.Builder{}
		fmt.Fprintf(&b, "(%d,%s,%s,%s,%s", names[id], n.Kind, n.Concept, n.Value, n.Reference)
		args := append([]Argument(nil), n.Arguments...)
		sort.Slice(args, func(i, j int) bool { return args[i].Role < args[j].Role })
		for _, a := range args {
			fmt.Fprintf(&b, ",%s:%s", a.Role, visit(a.Target))
		}
		b.WriteByte(')')
		return b.String()
	}
	return visit(m.Root), nil
}
func Equivalent(a, b Meaning, c *core.Catalog) (bool, error) {
	x, e := CanonicalKey(a, c)
	if e != nil {
		return false, e
	}
	y, e := CanonicalKey(b, c)
	return e == nil && x == y, e
}
