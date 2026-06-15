package widget

import (
	"errors"
	"strings"
	"testing"

	es "github.com/evanw/esbuild/pkg/api"
)

func TestProcessResult(t *testing.T) {
	// No messages: success.
	if err := processResult(es.BuildResult{}); err != nil {
		t.Fatalf("empty result: %v", err)
	}

	// Warnings only: still success.
	warnOnly := es.BuildResult{Warnings: []es.Message{{Text: "careful"}}}
	if err := processResult(warnOnly); err != nil {
		t.Fatalf("warnings only: %v", err)
	}

	// Any error fails with ErrBuild.
	withErr := es.BuildResult{Errors: []es.Message{{Text: "boom"}}}
	if err := processResult(withErr); !errors.Is(err, ErrBuild) {
		t.Fatalf("error result = %v, want ErrBuild", err)
	}
}

func TestNormalizeAndSort(t *testing.T) {
	result := es.BuildResult{
		Errors: []es.Message{
			{Text: "third", Location: &es.Location{Line: 3, Column: 1}},
			{Text: "first", Location: &es.Location{Line: 1, Column: 5}},
		},
		Warnings: []es.Message{
			{Text: "second", Location: &es.Location{Line: 1, Column: 9}},
		},
	}

	diags := normalizeAndSort(result)
	if len(diags) != 3 {
		t.Fatalf("got %d diagnostics, want 3", len(diags))
	}

	// Sorted by line then column: (1,5), (1,9), (3,1).
	wantOrder := []string{"first", "second", "third"}
	for i, w := range wantOrder {
		if !strings.HasSuffix(diags[i].message, w) {
			t.Errorf("diags[%d].message = %q, want suffix %q", i, diags[i].message, w)
		}
	}
}

func TestNormalizeMessage(t *testing.T) {
	// Without a location, only the lower-cased text is used.
	d := normalizeMessage(es.Message{Text: "Bad thing"}, severityError)
	if d.message != "bad thing" || d.severity != severityError {
		t.Errorf("got %+v, want lowercased error message", d)
	}

	// With a location, the message is prefixed and line/column are recorded.
	d = normalizeMessage(es.Message{
		Text:     "Unexpected token",
		Location: &es.Location{File: "a.js", Line: 3, Column: 7},
	}, severityWarning)
	if d.message != "a.js:3:7: unexpected token" {
		t.Errorf("message = %q", d.message)
	}
	if d.line != 3 || d.column != 7 {
		t.Errorf("line/column = %d/%d, want 3/7", d.line, d.column)
	}
}

func TestFormatLocation(t *testing.T) {
	loc := es.Location{File: "main.js", Line: 12, Column: 4}
	if got := formatLocation(loc); got != "main.js:12:4" {
		t.Errorf("formatLocation = %q, want main.js:12:4", got)
	}
	if got := formatLocation(es.Location{}); got != "(unknown)" {
		t.Errorf("formatLocation empty = %q, want (unknown)", got)
	}
}

func TestLowerFirst(t *testing.T) {
	cases := map[string]string{
		"":      "",
		"Hello": "hello",
		"ABC":   "aBC",
		"x":     "x",
	}
	for in, want := range cases {
		if got := lowerFirst(in); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", in, got, want)
		}
	}
}
