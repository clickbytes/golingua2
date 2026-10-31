// Package status defines honest, typed pipeline outcomes.
package status

type Outcome string

const (
	OK            Outcome = "ok"
	Ambiguous     Outcome = "ambiguous"
	Incomplete    Outcome = "incomplete"
	Unsupported   Outcome = "unsupported"
	Invalid       Outcome = "invalid"
	ResourceLimit Outcome = "resource_limit"
)

type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}
