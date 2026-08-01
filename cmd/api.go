package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// newAPICommand is the escape hatch.
//
// It exists so that an unwrapped endpoint is never a dead end: it reuses the
// resolved credentials, the base URL, the retry policy and --dry-run, and
// prints the response exactly as Mautic sent it. Anything this CLI has not
// wrapped is still one command away.
func newAPICommand() *cobra.Command {
	var (
		query []string
		data  string
	)
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Call any Mautic API endpoint directly",
		Long: `Call any Mautic API endpoint directly.

PATH is relative to the instance's /api root and must start with a slash, so
/contacts becomes https://<instance>/api/contacts. The response is printed
exactly as Mautic returned it, with no unwrapping — which is also the way to see
the envelope keys the wrapped commands hide.

Use this for endpoints this CLI does not wrap, for plugin endpoints, and for
checking what an endpoint really returns before trusting a summary of it.`,
		Example: `  mautic api GET /contacts --query limit=5 --query 'search=email:*@example.com'
  mautic api GET /stats/page_hits --query 'where[0][col]=date_hit' --query 'where[0][expr]=gte' --query 'where[0][val]=2026-08-01'
  mautic api POST /contacts/new --data '{"email":"a@example.com","firstname":"A"}'
  mautic api PATCH /contacts/42/edit --data @patch.json --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			switch method {
			case "GET", "POST", "PATCH", "PUT", "DELETE", "HEAD":
			default:
				return fmt.Errorf("unsupported method %q; use GET, POST, PATCH, PUT or DELETE", args[0])
			}
			path := args[1]
			if !strings.HasPrefix(path, "/") {
				return fmt.Errorf("PATH must start with a slash, for example /contacts (got %q)", path)
			}
			if strings.HasPrefix(path, "/api/") {
				return fmt.Errorf("PATH is relative to the API root, so drop the leading /api: use %q", strings.TrimPrefix(path, "/api"))
			}

			q := url.Values{}
			for _, item := range query {
				name, value, ok := strings.Cut(item, "=")
				if !ok {
					return fmt.Errorf("--query %q must be written as name=value", item)
				}
				q.Add(name, value)
			}

			var body any
			if data != "" {
				raw, err := readData(data, cmd.InOrStdin())
				if err != nil {
					return err
				}
				var parsed any
				if err := unmarshalStrict(raw, &parsed); err != nil {
					return fmt.Errorf("--data is not valid JSON: %w", err)
				}
				body = parsed
			}

			if method != "GET" {
				if err := confirmed(cmd, method+" "+path); err != nil {
					return err
				}
			}

			client, creds, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			if method != "GET" {
				noteTarget(cmd, creds)
			}
			resp, err := client.Do(cmd.Context(), method, path, q, body)
			if err != nil {
				return err
			}
			return emitRaw(cmd, resp)
		},
	}
	cmd.Flags().StringArrayVar(&query, "query", nil, "query parameter as name=value, repeatable")
	cmd.Flags().StringVar(&data, "data", "", "request body as JSON: inline, @file.json, or @- for stdin")
	return cmd
}
