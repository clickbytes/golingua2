// Package context holds explicit request-local information, never persistent world facts.
package context

type Deixis string

const (
	Unresolved Deixis = "unresolved"
	Proximal   Deixis = "proximal"
	Distal     Deixis = "distal"
)

type Request struct{ Deixis Deixis }

func (r Request) Valid() bool {
	return r.Deixis == Unresolved || r.Deixis == Proximal || r.Deixis == Distal || r.Deixis == ""
}
