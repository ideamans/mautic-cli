package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// A contact record nests its fields four levels deep, with a flattened copy
// under fields.all. The dotted-path flattening is what makes
// --fields id,fields.all.email work.
func TestFlattenDottedPaths(t *testing.T) {
	raw := json.RawMessage(`{
	  "id": 47,
	  "points": 10,
	  "owner": {"id": 1, "username": "joe"},
	  "fields": {"all": {"email": "a@example.com", "firstname": "A"}},
	  "tags": ["vip", "jp"],
	  "ipAddresses": {}
	}`)
	flat := flatten(raw)

	cases := map[string]string{
		"id":               "47",
		"points":           "10",
		"owner.username":   "joe",
		"fields.all.email": "a@example.com",
		"tags":             "vip,jp",
		"ipAddresses":      "{}",
	}
	for path, want := range cases {
		if got := flat[path]; got != want {
			t.Errorf("flatten[%q] = %q, want %q", path, got, want)
		}
	}
}

// Whole numbers must not come back as 47.000000, and a null must be empty
// rather than the string "null" — a CSV of "null" reads as data.
func TestScalarFormatting(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{float64(47), "47"},
		{float64(1.5), "1.5"},
		{nil, ""},
		{true, "true"},
		{"x", "x"},
	}
	for _, tc := range cases {
		if got := scalar(tc.in); got != tc.want {
			t.Errorf("scalar(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Byte-slicing a Japanese string mid-rune produces mojibake, and byte-width
// padding produces a ragged table. Mautic instances are full of Japanese.
func TestTruncateCellIsRuneSafe(t *testing.T) {
	long := strings.Repeat("あ", 60)
	got := truncateCell(long)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected an ellipsis, got %q", got)
	}
	for _, r := range got {
		if r == '�' {
			t.Fatalf("truncation produced an invalid rune: %q", got)
		}
	}
	// Newlines would break the one-record-per-line contract.
	if strings.Contains(truncateCell("a\nb"), "\n") {
		t.Error("newlines must be flattened out of a cell")
	}
}

func TestSelectedColumnsPrefersExplicitFields(t *testing.T) {
	defer resetFlags()
	rows := []map[string]string{{"id": "1", "name": "x"}}

	flagFields = "id, name , "
	if got := selectedColumns([]string{"other"}, rows); len(got) != 2 || got[0] != "id" || got[1] != "name" {
		t.Errorf("--fields should win and be trimmed, got %v", got)
	}

	// A default column that no record carries would render as an empty column,
	// hiding the data that is actually there.
	flagFields = ""
	if got := selectedColumns([]string{"id", "missing"}, rows); len(got) != 1 || got[0] != "id" {
		t.Errorf("absent default columns should be dropped, got %v", got)
	}
}

func resetFlags() {
	flagFields = ""
	flagFormat = formatTable
}

// The resource table drives the whole tree, so a typo in it is a broken
// command group rather than a compile error.
func TestResourceTableIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range resources {
		if r.Name == "" || r.Path == "" {
			t.Errorf("resource %+v is missing a name or path", r)
			continue
		}
		if seen[r.Name] {
			t.Errorf("duplicate resource name %q", r.Name)
		}
		seen[r.Name] = true

		if !r.NoGet && r.ItemKey == "" {
			t.Errorf("%s: get is enabled but ItemKey is empty, so records will not be unwrapped", r.Name)
		}
		if !r.NoList && r.ListKey == "" {
			t.Errorf("%s: list is enabled but ListKey is empty", r.Name)
		}
		if strings.HasPrefix(r.Path, "/") || strings.HasSuffix(r.Path, "/") {
			t.Errorf("%s: Path %q must not have leading or trailing slashes", r.Name, r.Path)
		}
		for _, f := range r.Fields {
			if f.Flag == "" || f.API == "" || f.Desc == "" {
				t.Errorf("%s: field %+v is incomplete", r.Name, f)
			}
			if strings.Contains(f.Flag, "_") {
				t.Errorf("%s: flag %q should be kebab-case", r.Name, f.Flag)
			}
		}
	}
}

// Credentials must stay overridable on every command, so no leaf may define a
// local flag that shadows a global one.
func TestNoCommandShadowsAGlobalFlag(t *testing.T) {
	root := Root()
	globals := map[string]bool{}
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) { globals[f.Name] = true })

	var walk func(c *cobra.Command, path string)
	walk = func(c *cobra.Command, path string) {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			// Inherited flags show up here too; only locally defined ones are
			// a problem.
			if c.LocalFlags().Lookup(f.Name) == nil {
				return
			}
			if globals[f.Name] {
				t.Errorf("%s defines --%s, which shadows the global flag of the same name", path, f.Name)
			}
		})
		for _, sub := range c.Commands() {
			walk(sub, path+" "+sub.Name())
		}
	}
	for _, sub := range root.Commands() {
		walk(sub, sub.Name())
	}
}

// Every leaf that reaches the API must be runnable; a missing RunE is a group
// that silently does nothing.
func TestEveryLeafCommandRuns(t *testing.T) {
	var walk func(c *cobra.Command, path string)
	walk = func(c *cobra.Command, path string) {
		subs := c.Commands()
		if len(subs) == 0 {
			if c.RunE == nil && c.Run == nil {
				t.Errorf("%s is a leaf command with nothing to run", path)
			}
			return
		}
		for _, sub := range subs {
			walk(sub, path+" "+sub.Name())
		}
	}
	walk(Root(), "mautic")
}

func TestCoerceKeepsAmbiguousValuesAsStrings(t *testing.T) {
	cases := map[string]any{
		"true":  true,
		"false": false,
		"42":    42,
		"0":     0,
		// A leading zero means an identifier that happens to be digits — a
		// Japanese postcode, a phone number. Sending 120 for "0120" would
		// corrupt the record with no error.
		"0120":          "0120",
		"03-1234-5678":  "03-1234-5678",
		"+819012345678": "+819012345678",
		"abc":           "abc",
		"":              "",
	}
	for in, want := range cases {
		if got := coerce(in); got != want {
			t.Errorf("coerce(%q) = %#v, want %#v", in, got, want)
		}
	}
}
