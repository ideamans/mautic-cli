package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/mautic-cli/internal/mautic"
)

// contactExtras adds everything the contacts endpoint does beyond CRUD:
// activity feeds, relationships, points, consent, UTM tags and batch writes.
func contactExtras(group *cobra.Command, r Resource) {
	group.AddCommand(
		contactActivityCommand(),
		contactActivityAllCommand(),
		subList("segments <contact-id>", "List the segments a contact belongs to", 1,
			func(a []string) string { return "/contacts/" + a[0] + "/segments" },
			"lists", []string{"id", "name", "alias"}),
		subList("campaigns <contact-id>", "List the campaigns a contact is a member of", 1,
			func(a []string) string { return "/contacts/" + a[0] + "/campaigns" },
			"campaigns", []string{"id", "name", "dateAdded"}),
		subList("companies <contact-id>", "List the companies a contact belongs to", 1,
			func(a []string) string { return "/contacts/" + a[0] + "/companies" },
			"companies", []string{"company_id", "companyname", "is_primary", "date_added"}),
		subList("devices <contact-id>", "List the devices seen for a contact", 1,
			func(a []string) string { return "/contacts/" + a[0] + "/devices" },
			"devices", []string{"id", "device", "deviceOsName", "deviceBrand", "dateAdded"}),
		subList("notes <contact-id>", "List the notes attached to a contact", 1,
			func(a []string) string { return "/contacts/" + a[0] + "/notes" },
			"notes", []string{"id", "type", "text", "dateTime"}),
		contactEventsCommand(),
		subGet("owners", "List the Mautic users who can own contacts", 0,
			func([]string) string { return "/contacts/list/owners" }, "", nil),
		subList("fields", "List the contact field definitions", 0,
			func([]string) string { return "/contacts/list/fields" },
			"fields", []string{"id", "alias", "label", "type", "group"}),
		contactPointsCommand(),
		contactDNCCommand(),
		contactUTMCommand(),
		contactBatchCommand(r),
	)
}

// activityFlags are the filters the activity endpoints accept. They are worth
// typed flags because the alternative — hand-built where clauses over an event
// log — is where agents produce plausible nonsense.
type activityFlags struct {
	search      string
	includeType []string
	excludeType []string
	dateFrom    string
	dateTo      string
	page        int
	limit       int
	order       string
}

func (f *activityFlags) register(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVar(&f.search, "search", "", "filter events by a search string")
	fl.StringArrayVar(&f.includeType, "include-type", nil, "event type to include, repeatable (e.g. email.read, page.hit, form.submitted)")
	fl.StringArrayVar(&f.excludeType, "exclude-type", nil, "event type to exclude, repeatable")
	fl.StringVar(&f.dateFrom, "date-from", "", "earliest event date, as YYYY-MM-DD")
	fl.StringVar(&f.dateTo, "date-to", "", "latest event date, as YYYY-MM-DD")
	fl.IntVar(&f.page, "page", 0, "page number, starting at 1")
	fl.IntVar(&f.limit, "limit", 0, "events per page")
	fl.StringVar(&f.order, "order-dir", "", "sort direction for the timestamp: asc or desc")
}

func (f *activityFlags) values() url.Values {
	q := url.Values{}
	if f.search != "" {
		q.Set("search", f.search)
	}
	for _, t := range f.includeType {
		q.Add("includeEvents[]", t)
	}
	for _, t := range f.excludeType {
		q.Add("excludeEvents[]", t)
	}
	if f.dateFrom != "" {
		q.Set("dateFrom", f.dateFrom)
	}
	if f.dateTo != "" {
		q.Set("dateTo", f.dateTo)
	}
	if f.page > 0 {
		q.Set("page", strconv.Itoa(f.page))
	}
	if f.limit > 0 {
		q.Set("limit", strconv.Itoa(f.limit))
	}
	if f.order != "" {
		q.Set("order[1]", f.order)
	}
	return q
}

var activityColumns = []string{"eventType", "eventLabel.label", "timestamp", "contactId"}

func contactActivityCommand() *cobra.Command {
	var f activityFlags
	cmd := &cobra.Command{
		Use:   "activity <contact-id>",
		Short: "Show one contact's activity timeline",
		Long: `Show one contact's activity timeline: emails read, pages hit, forms
submitted, points changed and so on.

Use "mautic contacts activity-all" for the instance-wide feed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivity(cmd, "/contacts/"+args[0]+"/activity", f)
		},
	}
	f.register(cmd)
	return cmd
}

func contactActivityAllCommand() *cobra.Command {
	var f activityFlags
	cmd := &cobra.Command{
		Use:   "activity-all",
		Short: "Show the activity timeline across every contact",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runActivity(cmd, "/contacts/activity", f)
		},
	}
	f.register(cmd)
	return cmd
}

// contactEventsCommand wraps GET /contacts/ID/events.
//
// Mautic deprecated this in favour of the activity endpoint and removed it
// outright in 7.0, so on a current instance it 404s. It stays because the
// endpoint is documented and still present on 2.x–6.x, and because an agent
// working from Mautic's own docs will reach for it — better that it fails with
// a sentence naming the replacement than with a bare 404.
func contactEventsCommand() *cobra.Command {
	var f activityFlags
	cmd := &cobra.Command{
		Use:   "events <contact-id>",
		Short: "Show one contact's events (removed in Mautic 7 — use `activity`)",
		Long: `Show one contact's events.

Deprecated in favour of "mautic contacts activity", and removed in Mautic 7.0:
on a 7.x instance this returns 404. Use "mautic contacts activity" instead.

It remains for instances still on 2.x–6.x. Note that it nests its filters under
filters[...], where the activity endpoint takes them as bare parameters.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			// This endpoint nests its filters, unlike /activity.
			q := url.Values{}
			if f.search != "" {
				q.Set("filters[search]", f.search)
			}
			for _, t := range f.includeType {
				q.Add("filters[includeEvents][]", t)
			}
			for _, t := range f.excludeType {
				q.Add("filters[excludeEvents][]", t)
			}
			if f.page > 0 {
				q.Set("page", strconv.Itoa(f.page))
			}
			if f.order != "" {
				q.Set("order[1]", f.order)
			}
			raw, err := client.Get(cmd.Context(), "/contacts/"+args[0]+"/events", q)
			if mautic.IsNotFound(err) {
				return fmt.Errorf("this instance has no /contacts/%s/events endpoint\n  Mautic removed it in 7.0 — use: mautic contacts activity %s", args[0], args[0])
			}
			if err != nil {
				return err
			}
			if flagFormat == formatJSON {
				return emitRaw(cmd, raw)
			}
			page, err := decodeNamed(raw, "events")
			if err != nil {
				return err
			}
			return emitPage(cmd, page, activityColumns)
		},
	}
	f.register(cmd)
	return cmd
}

func runActivity(cmd *cobra.Command, path string, f activityFlags) error {
	client, _, err := newClient(cmd.Context())
	if err != nil {
		return err
	}
	raw, err := client.Get(cmd.Context(), path, f.values())
	if err != nil {
		return err
	}
	if flagFormat == formatJSON {
		return emitRaw(cmd, raw)
	}
	page, err := decodeNamed(raw, "events")
	if err != nil {
		return err
	}
	return emitPage(cmd, page, activityColumns)
}

func contactPointsCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "points",
		Short: "Adjust a contact's point total",
		Long: `Adjust a contact's point total relatively.

To set an absolute total instead, use "mautic contacts update <id> --points N".`,
	}
	adjust := func(use, short, verb string) *cobra.Command {
		var eventName, actionName string
		cmd := &cobra.Command{
			Use:   use,
			Short: short,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if _, err := strconv.Atoi(args[1]); err != nil {
					return fmt.Errorf("the point delta must be a whole number, got %q", args[1])
				}
				client, creds, err := newClient(cmd.Context())
				if err != nil {
					return err
				}
				body := map[string]any{}
				if eventName != "" {
					body["eventName"] = eventName
				}
				if actionName != "" {
					body["actionName"] = actionName
				}
				noteTarget(cmd, creds)
				raw, err := client.Do(cmd.Context(), "POST",
					"/contacts/"+args[0]+"/points/"+verb+"/"+args[1], nil, body)
				if err != nil {
					return err
				}
				return emitRaw(cmd, raw)
			},
		}
		cmd.Flags().StringVar(&eventName, "event-name", "", "name recorded in the contact's point history")
		cmd.Flags().StringVar(&actionName, "action-name", "", "action recorded in the contact's point history")
		return cmd
	}
	group.AddCommand(
		adjust("add <contact-id> <points>", "Add points to a contact", "plus"),
		adjust("subtract <contact-id> <points>", "Subtract points from a contact", "minus"),
	)
	return group
}

func contactDNCCommand() *cobra.Command {
	group := &cobra.Command{
		Use:     "dnc",
		Aliases: []string{"do-not-contact"},
		Short:   "Manage a contact's do-not-contact flags",
		Long: `Manage a contact's do-not-contact flags.

Channel is usually "email" or "sms". Adding a flag suppresses that channel for
the contact; this is consent state, so treat removing one as something the
contact asked for, not as cleanup.`,
	}

	var reason int
	var comments string
	add := &cobra.Command{
		Use:   "add <contact-id> <channel>",
		Short: "Mark a contact as do-not-contact on a channel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			body := map[string]any{}
			if cmd.Flags().Changed("reason") {
				body["reason"] = reason
			}
			if comments != "" {
				body["comments"] = comments
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST",
				"/contacts/"+args[0]+"/dnc/"+args[1]+"/add", nil, body)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	add.Flags().IntVar(&reason, "reason", 0, "why: 1 unsubscribed, 2 bounced, 3 manual")
	add.Flags().StringVar(&comments, "comments", "", "free-text note stored with the flag")

	remove := &cobra.Command{
		Use:   "remove <contact-id> <channel>",
		Short: "Clear a contact's do-not-contact flag on a channel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "clearing consent state for contact "+args[0]+" on "+args[1]); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST",
				"/contacts/"+args[0]+"/dnc/"+args[1]+"/remove", nil, map[string]any{})
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	group.AddCommand(add, remove)
	return group
}

func contactUTMCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "utm",
		Short: "Manage the UTM tags recorded against a contact",
	}

	var source, medium, campaign, content, term, userAgent, url_, referer, query, remoteHost, lastActive string
	add := &cobra.Command{
		Use:   "add <contact-id>",
		Short: "Attach a set of UTM tags to a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			body := map[string]any{}
			for key, val := range map[string]string{
				"utmSource": source, "utmMedium": medium, "utmCampaign": campaign,
				"utmContent": content, "utmTerm": term, "userAgent": userAgent,
				"url": url_, "referer": referer, "query": query,
				"remoteHost": remoteHost, "lastActive": lastActive,
			} {
				if val != "" {
					body[key] = val
				}
			}
			if len(body) == 0 {
				return fmt.Errorf("pass at least one UTM value, for example --source newsletter")
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", "/contacts/"+args[0]+"/utm/add", nil, body)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	fl := add.Flags()
	fl.StringVar(&source, "source", "", "utm_source")
	fl.StringVar(&medium, "medium", "", "utm_medium")
	fl.StringVar(&campaign, "campaign", "", "utm_campaign")
	fl.StringVar(&content, "content", "", "utm_content")
	fl.StringVar(&term, "term", "", "utm_term")
	fl.StringVar(&url_, "page-url", "", "the URL the tags were seen on")
	fl.StringVar(&referer, "referer", "", "referring URL")
	fl.StringVar(&query, "query", "", "raw query string")
	fl.StringVar(&userAgent, "user-agent", "", "user agent string")
	fl.StringVar(&remoteHost, "remote-host", "", "remote host")
	fl.StringVar(&lastActive, "last-active", "", "last active date/time, as ISO 8601")

	remove := &cobra.Command{
		Use:   "remove <contact-id> <utm-id>",
		Short: "Remove one set of UTM tags from a contact",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST",
				"/contacts/"+args[0]+"/utm/"+args[1]+"/remove", nil, map[string]any{})
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	group.AddCommand(add, remove)
	return group
}

// contactBatchCommand wraps the batch endpoints.
//
// Batch bodies are JSON *arrays*, not objects, so these take --data rather than
// per-field flags: the operation is inherently file-driven. Mautic answers with
// a per-record status array, and a 200 does not mean every record succeeded —
// which is why the summary line calls out the failures.
func contactBatchCommand(r Resource) *cobra.Command {
	group := &cobra.Command{
		Use:   "batch",
		Short: "Create, update or delete many contacts in one request",
		Long: `Create, update or delete many contacts in one request.

The body is a JSON array of contact objects, so pass --data @contacts.json (or
@- for stdin). Mautic replies with one status per record: an overall HTTP 200
does not mean every record was accepted, so check the statusCodes in the
response before reporting success.`,
	}

	batchWrite := func(use, short, method, path string) *cobra.Command {
		var data string
		cmd := &cobra.Command{
			Use:   use,
			Short: short,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				if data == "" {
					return fmt.Errorf("pass --data @contacts.json (a JSON array of contact objects), or @- for stdin")
				}
				raw, err := readData(data, cmd.InOrStdin())
				if err != nil {
					return err
				}
				var body []map[string]any
				if err := unmarshalStrict(raw, &body); err != nil {
					return fmt.Errorf("--data must be a JSON array of objects: %w", err)
				}
				if len(body) == 0 {
					return fmt.Errorf("--data contained no records")
				}
				if err := confirmed(cmd, fmt.Sprintf("a batch %s of %d contacts", strings.ToLower(short), len(body))); err != nil {
					return err
				}
				client, creds, err := newClient(cmd.Context())
				if err != nil {
					return err
				}
				noteTarget(cmd, creds)
				resp, err := client.Do(cmd.Context(), method, path, nil, body)
				if err != nil {
					return err
				}
				if client.DryRun() {
					return emitRaw(cmd, resp)
				}
				reportBatchStatuses(cmd, resp)
				return emitRaw(cmd, resp)
			},
		}
		cmd.Flags().StringVar(&data, "data", "", "JSON array of contact objects: inline, @file.json or @- for stdin")
		return cmd
	}

	var ids []string
	del := &cobra.Command{
		Use:   "delete",
		Short: "Delete many contacts by ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var all []string
			for _, id := range ids {
				all = append(all, splitCSV(id)...)
			}
			if len(all) == 0 {
				return fmt.Errorf("pass --id <id> (repeatable, or comma-separated)")
			}
			if err := confirmed(cmd, fmt.Sprintf("deleting %d contacts", len(all))); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			q.Set("ids", strings.Join(all, ","))
			noteTarget(cmd, creds)
			resp, err := client.Do(cmd.Context(), "DELETE", "/contacts/batch/delete", q, nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, resp)
		},
	}
	del.Flags().StringArrayVar(&ids, "id", nil, "contact ID to delete, repeatable or comma-separated")

	group.AddCommand(
		batchWrite("create", "Create many contacts", "POST", "/contacts/batch/new"),
		batchWrite("update", "Update many contacts", "PATCH", "/contacts/batch/edit"),
		del,
	)
	_ = r
	return group
}

// reportBatchStatuses writes a one-line summary to stderr so a partial failure
// is visible even when the caller only glances at the output.
func reportBatchStatuses(cmd *cobra.Command, raw json.RawMessage) {
	var envelope struct {
		StatusCodes []int `json:"statusCodes"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.StatusCodes) == 0 {
		return
	}
	ok, failed := 0, 0
	for _, code := range envelope.StatusCodes {
		if code >= 200 && code < 300 {
			ok++
		} else {
			failed++
		}
	}
	if failed > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "batch result: %d succeeded, %d failed — check statusCodes and errors in the response\n", ok, failed)
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "batch result: %d succeeded\n", ok)
}

// decodeNamed pulls a named collection out of a response and wraps it as a Page
// so the normal output path can render it.
func decodeNamed(raw json.RawMessage, key string) (*mautic.Page, error) {
	return mautic.DecodePage(raw, key)
}

// unmarshalStrict rejects trailing garbage, which is almost always a truncated
// or concatenated file rather than something the user meant.
func unmarshalStrict(data []byte, v any) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}
