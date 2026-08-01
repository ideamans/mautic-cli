---
name: mautic-usage
description: Work with a Mautic marketing automation instance from the command line — look up contacts and their activity, manage segments, campaigns, forms and emails, send email, query page-hit and email statistics, and switch between Mautic instances with named profiles. Use when someone asks about Mautic contacts, leads, segments, campaigns, landing pages, email performance, form submissions, or marketing automation data.
license: MIT
compatibility: Requires the `mautic` binary on PATH — run the mautic-install skill if it is missing — and credentials for a Mautic instance, either MAUTIC_BASE_URL plus MAUTIC_API_USER and MAUTIC_API_KEY, or a profile configured with `mautic configure`.
allowed-tools: Bash(mautic:*) Bash(jq:*) Bash(command:*) Read Write
---

# Working with Mautic

`mautic` wraps the Mautic REST API. This skill is the procedure; the complete
reference is in the binary, via `mautic llm`.

## 1. Confirm the tool and the credentials

```bash
command -v mautic && mautic auth status --verify
```

- `command not found` → run the **mautic-install** skill.
- `credentials: NOT USABLE` → the output names exactly what is missing. Ask the
  user for the Mautic instance URL, an API user and its password, then either
  export `MAUTIC_BASE_URL`, `MAUTIC_API_USER` and `MAUTIC_API_KEY`, or run
  `mautic configure --url … --username … --password …`.
- `HTTP 401` or `403` → the credentials are wrong, or Mautic's API is switched
  off (Configuration → API Settings), or the user's role lacks the permission.

**Never echo a password or token back into the conversation.** Read the value
from the environment; do not print it.

The status output names the instance and the identity. Check it before any
write — a Mautic instance usually holds real customer data, and several
instances usually look alike.

## 2. Before writing anything

Writes here reach real people and some are irreversible. In order:

| Command | What it actually does |
| --- | --- |
| `mautic emails send <id>` | mails **every contact** in the email's segments. Cannot be recalled |
| `mautic contacts batch delete --id …` | permanent |
| `mautic contacts dnc remove <id> email` | reverses a consent decision the contact made |
| `mautic <resource> update <id> --replace` | PUT: blanks every field you did not pass |
| `mautic fields delete <id>` | drops the column and every value in it |

Everything that writes accepts `--dry-run`, which prints the exact request and
sends nothing, and refuses to run without `--yes`. Use both:

```bash
mautic emails send 42 --dry-run      # read the target: line on stderr
mautic emails send 42 --yes
```

Report what a dry run would do and let the user decide. Do not pass `--yes` on
their behalf unless they asked for the action itself.

## 3. Pick the command by the question

| The user asks | Command |
| --- | --- |
| "how many contacts do we have?" | `mautic contacts list --limit 1 --format json` — read `total`, not `count` |
| "find this person" | `mautic contacts list --search 'email:someone@example.com' --format json` |
| "what has this person done?" | `mautic contacts activity <id> --limit 20` |
| "which segments is this contact in?" | `mautic contacts segments <id>` |
| "what pages got traffic yesterday?" | `mautic stats query page_hits --where date_hit:gte:<date> --order date_hit:desc --all --format csv` |
| "how did that email do?" | `mautic data get emails.in.time --date-from … --date-to … --time-unit d --dataset sent --dataset opened` |
| "list our segments / campaigns / forms" | `mautic segments list` · `mautic campaigns list` · `mautic forms list` |
| "add this person to a segment" | `mautic segments add-contact <segment-id> <contact-id>` |
| "who submitted this form?" | `mautic forms submissions <form-id>` |
| "stop emailing them" | `mautic contacts dnc add <id> email` |
| "what fields can I set?" | `mautic fields list` |
| something not listed here | `mautic api GET /some/path` |

Field aliases are configured per instance. Run `mautic fields list` before
writing contact data — Mautic ignores properties it does not recognise, so a
wrong alias looks like a successful write.

## 4. Read only the reference you need

The command catalog is long. Take a slice:

```bash
mautic llm | sed -n '1,/^# Pitfalls/p'                                          # the guide
mautic llm --format json | jq -r '.[] | select(.file=="20-analytics.md") | .body'
mautic contacts list --help
```

## 5. Run it, and interpret the output honestly

- **Parse `--format json`, `jsonl` or `csv`. Never the table** — it truncates
  cells and shows a curated subset of columns.
- **`total` is not `count`.** `--format json` returns
  `{"total":N,"count":M,"items":[…]}`; the default page is 30 records. Say "the
  first 30 of 4,318" rather than "4,318" when you only fetched 30. Use `--all`
  to page through everything.
- **Most contacts are anonymous visitors** with no email or name. Filter with
  `--where email:isNotNull` before reporting on "contacts".
- **`--order-by` and `--where` take database columns in snake_case** —
  `date_identified`, not `dateIdentified`. A wrong column is ignored silently
  and the rows come back in some other order.
- **Stats timestamps are UTC**, while entity dates use the instance's timezone.
  When a daily figure is off by one day, this is why.
- **A batch 200 is not per-record success.** Check `statusCodes` in the
  response.

## Failure modes

| Symptom | Fix |
| --- | --- |
| `command not found: mautic` | run the **mautic-install** skill |
| `credentials are incomplete` | the message lists every missing value — set `MAUTIC_BASE_URL`, `MAUTIC_API_USER`, `MAUTIC_API_KEY`, or run `mautic configure` |
| `HTTP 401` | wrong credentials; for OAuth2 the token expired — `mautic auth login` |
| `HTTP 403` | the Mautic user's role lacks the permission — check `mautic users self` |
| `HTTP 404` on every command | the API is disabled (Configuration → API Settings), or the URL is not the instance root |
| `the response was not JSON` | the URL points at the website rather than Mautic, or a login page came back |
| `HTTP 422` with rejected fields | wrong field alias or type — `mautic fields list` |
| `HTTP 429` | already retried with backoff; space the calls out or use `--all` instead of a loop |
| `... is destructive and was not performed` | intended: re-run with `--dry-run` to inspect, `--yes` to proceed |
| write succeeded but nothing changed | the property name was not one Mautic knows — verify with `get`, check `mautic fields list` |
| wrong instance was touched | `mautic auth status` names the instance; select another with `--profile <name>` |
