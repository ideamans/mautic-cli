package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// subList builds a read-only subcommand that lists a nested collection, such as
// the contacts in a campaign. It accepts the same query parameters as a
// top-level list so that paging and filtering work everywhere.
func subList(use, short string, argc int, path func(args []string) string, key string, cols []string) *cobra.Command {
	var f listFlags
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(argc),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			params, err := f.params()
			if err != nil {
				return err
			}
			page, err := fetchList(cmd, client, path(args), params, key, f)
			if err != nil {
				return err
			}
			return emitPage(cmd, page, cols)
		},
	}
	f.register(cmd)
	return cmd
}

// subGet builds a read-only subcommand returning a single document.
func subGet(use, short string, argc int, path func(args []string) string, key string, cols []string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(argc),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			raw, err := client.Get(cmd.Context(), path(args), nil)
			if err != nil {
				return err
			}
			return emitObject(cmd, unwrap(raw, key), cols)
		},
	}
}

// linkCommand builds an add/remove membership command.
//
// Mautic spells both as POST with an empty body. They are reversible, so they
// do not demand --yes, but they do print the target instance because attaching
// a contact to the wrong client's campaign is a real mistake.
func linkCommand(use, short string, path func(args []string) string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", path(args), nil, map[string]any{})
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
}

func campaignExtras(group *cobra.Command, r Resource) {
	group.AddCommand(
		subList("contacts <campaign-id>", "List the contacts in a campaign", 1,
			func(a []string) string { return "/campaigns/" + a[0] + "/contacts" },
			"contacts", []string{"campaign_id", "lead_id", "date_added", "manually_added"}),
		linkCommand("add-contact <campaign-id> <contact-id>", "Add a contact to a campaign",
			func(a []string) string { return "/campaigns/" + a[0] + "/contact/" + a[1] + "/add" }),
		linkCommand("remove-contact <campaign-id> <contact-id>", "Remove a contact from a campaign",
			func(a []string) string { return "/campaigns/" + a[0] + "/contact/" + a[1] + "/remove" }),
		&cobra.Command{
			Use:   "clone <campaign-id>",
			Short: "Clone a campaign, including its events",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, creds, err := newClient(cmd.Context())
				if err != nil {
					return err
				}
				noteTarget(cmd, creds)
				raw, err := client.Do(cmd.Context(), "POST", "/campaigns/clone/"+args[0], nil, map[string]any{})
				if err != nil {
					return err
				}
				if client.DryRun() {
					return emitRaw(cmd, raw)
				}
				return emitObject(cmd, unwrap(raw, r.ItemKey), r.Columns)
			},
		},
	)
}

func companyExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(
		linkCommand("add-contact <company-id> <contact-id>", "Attach a contact to a company",
			func(a []string) string { return "/companies/" + a[0] + "/contact/" + a[1] + "/add" }),
		linkCommand("remove-contact <company-id> <contact-id>", "Detach a contact from a company",
			func(a []string) string { return "/companies/" + a[0] + "/contact/" + a[1] + "/remove" }),
	)
}

func segmentExtras(group *cobra.Command, _ Resource) {
	var ids []string
	addMany := &cobra.Command{
		Use:   "add-contacts <segment-id>",
		Short: "Add many contacts to a segment in one request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(ids) == 0 {
				return fmt.Errorf("pass at least one --contact <id>")
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", "/segments/"+args[0]+"/contacts/add", nil,
				map[string]any{"ids": ids})
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	addMany.Flags().StringArrayVar(&ids, "contact", nil, "contact ID to add, repeatable")

	group.AddCommand(
		linkCommand("add-contact <segment-id> <contact-id>", "Add one contact to a segment",
			func(a []string) string { return "/segments/" + a[0] + "/contact/" + a[1] + "/add" }),
		linkCommand("remove-contact <segment-id> <contact-id>", "Remove one contact from a segment",
			func(a []string) string { return "/segments/" + a[0] + "/contact/" + a[1] + "/remove" }),
		addMany,
	)
}

func stageExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(
		linkCommand("add-contact <stage-id> <contact-id>", "Move a contact into a stage",
			func(a []string) string { return "/stages/" + a[0] + "/contact/" + a[1] + "/add" }),
		linkCommand("remove-contact <stage-id> <contact-id>", "Remove a contact from a stage",
			func(a []string) string { return "/stages/" + a[0] + "/contact/" + a[1] + "/remove" }),
	)
}

func emailExtras(group *cobra.Command, _ Resource) {
	send := &cobra.Command{
		Use:   "send <email-id>",
		Short: "Send a segment email to everyone in its segments",
		Long: `Send a segment email to every contact in the segments assigned to it.

This reaches real people and cannot be recalled. The command refuses to run
without --yes; use --dry-run first to see exactly what would be sent, and note
which instance the target line names.

Only emails with emailType "list" can be broadcast this way. A template email
has to be sent per contact with "mautic emails send-to-contact".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "sending email "+args[0]+" to its segments"); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", "/emails/"+args[0]+"/send", nil, map[string]any{})
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}

	var tokens string
	sendTo := &cobra.Command{
		Use:   "send-to-contact <email-id> <contact-id>",
		Short: "Send an email to one contact",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "sending email "+args[0]+" to contact "+args[1]); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			body := map[string]any{}
			if tokens != "" {
				parsed, err := decodeJSONObject(cmd, tokens, "--tokens")
				if err != nil {
					return err
				}
				body["tokens"] = parsed
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "POST", "/emails/"+args[0]+"/contact/"+args[1]+"/send", nil, body)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	sendTo.Flags().StringVar(&tokens, "tokens", "", "token replacements as a JSON object, inline or @file.json")

	group.AddCommand(send, sendTo)
}

func smsExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(&cobra.Command{
		Use:   "send <sms-id> <contact-id>",
		Short: "Send a text message to one contact",
		Long: `Send a text message to one contact.

Mautic exposes this as a GET, which is unusual for something that sends a
message — so --dry-run does not suppress it. Check the arguments before running.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "sending SMS "+args[0]+" to contact "+args[1]); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			noteTarget(cmd, creds)
			// Mautic really does define this endpoint as GET.
			raw, err := client.Get(cmd.Context(), "/smses/"+args[0]+"/contact/"+args[1]+"/send", nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	})
}

func formExtras(group *cobra.Command, _ Resource) {
	submissionCols := []string{"id", "dateSubmitted", "lead.id", "ipAddress.ipAddress", "referer"}
	group.AddCommand(
		subList("submissions <form-id>", "List submissions for a form", 1,
			func(a []string) string { return "/forms/" + a[0] + "/submissions" },
			"submissions", submissionCols),
		subGet("submission <form-id> <submission-id>", "Get one form submission", 2,
			func(a []string) string { return "/forms/" + a[0] + "/submissions/" + a[1] },
			"submission", nil),
		subList("contact-submissions <form-id> <contact-id>", "List one contact's submissions of a form", 2,
			func(a []string) string { return "/forms/" + a[0] + "/submissions/contact/" + a[1] },
			"submissions", submissionCols),
		deleteChildren("delete-fields <form-id>", "Delete fields from a form", "field",
			func(id string) string { return "/forms/" + id + "/fields/delete" }, "fields"),
		deleteChildren("delete-actions <form-id>", "Delete actions from a form", "action",
			func(id string) string { return "/forms/" + id + "/actions/delete" }, "actions"),
	)
}

// deleteChildren builds the "delete some children of this record" commands,
// which Mautic expresses as a DELETE with a repeated query parameter.
func deleteChildren(use, short, itemFlag string, path func(id string) string, param string) *cobra.Command {
	var ids []string
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(ids) == 0 {
				return fmt.Errorf("pass at least one --%s <id>", itemFlag)
			}
			if err := confirmed(cmd, fmt.Sprintf("deleting %d %s(s) from %s", len(ids), itemFlag, args[0])); err != nil {
				return err
			}
			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			for _, id := range ids {
				q.Add(param+"[]", id)
			}
			noteTarget(cmd, creds)
			raw, err := client.Do(cmd.Context(), "DELETE", path(args[0]), q, nil)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	cmd.Flags().StringArrayVar(&ids, itemFlag, nil, "ID to delete, repeatable")
	return cmd
}

func pointExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(subGet("types", "List the available point action types", 0,
		func([]string) string { return "/points/actions/types" }, "pointActionTypes", nil))
}

func pointTriggerExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(
		subGet("event-types", "List the available point trigger event types", 0,
			func([]string) string { return "/points/triggers/events/types" }, "eventTypes", nil),
		deleteChildren("delete-events <trigger-id>", "Delete events from a point trigger", "event",
			func(id string) string { return "/points/triggers/" + id + "/events/delete" }, "events"),
	)
}

func userExtras(group *cobra.Command, r Resource) {
	var permission string
	check := &cobra.Command{
		Use:   "permission-check <user-id>",
		Short: "Check whether a user holds a permission",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if permission == "" {
				return fmt.Errorf("pass --permission <bundle:area:level>, for example lead:leads:viewown")
			}
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			q.Set("permissions[]", permission)
			raw, err := client.Get(cmd.Context(), "/users/"+args[0]+"/permissioncheck", q)
			if err != nil {
				return err
			}
			return emitRaw(cmd, raw)
		},
	}
	check.Flags().StringVar(&permission, "permission", "", "permission to test, as bundle:area:level")

	group.AddCommand(
		subGet("self", "Show the user the current credentials authenticate as", 0,
			func([]string) string { return "/users/self" }, "user", r.Columns),
		check,
	)
}

func webhookExtras(group *cobra.Command, _ Resource) {
	group.AddCommand(subGet("triggers", "List the events a webhook can subscribe to", 0,
		func([]string) string { return "/hooks/triggers" }, "triggers", nil))
}

func reportExtras(group *cobra.Command, _ Resource) {
	var dateFrom, dateTo string
	var limit, page int
	results := &cobra.Command{
		Use:   "results <report-id>",
		Short: "Run a report and return its rows",
		Long: `Run a report and return its rows.

The rows come back under "data" with their own column set, which depends on the
report definition rather than on any Mautic entity — so --fields is usually
needed to get a readable table.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			q := url.Values{}
			if dateFrom != "" {
				q.Set("dateFrom", dateFrom)
			}
			if dateTo != "" {
				q.Set("dateTo", dateTo)
			}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if page > 0 {
				q.Set("page", fmt.Sprint(page))
			}
			raw, err := client.Get(cmd.Context(), "/reports/"+args[0], q)
			if err != nil {
				return err
			}
			if flagFormat == formatJSON {
				return emitRaw(cmd, raw)
			}
			p, err := decodeNamed(raw, "data")
			if err != nil {
				return err
			}
			return emitPage(cmd, p, nil)
		},
	}
	fl := results.Flags()
	fl.StringVar(&dateFrom, "date-from", "", "start of the reporting window, as YYYY-MM-DD")
	fl.StringVar(&dateTo, "date-to", "", "end of the reporting window, as YYYY-MM-DD")
	fl.IntVar(&limit, "limit", 0, "rows per page")
	fl.IntVar(&page, "page", 0, "page number, starting at 1")

	group.AddCommand(results)
}

// decodeJSONObject reads a JSON object from an inline value, @file or @-.
func decodeJSONObject(cmd *cobra.Command, spec, flag string) (map[string]any, error) {
	raw, err := readData(spec, cmd.InOrStdin())
	if err != nil {
		return nil, err
	}
	var v map[string]any
	if err := unmarshalStrict(raw, &v); err != nil {
		return nil, fmt.Errorf("%s is not a JSON object: %w", flag, err)
	}
	return v, nil
}

// splitCSV trims and drops empties, so --ids "1, 2 ,3" behaves.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
