package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/mautic-cli/internal/mautic"
)

// FieldType is how a resource field is taken on the command line and encoded
// into the request body.
type FieldType string

const (
	fieldString  FieldType = "string"
	fieldInt     FieldType = "int"
	fieldBool    FieldType = "bool"
	fieldJSON    FieldType = "json"
	fieldStrings FieldType = "strings"
)

// Field is one writable property of a resource.
//
// Every field a caller is likely to want gets a typed flag. That is the whole
// point: an agent must never have to hand-assemble a request body, which is the
// single largest source of confidently-wrong output. The long tail is still
// reachable through --set and --data.
type Field struct {
	// Flag is the command line name, in kebab-case.
	Flag string
	// API is the Mautic property name, usually camelCase.
	API string
	// Type decides parsing and encoding.
	Type FieldType
	// Desc is the flag's help text.
	Desc string
}

// Resource is one Mautic API collection that follows the standard CRUD shape:
//
//	GET    /<path>            list
//	GET    /<path>/<id>       get
//	POST   /<path>/new        create
//	PATCH  /<path>/<id>/edit  update
//	DELETE /<path>/<id>/delete delete
type Resource struct {
	// Name is the command noun, in the plural form the API uses.
	Name    string
	Aliases []string
	// Path is the API path base, without a leading or trailing slash.
	Path string
	// ItemKey and ListKey are the response envelope keys.
	ItemKey string
	ListKey string
	// Short is the one-line help.
	Short string
	// Long, when set, replaces the generated long help.
	Long string
	// Columns are the default table columns, as dotted paths.
	Columns []string
	// Fields are the writable properties.
	Fields []Field
	// IDName describes the positional argument, for help text.
	IDName string

	NoCreate bool
	NoUpdate bool
	NoDelete bool
	NoList   bool
	NoGet    bool

	// Extra adds resource-specific subcommands to the group.
	Extra func(group *cobra.Command, r Resource)
	// Validate, when set, checks a create or update body before it is sent.
	// Use it for things Mautic accepts over the API but then breaks on.
	Validate func(payload map[string]any) error
	// BeforePatch, when set, runs before a PATCH update (not --replace) and may
	// add properties to the body. Use it for properties Mautic resets to their
	// default when a PATCH omits them. It returns a note for stderr, or "".
	BeforePatch func(ctx context.Context, client *mautic.Client, id string, payload map[string]any) (string, error)
	// UpdateNote, when set, is appended to the update command's long help.
	UpdateNote string
}

func (r Resource) idName() string {
	if r.IDName != "" {
		return r.IDName
	}
	return "id"
}

func (r Resource) apiPath(parts ...string) string {
	segments := append([]string{"/" + strings.Trim(r.Path, "/")}, parts...)
	return strings.Join(segments, "/")
}

// resourceCommands builds a command group for every registered resource.
func resourceCommands() []*cobra.Command {
	cmds := make([]*cobra.Command, 0, len(resources))
	for _, r := range resources {
		cmds = append(cmds, r.command())
	}
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Use < cmds[j].Use })
	return cmds
}

func (r Resource) command() *cobra.Command {
	group := &cobra.Command{
		Use:     r.Name,
		Aliases: r.Aliases,
		Short:   r.Short,
		Long:    r.longHelp(),
	}
	if !r.NoList {
		group.AddCommand(r.listCommand())
	}
	if !r.NoGet {
		group.AddCommand(r.getCommand())
	}
	if !r.NoCreate {
		group.AddCommand(r.createCommand())
	}
	if !r.NoUpdate {
		group.AddCommand(r.updateCommand())
	}
	if !r.NoDelete {
		group.AddCommand(r.deleteCommand())
	}
	if r.Extra != nil {
		r.Extra(group, r)
	}
	return group
}

func (r Resource) longHelp() string {
	if r.Long != "" {
		return r.Long
	}
	return fmt.Sprintf("%s\n\nWraps the Mautic /%s endpoints. Fields not covered by a flag can be set with\n--set name=value, and an entire body can be supplied with --data @file.json.",
		r.Short, strings.Trim(r.Path, "/"))
}

// listFlags are the query parameters every collection endpoint accepts.
type listFlags struct {
	search        string
	limit         int
	start         int
	orderBy       string
	orderDir      string
	publishedOnly bool
	minimal       bool
	where         []string
	order         []string
	all           bool
	pageSize      int
}

func (f *listFlags) register(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVar(&f.search, "search", "", "search string or Mautic search command to filter by")
	fl.IntVar(&f.limit, "limit", 30, "maximum number of records to return")
	fl.IntVar(&f.start, "start", 0, "starting row")
	fl.StringVar(&f.orderBy, "order-by", "", "column to sort by, in snake_case (dateIdentified is date_identified)")
	fl.StringVar(&f.orderDir, "order-dir", "", "sort direction: asc or desc")
	fl.BoolVar(&f.publishedOnly, "published-only", false, "return only published records")
	fl.BoolVar(&f.minimal, "minimal", false, "omit the nested lists from each record, which is much faster")
	fl.StringArrayVar(&f.where, "where", nil, "advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull)")
	fl.StringArrayVar(&f.order, "order", nil, "advanced sort as col:dir, repeatable")
	fl.BoolVar(&f.all, "all", false, "fetch every page, not just the first")
	fl.IntVar(&f.pageSize, "page-size", 100, "records per request when --all is set")
}

func (f *listFlags) params() (mautic.ListParams, error) {
	p := mautic.ListParams{
		Search:        f.search,
		Start:         f.start,
		Limit:         f.limit,
		OrderBy:       f.orderBy,
		OrderByDir:    f.orderDir,
		PublishedOnly: f.publishedOnly,
		Minimal:       f.minimal,
	}
	for _, w := range f.where {
		c, err := mautic.ParseCondition(w)
		if err != nil {
			return p, err
		}
		p.Where = append(p.Where, c)
	}
	for _, o := range f.order {
		c, err := mautic.ParseOrder(o)
		if err != nil {
			return p, err
		}
		p.Order = append(p.Order, c)
	}
	return p, nil
}

func (r Resource) listCommand() *cobra.Command {
	var f listFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List " + r.Name,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			params, err := f.params()
			if err != nil {
				return err
			}
			page, err := fetchList(cmd, client, r.apiPath(), params, r.ListKey, f)
			if err != nil {
				return err
			}
			return emitPage(cmd, page, r.Columns)
		},
	}
	f.register(cmd)
	return cmd
}

// fetchList runs one page or every page, depending on --all.
func fetchList(cmd *cobra.Command, client *mautic.Client, path string, params mautic.ListParams, key string, f listFlags) (*mautic.Page, error) {
	if f.all {
		max := 0
		if cmd.Flags().Changed("limit") {
			max = f.limit
		}
		return client.ListAll(cmd.Context(), path, params, key, f.pageSize, max)
	}
	raw, err := client.Get(cmd.Context(), path, params.Values())
	if err != nil {
		return nil, err
	}
	return mautic.DecodePage(raw, key)
}

func (r Resource) getCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <" + r.idName() + ">",
		Short: "Get one record by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			raw, err := client.Get(cmd.Context(), r.apiPath(args[0]), nil)
			if err != nil {
				return err
			}
			return emitObject(cmd, unwrap(raw, r.ItemKey), r.Columns)
		},
	}
}

func (r Resource) createCommand() *cobra.Command {
	body := newBodyFlags(r.Fields)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a " + singular(r.Name),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			payload, err := body.build(cmd)
			if err != nil {
				return err
			}
			if len(payload) == 0 {
				return fmt.Errorf("nothing to create: pass at least one field flag, --set name=value or --data @file.json")
			}
			if r.Validate != nil {
				if err := r.Validate(payload); err != nil {
					return err
				}
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", r.apiPath("new"), nil, payload)
			if err != nil {
				return err
			}
			if client.DryRun() {
				return emitRaw(cmd, raw)
			}
			return emitObject(cmd, unwrap(raw, r.ItemKey), r.Columns)
		},
	}
	body.register(cmd)
	return cmd
}

func (r Resource) updateCommand() *cobra.Command {
	body := newBodyFlags(r.Fields)
	var replace bool
	long := fmt.Sprintf("Update a %s.\n\nBy default this sends PATCH, which changes only the fields you pass. With\n--replace it sends PUT, which replaces the whole record and blanks every field\nyou did not provide.", singular(r.Name))
	if r.UpdateNote != "" {
		long += "\n\n" + r.UpdateNote
	}
	cmd := &cobra.Command{
		Use:   "update <" + r.idName() + ">",
		Short: "Update a " + singular(r.Name),
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			payload, err := body.build(cmd)
			if err != nil {
				return err
			}
			if len(payload) == 0 {
				return fmt.Errorf("nothing to update: pass at least one field flag, --set name=value or --data @file.json")
			}
			if r.Validate != nil {
				if err := r.Validate(payload); err != nil {
					return err
				}
			}
			method := "PATCH"
			if replace {
				method = "PUT"
				if err := confirmed(cmd, "--replace overwrites every field that was not supplied, and"); err != nil {
					return err
				}
			}
			if method == "PATCH" && r.BeforePatch != nil {
				note, err := r.BeforePatch(cmd.Context(), client, args[0], payload)
				if err != nil {
					return err
				}
				if note != "" {
					fmt.Fprintln(cmd.ErrOrStderr(), "note: "+note)
				}
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), method, r.apiPath(args[0], "edit"), nil, payload)
			if err != nil {
				return err
			}
			if client.DryRun() {
				return emitRaw(cmd, raw)
			}
			return emitObject(cmd, unwrap(raw, r.ItemKey), r.Columns)
		},
	}
	body.register(cmd)
	cmd.Flags().BoolVar(&replace, "replace", false, "send PUT instead of PATCH, replacing the whole record (needs --yes)")
	return cmd
}

func (r Resource) deleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <" + r.idName() + ">",
		Short: "Delete a " + singular(r.Name),
		Long:  fmt.Sprintf("Delete a %s.\n\nThis cannot be undone. The command refuses to run without --yes.", singular(r.Name)),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, fmt.Sprintf("deleting %s %s", singular(r.Name), args[0])); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "DELETE", r.apiPath(args[0], "delete"), nil, nil)
			if err != nil {
				return err
			}
			if client.DryRun() {
				return emitRaw(cmd, raw)
			}
			return emitObject(cmd, unwrap(raw, r.ItemKey), r.Columns)
		},
	}
}

// bodyFlags turns a resource's field list into flags and back into a request
// body, plus the generic escape hatches.
type bodyFlags struct {
	fields  []Field
	strings map[string]*string
	ints    map[string]*int
	bools   map[string]*bool
	slices  map[string]*[]string
	set     []string
	setJSON []string
	data    string
}

func newBodyFlags(fields []Field) *bodyFlags {
	return &bodyFlags{
		fields:  fields,
		strings: map[string]*string{},
		ints:    map[string]*int{},
		bools:   map[string]*bool{},
		slices:  map[string]*[]string{},
	}
}

func (b *bodyFlags) register(cmd *cobra.Command) {
	fl := cmd.Flags()
	for _, f := range b.fields {
		switch f.Type {
		case fieldInt:
			v := new(int)
			b.ints[f.Flag] = v
			fl.IntVar(v, f.Flag, 0, f.Desc)
		case fieldBool:
			v := new(bool)
			b.bools[f.Flag] = v
			fl.BoolVar(v, f.Flag, false, f.Desc)
		case fieldStrings:
			v := new([]string)
			b.slices[f.Flag] = v
			fl.StringArrayVar(v, f.Flag, nil, f.Desc+" (repeatable)")
		default: // string and json both arrive as text
			v := new(string)
			b.strings[f.Flag] = v
			fl.StringVar(v, f.Flag, "", f.Desc)
		}
	}
	fl.StringArrayVar(&b.set, "set", nil, "set any other property as name=value, repeatable")
	fl.StringArrayVar(&b.setJSON, "set-json", nil, "set any other property to a JSON value as name=<json>, repeatable")
	fl.StringVar(&b.data, "data", "", "request body as JSON: inline, @file.json, or @- for stdin. Field flags override it")
}

// build assembles the request body. Precedence, lowest first: --data, then the
// typed field flags, then --set / --set-json.
func (b *bodyFlags) build(cmd *cobra.Command) (map[string]any, error) {
	payload := map[string]any{}

	if b.data != "" {
		raw, err := readData(b.data, cmd.InOrStdin())
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("--data is not a JSON object: %w", err)
		}
	}

	for _, f := range b.fields {
		if !cmd.Flags().Changed(f.Flag) {
			continue
		}
		switch f.Type {
		case fieldInt:
			payload[f.API] = *b.ints[f.Flag]
		case fieldBool:
			payload[f.API] = *b.bools[f.Flag]
		case fieldStrings:
			payload[f.API] = *b.slices[f.Flag]
		case fieldJSON:
			var v any
			if err := json.Unmarshal([]byte(*b.strings[f.Flag]), &v); err != nil {
				return nil, fmt.Errorf("--%s expects JSON: %w", f.Flag, err)
			}
			payload[f.API] = v
		default:
			payload[f.API] = *b.strings[f.Flag]
		}
	}

	for _, kv := range b.set {
		name, value, err := splitKV(kv, "--set")
		if err != nil {
			return nil, err
		}
		payload[name] = coerce(value)
	}
	for _, kv := range b.setJSON {
		name, value, err := splitKV(kv, "--set-json")
		if err != nil {
			return nil, err
		}
		var v any
		if err := json.Unmarshal([]byte(value), &v); err != nil {
			return nil, fmt.Errorf("--set-json %s: the value is not valid JSON: %w", name, err)
		}
		payload[name] = v
	}
	return payload, nil
}

func splitKV(kv, flag string) (string, string, error) {
	name, value, ok := strings.Cut(kv, "=")
	if !ok || name == "" {
		return "", "", fmt.Errorf("%s %q must be written as name=value", flag, kv)
	}
	return name, value, nil
}

// coerce keeps --set usable for numbers and booleans without a second flag,
// while leaving anything ambiguous as a string.
//
// A leading zero means the value is an identifier that happens to be digits —
// a Japanese postcode, a phone number, a product code — so it stays a string.
// Turning "0120" into 120 would corrupt the record silently. Use --set-json
// when a number really is wanted.
func coerce(v string) any {
	switch strings.ToLower(v) {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}
	if len(v) > 1 && (v[0] == '0' || (v[0] == '+' && len(v) > 1)) {
		return v
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return v
}

// readData resolves an inline value, @file or @- (stdin). Long or symbol-heavy
// values belong in a file: shell quoting is where agents break.
func readData(spec string, stdin io.Reader) ([]byte, error) {
	if !strings.HasPrefix(spec, "@") {
		return []byte(spec), nil
	}
	path := strings.TrimPrefix(spec, "@")
	if path == "-" {
		return io.ReadAll(stdin)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}
	return data, nil
}

// unwrap pulls the record out of its envelope, so that `get` prints the record
// rather than {"contact": {...}}. An unexpected shape is passed through
// untouched rather than dropped.
func unwrap(raw json.RawMessage, key string) json.RawMessage {
	if key == "" {
		return raw
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return raw
	}
	if inner, ok := envelope[key]; ok {
		return inner
	}
	return raw
}

// singular is a crude de-pluraliser, used only in help text.
func singular(name string) string {
	switch {
	case strings.HasSuffix(name, "ies"):
		return strings.TrimSuffix(name, "ies") + "y"
	case strings.HasSuffix(name, "ses"):
		return strings.TrimSuffix(name, "es")
	case strings.HasSuffix(name, "s"):
		return strings.TrimSuffix(name, "s")
	}
	return name
}
