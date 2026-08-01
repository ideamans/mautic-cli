package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"

	"github.com/ideamans/mautic-cli/internal/mautic"
)

// Output formats. table is for humans and may truncate cells; anything being
// parsed must ask for json, jsonl or csv.
const (
	formatTable = "table"
	formatJSON  = "json"
	formatJSONL = "jsonl"
	formatCSV   = "csv"
)

// maxCellWidth bounds a table cell. Contact records carry HTML bodies and
// 2,000-character notes; printing them whole makes the table unreadable.
const maxCellWidth = 48

// listEnvelope is the shape `--format json` produces for every collection, so
// that a caller can rely on one structure across all resources.
//
// total is what Mautic reports for the whole collection; count is how many
// records this invocation actually returned. Comparing the two before drawing a
// conclusion is the difference between "we have 12 contacts" and "we fetched
// the first 12 of 4,318".
type listEnvelope struct {
	Total int               `json:"total"`
	Count int               `json:"count"`
	Items []json.RawMessage `json:"items"`
}

func validateFormat() error {
	switch flagFormat {
	case formatTable, formatJSON, formatJSONL, formatCSV:
		return nil
	default:
		return fmt.Errorf("unknown --format %q; use table, json, jsonl or csv", flagFormat)
	}
}

// emitRaw writes a JSON document through unchanged. Used by `mautic api` and by
// dry-run output, where the point is to show exactly what the API said.
func emitRaw(cmd *cobra.Command, raw json.RawMessage) error {
	var pretty json.RawMessage
	buf := new(strings.Builder)
	if err := indentJSON(buf, raw); err != nil {
		pretty = raw
		_, err := fmt.Fprintln(cmd.OutOrStdout(), string(pretty))
		return err
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), buf.String())
	return err
}

func indentJSON(b *strings.Builder, raw json.RawMessage) error {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	enc := json.NewEncoder(b)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	// Encode appends a newline; the caller adds its own.
	s := strings.TrimRight(b.String(), "\n")
	b.Reset()
	b.WriteString(s)
	return nil
}

// emitObject renders one record.
func emitObject(cmd *cobra.Command, raw json.RawMessage, defaultCols []string) error {
	if err := validateFormat(); err != nil {
		return err
	}
	switch flagFormat {
	case formatJSON:
		return emitRaw(cmd, raw)
	case formatJSONL, formatCSV:
		return emitItems(cmd, []json.RawMessage{raw}, defaultCols, 1)
	}

	// table: a two-column key/value view reads far better than a one-row table
	// for a single record.
	flat := flatten(raw)
	keys := selectedColumns(defaultCols, []map[string]string{flat})
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{k, truncateCell(flat[k])})
	}
	return writeTable(cmd.OutOrStdout(), nil, rows)
}

// writeTable pads columns by display width rather than by byte count.
//
// text/tabwriter counts bytes, so a table containing Japanese — which this CLI
// will be full of, since Mautic instances are full of it — comes out ragged.
// runewidth also gives correct truncation, which byte slicing does not: cutting
// a multi-byte rune in half produces mojibake, not a shorter string.
func writeTable(w io.Writer, header []string, rows [][]string) error {
	if len(rows) == 0 && len(header) == 0 {
		return nil
	}
	cols := len(header)
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	widths := make([]int, cols)
	measure := func(cells []string) {
		for i, c := range cells {
			if n := runewidth.StringWidth(c); n > widths[i] {
				widths[i] = n
			}
		}
	}
	if len(header) > 0 {
		measure(header)
	}
	for _, r := range rows {
		measure(r)
	}

	writeRow := func(cells []string) error {
		var b strings.Builder
		for i, c := range cells {
			if i > 0 {
				b.WriteString("  ")
			}
			// The last column needs no padding, and padding it would leave
			// trailing whitespace on every line.
			if i == len(cells)-1 {
				b.WriteString(c)
				continue
			}
			b.WriteString(runewidth.FillRight(c, widths[i]))
		}
		_, err := fmt.Fprintln(w, b.String())
		return err
	}

	if len(header) > 0 {
		if err := writeRow(header); err != nil {
			return err
		}
	}
	for _, r := range rows {
		if err := writeRow(r); err != nil {
			return err
		}
	}
	return nil
}

// emitPage renders a collection.
func emitPage(cmd *cobra.Command, page *mautic.Page, defaultCols []string) error {
	if err := validateFormat(); err != nil {
		return err
	}
	if flagFormat == formatJSON {
		env := listEnvelope{Total: page.Total, Count: len(page.Items), Items: page.Items}
		if env.Items == nil {
			env.Items = []json.RawMessage{}
		}
		data, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	}
	return emitItems(cmd, page.Items, defaultCols, page.Total)
}

func emitItems(cmd *cobra.Command, items []json.RawMessage, defaultCols []string, total int) error {
	out := cmd.OutOrStdout()

	if flagFormat == formatJSONL {
		for _, item := range items {
			compact := new(bytes.Buffer)
			if err := json.Compact(compact, item); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(out, compact.String()); err != nil {
				return err
			}
		}
		return nil
	}

	rows := make([]map[string]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, flatten(item))
	}
	cols := selectedColumns(defaultCols, rows)

	switch flagFormat {
	case formatCSV:
		w := csv.NewWriter(out)
		if err := w.Write(cols); err != nil {
			return err
		}
		for _, row := range rows {
			rec := make([]string, len(cols))
			for i, c := range cols {
				rec[i] = row[c]
			}
			if err := w.Write(rec); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	default:
		if len(items) == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "no records matched")
			return nil
		}
		table := make([][]string, 0, len(rows))
		for _, row := range rows {
			cells := make([]string, len(cols))
			for i, c := range cols {
				cells[i] = truncateCell(row[c])
			}
			table = append(table, cells)
		}
		if err := writeTable(out, cols, table); err != nil {
			return err
		}
		// The count line goes to stderr so that stdout stays parseable even
		// when someone pipes the table anyway.
		if total > len(items) {
			fmt.Fprintf(cmd.ErrOrStderr(), "showing %d of %d — use --limit or --all for more\n", len(items), total)
		}
		return nil
	}
}

// selectedColumns decides which columns to show: --fields wins, then the
// resource's curated defaults, then whatever the records actually carry.
func selectedColumns(defaults []string, rows []map[string]string) []string {
	if flagFields != "" {
		parts := strings.Split(flagFields, ",")
		cols := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				cols = append(cols, p)
			}
		}
		if len(cols) > 0 {
			return cols
		}
	}
	if len(defaults) > 0 {
		// Keep a default column only when at least one record has it; a table
		// full of empty columns hides the data that is there.
		present := make([]string, 0, len(defaults))
		for _, c := range defaults {
			for _, row := range rows {
				if _, ok := row[c]; ok {
					present = append(present, c)
					break
				}
			}
		}
		if len(present) > 0 {
			return present
		}
	}
	return inferColumns(rows)
}

// inferColumns picks a readable subset when nothing else specified one.
func inferColumns(rows []map[string]string) []string {
	seen := map[string]int{}
	for _, row := range rows {
		for k := range row {
			seen[k]++
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		// Deeply nested paths make terrible columns.
		if strings.Count(k, ".") > 1 {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		// Most-populated first, then alphabetical, so the table is stable.
		if seen[keys[i]] != seen[keys[j]] {
			return seen[keys[i]] > seen[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > 8 {
		keys = keys[:8]
	}
	// id first when present; everything else alphabetical for determinism.
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] == "id" && keys[j] != "id"
	})
	return keys
}

// flatten turns a record into dotted-path -> printable value pairs.
func flatten(raw json.RawMessage) map[string]string {
	out := map[string]string{}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		out["value"] = strings.TrimSpace(string(raw))
		return out
	}
	walk("", v, out)
	return out
}

func walk(prefix string, v any, out map[string]string) {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 0 {
			out[prefix] = "{}"
			return
		}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			walk(joinPath(prefix, k), t[k], out)
		}
	case []any:
		if len(t) == 0 {
			out[prefix] = "[]"
			return
		}
		// Arrays of scalars read better joined; arrays of objects get indexed.
		scalars := make([]string, 0, len(t))
		allScalar := true
		for _, item := range t {
			switch item.(type) {
			case map[string]any, []any:
				allScalar = false
			default:
				scalars = append(scalars, scalar(item))
			}
		}
		if allScalar {
			out[prefix] = strings.Join(scalars, ",")
			return
		}
		for i, item := range t {
			walk(joinPath(prefix, strconv.Itoa(i)), item, out)
		}
	default:
		out[prefix] = scalar(v)
	}
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

// truncateCell bounds a cell by display width, cutting on rune boundaries.
func truncateCell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if runewidth.StringWidth(s) <= maxCellWidth {
		return s
	}
	return runewidth.Truncate(s, maxCellWidth, "…")
}
