// Command golingua2 exposes the compiled A1 semantic bridge.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/status"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

const version = "0.1.0-m1"

func usage() {
	fmt.Fprintln(os.Stdout, "GoLingua 2 — compiled procedural PL ↔ Semantic IR ↔ EN")
	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  golingua2 translate --from pl --to en --text 'To jest kot.' --near [--json]")
	fmt.Fprintln(os.Stdout, "  golingua2 analyze --from en --text 'This cat is black.' [--json]")
	fmt.Fprintln(os.Stdout, "  golingua2 --version")
	fmt.Fprintln(os.Stdout, "M1: classification and demonstrative properties; A1 scenarios 001–020 only.")
}
func run() int {
	if len(os.Args) <= 1 || os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		usage()
		return 0
	}
	if os.Args[1] == "--version" || os.Args[1] == "-version" {
		fmt.Println(version)
		return 0
	}
	command := os.Args[1]
	if command != "analyze" && command != "translate" {
		usage()
		return 2
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	from := fs.String("from", "", "source language: pl or en")
	to := fs.String("to", "", "target language: pl or en")
	text := fs.String("text", "", "source sentence")
	near := fs.Bool("near", false, "explicit proximal/deictic context for Polish 'to/ten/ta'")
	jsonOut := fs.Bool("json", false, "display JSON result (no JSON sentence fixtures are used)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return 2
	}
	if *text == "" || *from == "" || (command == "translate" && *to == "") {
		fmt.Fprintln(os.Stderr, "missing --from/--to/--text")
		return 2
	}
	catalog, err := core.BaseA1()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e, err := engine.New(catalog, pl.Definition(), en.Definition())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	ctx := context.Request{Deixis: context.Unresolved}
	if *near {
		ctx.Deixis = context.Proximal
	}
	var result engine.Result
	if command == "analyze" {
		result = e.Analyze(*from, *text, ctx)
	} else {
		result = e.Translate(*from, *to, *text, ctx)
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		fmt.Println("outcome:", result.Outcome)
		for _, candidate := range result.Candidates {
			if command == "analyze" {
				b, _ := json.Marshal(candidate.Meaning)
				fmt.Println(string(b))
			} else {
				fmt.Println(candidate.Text)
			}
		}
		for _, diag := range result.Diagnostics {
			fmt.Fprintln(os.Stderr, diag.Code, diag.Message)
		}
	}
	if result.Outcome == status.OK || result.Outcome == status.Ambiguous {
		return 0
	}
	return 3
}
func main() { os.Exit(run()) }
