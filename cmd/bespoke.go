package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// fields — /fields/{object}
//
// Custom fields hang off an object type rather than a fixed path, so this group
// carries --object instead of being two separate nouns. contact is the default
// because it is what almost every question is about.
// ---------------------------------------------------------------------------

var fieldObject string

var fieldFields = []Field{
	{Flag: "label", API: "label", Type: fieldString, Desc: "human readable label (required)"},
	{Flag: "alias", API: "alias", Type: fieldString, Desc: "field alias used in the API and in tokens; derived from the label when omitted"},
	{Flag: "type", API: "type", Type: fieldString, Desc: "field type: text, textarea, email, number, tel, url, select, multiselect, boolean, date, datetime, time, lookup, country, region, timezone, locale"},
	{Flag: "group", API: "group", Type: fieldString, Desc: "which tab the field appears on: core, social, personal or professional"},
	{Flag: "object", API: "object", Type: fieldString, Desc: "the object the field belongs to: lead or company"},
	{Flag: "default-value", API: "defaultValue", Type: fieldString, Desc: "default value"},
	{Flag: "required", API: "isRequired", Type: fieldBool, Desc: "make the field required"},
	{Flag: "unique-identifier", API: "isUniqueIdentifier", Type: fieldBool, Desc: "use the field to deduplicate records"},
	{Flag: "publicly-updatable", API: "isPubliclyUpdatable", Type: fieldBool, Desc: "let tracked visitors update it themselves"},
	{Flag: "properties", API: "properties", Type: fieldJSON, Desc: `type-specific options as JSON, e.g. {"list":[{"label":"A","value":"a"}]}`},
	{Flag: "order", API: "order", Type: fieldInt, Desc: "display order"},
	{Flag: "published", API: "isPublished", Type: fieldBool, Desc: "published state"},
}

func fieldsPath(parts ...string) string {
	segments := append([]string{"/fields/" + fieldObject}, parts...)
	return strings.Join(segments, "/")
}

func newFieldsCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "fields",
		Short: "Custom field definitions for contacts and companies",
		Long: `Custom field definitions for contacts and companies.

Run this first when a write fails validation: contact fields are configured per
instance, and "mautic contacts create --set <alias>=<value>" needs the alias
exactly as it appears here.

Use --object company for company fields.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			switch fieldObject {
			case "contact", "company":
				return nil
			default:
				return fmt.Errorf("--object must be contact or company, got %q", fieldObject)
			}
		},
	}
	group.PersistentFlags().StringVar(&fieldObject, "object", "contact", "which object's fields to work with: contact or company")

	cols := []string{"id", "alias", "label", "type", "group", "isRequired", "isUniqueIdentifier"}

	var lf listFlags
	list := &cobra.Command{
		Use:   "list",
		Short: "List field definitions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			params, err := lf.params()
			if err != nil {
				return err
			}
			page, err := fetchList(cmd, client, fieldsPath(), params, "fields", lf)
			if err != nil {
				return err
			}
			return emitPage(cmd, page, cols)
		},
	}
	lf.register(list)

	get := &cobra.Command{
		Use:   "get <field-id>",
		Short: "Get one field definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			raw, err := client.Get(cmd.Context(), fieldsPath(args[0]), nil)
			if err != nil {
				return err
			}
			return emitObject(cmd, unwrap(raw, "field"), cols)
		},
	}

	createBody := newBodyFlags(fieldFields)
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a field definition",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeField(cmd, createBody, "POST", fieldsPath("new"), cols)
		},
	}
	createBody.register(create)

	updateBody := newBodyFlags(fieldFields)
	update := &cobra.Command{
		Use:   "update <field-id>",
		Short: "Update a field definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeField(cmd, updateBody, "PATCH", fieldsPath(args[0], "edit"), cols)
		},
	}
	updateBody.register(update)

	del := &cobra.Command{
		Use:   "delete <field-id>",
		Short: "Delete a field definition",
		Long: `Delete a field definition.

This drops the column and every value stored in it across all records. It cannot
be undone, and the command refuses to run without --yes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "deleting field "+args[0]+" and every value stored in it"); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "DELETE", fieldsPath(args[0], "delete"), nil, nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	group.AddCommand(list, get, create, update, del)
	return group
}

func writeField(cmd *cobra.Command, body *bodyFlags, method, path string, cols []string) error {
	client, creds, err := newClient(cmd.Context())
	if err != nil {
		return err
	}
	payload, err := body.build(cmd)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return fmt.Errorf("nothing to write: pass at least one field flag, --set name=value or --data @file.json")
	}
	noteTarget(cmd, creds)
	raw, err := client.Do(cmd.Context(), method, path, nil, payload)
	if err != nil {
		return err
	}
	if client.DryRun() {
		return emitRaw(cmd, raw)
	}
	return emitObject(cmd, unwrap(raw, "field"), cols)
}

// ---------------------------------------------------------------------------
// files — /files/{dir}
// ---------------------------------------------------------------------------

func newFilesCommand() *cobra.Command {
	var dir, subdir string

	group := &cobra.Command{
		Use:   "files",
		Short: "Files in Mautic's media directories",
		Long: `Files in Mautic's media directories.

--dir images covers media/images and supports subdirectories; --dir assets
covers media/files, where asset files live, and has no subdirectories.

Uploading a file does not create an Asset. Upload first, then point
"mautic assets create --file <name>" at the uploaded file name.`,
	}
	pf := group.PersistentFlags()
	pf.StringVar(&dir, "dir", "images", "which directory: images or assets")
	pf.StringVar(&subdir, "subdir", "", "subdirectory under images (assets does not support them)")

	group.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if dir != "images" && dir != "assets" {
			return fmt.Errorf("--dir must be images or assets, got %q", dir)
		}
		if subdir != "" && dir != "images" {
			return fmt.Errorf("--subdir only applies to --dir images; assets has no subdirectories")
		}
		return nil
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List files in a media directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			if subdir != "" {
				q.Set("subdir", subdir)
			}
			raw, err := client.Get(cmd.Context(), "/files/"+dir, q)
			if err != nil {
				return err
			}
			if flagFormat == formatJSON {
				return emitRaw(cmd, raw)
			}
			page, err := decodeNamed(raw, "files")
			if err != nil {
				return err
			}
			return emitPage(cmd, page, nil)
		},
	}

	upload := &cobra.Command{
		Use:   "upload <path>",
		Short: "Upload a file to a media directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			target := dir
			if subdir != "" {
				target = dir + "/" + subdir
			}
			noteTarget(cmd, creds)
			raw, err := client.Upload(cmd.Context(), "/files/"+target+"/new", "file", args[0], nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	del := &cobra.Command{
		Use:   "delete <filename>",
		Short: "Delete a file from a media directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "deleting file "+args[0]); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			target := dir
			if subdir != "" {
				target = dir + "/" + subdir
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "DELETE", "/files/"+target+"/"+args[0]+"/delete", nil, nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	group.AddCommand(list, upload, del)
	return group
}

// ---------------------------------------------------------------------------
// themes — /themes
// ---------------------------------------------------------------------------

func newThemesCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "themes",
		Short: "Installed themes",
		Long: `Installed themes.

Themes are identified by name rather than by numeric ID. "get" returns the
theme's config; "download" fetches the theme as a zip, which is what the API
actually serves for that path.`,
	}

	list := subGet("list", "List installed themes and their config", 0,
		func([]string) string { return "/themes" }, "themes", nil)

	var output string
	download := &cobra.Command{
		Use:   "download <theme-name>",
		Short: "Download a theme as a zip archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			path := output
			if path == "" {
				path = args[0] + ".zip"
			}
			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("could not create %s: %w", path, err)
			}
			defer f.Close()
			n, err := client.Download(cmd.Context(), "/themes/"+args[0], f)
			if err != nil {
				os.Remove(path)
				return err
			}
			// The archive is the result, but it went to a file — so the path
			// goes to stdout and the size note to stderr.
			fmt.Fprintf(cmd.ErrOrStderr(), "wrote %d bytes\n", n)
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	download.Flags().StringVar(&output, "output", "", "where to write the zip (default: <theme-name>.zip)")

	upload := &cobra.Command{
		Use:   "create <zip-path>",
		Short: "Install a theme from a zip archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ext := strings.ToLower(filepath.Ext(args[0])); ext != ".zip" {
				return fmt.Errorf("a theme must be a .zip archive, got %s", args[0])
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Upload(cmd.Context(), "/themes/new", "file", args[0], nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	del := &cobra.Command{
		Use:   "delete <theme-name>",
		Short: "Delete an installed theme",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "deleting theme "+args[0]); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "DELETE", "/themes/"+args[0]+"/delete", nil, nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	group.AddCommand(list, subGet("get <theme-name>", "Show one theme's config", 1,
		func(a []string) string { return "/themes/" + a[0] }, "", nil), download, upload, del)
	return group
}

// ---------------------------------------------------------------------------
// stats — /stats
//
// This is the closest thing Mautic has to raw table access, and on instances
// where /contacts misbehaves it is often the only thing that works. It is also
// the fastest way to answer "how many page hits yesterday" without pulling
// every contact.
// ---------------------------------------------------------------------------

func newStatsCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "stats",
		Short: "Query Mautic's statistics tables directly",
		Long: `Query Mautic's statistics tables directly.

"tables" lists what can be queried and the columns of each table; "query" reads
one of them with the same where/order/limit parameters as any other list.

Column names here are database columns in snake_case (date_hit, lead_id), not
the camelCase property names used elsewhere in the API.`,
	}

	tables := &cobra.Command{
		Use:   "tables",
		Short: "List the queryable tables and their columns",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			raw, err := client.Get(cmd.Context(), "/stats", nil)
			if err != nil {
				return err
			}
			if flagFormat != formatTable {
				return emitRaw(cmd, raw)
			}
			var envelope struct {
				AvailableTables []string            `json:"availableTables"`
				TableColumns    map[string][]string `json:"tableColumns"`
			}
			if err := unmarshalStrict(raw, &envelope); err != nil {
				return emitRaw(cmd, raw)
			}
			out := cmd.OutOrStdout()
			for _, t := range envelope.AvailableTables {
				fmt.Fprintf(out, "%s\n", t)
				if cols := envelope.TableColumns[t]; len(cols) > 0 {
					fmt.Fprintf(out, "  %s\n", strings.Join(cols, ", "))
				}
			}
			return nil
		},
	}

	var lf listFlags
	query := &cobra.Command{
		Use:   "query <table>",
		Short: "Read rows from a statistics table",
		Args:  cobra.ExactArgs(1),
		Example: `  # page hits since a date, newest first
  mautic stats query page_hits --where date_hit:gte:2026-08-01 --order date_hit:desc --limit 50

  # every email send recorded for one contact
  mautic stats query email_stats --where lead_id:eq:1234 --all --format csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			params, err := lf.params()
			if err != nil {
				return err
			}
			page, err := fetchList(cmd, client, "/stats/"+args[0], params, "stats", lf)
			if err != nil {
				return err
			}
			return emitPage(cmd, page, nil)
		},
	}
	lf.register(query)

	group.AddCommand(tables, query)
	return group
}

// ---------------------------------------------------------------------------
// data — /data
// ---------------------------------------------------------------------------

func newDataCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "data",
		Short: "Pre-built chart datasets, as used by Mautic's dashboard widgets",
		Long: `Pre-built chart datasets, as used by Mautic's dashboard widgets.

These return aggregates that would otherwise mean paging through the stats
tables and summing by hand — prefer them when the question is "how many, over
time".`,
	}

	types := subGet("types", "List the available datasets", 0,
		func([]string) string { return "/data" }, "types", nil)

	var dateFrom, dateTo, timeUnit, dataFormat string
	var limit, offset int
	var filters, datasets []string
	var withCounts bool

	get := &cobra.Command{
		Use:   "get <type>",
		Short: "Fetch one dataset",
		Args:  cobra.ExactArgs(1),
		Example: `  mautic data get emails.in.time --date-from 2026-07-01 --date-to 2026-07-31 --time-unit d \
    --dataset sent --dataset opened --with-counts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			setIf := func(k, v string) {
				if v != "" {
					q.Set(k, v)
				}
			}
			setIf("dateFrom", dateFrom)
			setIf("dateTo", dateTo)
			setIf("timeUnit", timeUnit)
			setIf("dataFormat", dataFormat)
			if limit > 0 {
				q.Set("limit", strconv.Itoa(limit))
			}
			if offset > 0 {
				q.Set("offset", strconv.Itoa(offset))
			}
			if withCounts {
				q.Set("withCounts", "true")
			}
			for _, d := range datasets {
				q.Add("dataset[]", d)
			}
			for _, f := range filters {
				name, value, ok := strings.Cut(f, "=")
				if !ok {
					return fmt.Errorf("--filter %q must be written as name=value, e.g. segmentId=3", f)
				}
				q.Set("filter["+name+"]", value)
			}
			raw, err := client.Get(cmd.Context(), "/data/"+args[0], q)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	fl := get.Flags()
	fl.StringVar(&dateFrom, "date-from", "", "start of the window, as YYYY-MM-DD")
	fl.StringVar(&dateTo, "date-to", "", "end of the window, as YYYY-MM-DD")
	fl.StringVar(&timeUnit, "time-unit", "", "bucket size: Y, m, W, d or H")
	fl.StringVar(&dataFormat, "data-format", "", "set to raw for unformatted numbers")
	fl.IntVar(&limit, "limit", 0, "row limit, for datasets that return lists")
	fl.IntVar(&offset, "offset", 0, "row offset, for datasets that return lists")
	fl.BoolVar(&withCounts, "with-counts", false, "include totals alongside the series")
	fl.StringArrayVar(&datasets, "dataset", nil, "series to include, repeatable (e.g. sent, opened, clicked, unsubscribed)")
	fl.StringArrayVar(&filters, "filter", nil, "scope as name=value, repeatable (campaignId, companyId, segmentId)")

	group.AddCommand(types, get)
	return group
}
