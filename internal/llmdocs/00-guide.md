# mautic — guide

`mautic` wraps the Mautic REST API (`https://<instance>/api`). It is
non-interactive: nothing prompts, destructive commands refuse to run without
`--yes`, and everything that writes accepts `--dry-run`. **stdout carries the
result and nothing else** — progress, warnings, the target line and record
counts all go to stderr. Exit code is 0 on success and non-zero on failure, with
errors prefixed `Error:`.

This reference is embedded in the binary you are running, so it describes
exactly this build and beats anything you remember about Mautic's API.

## The six rules

1. **Never hand-assemble a request body.** Every writable property has a flag:
   `mautic contacts create --email a@example.com --firstname A`. For fields this
   instance added, use `--set alias=value`. Reach for `--data @file.json` only
   for genuinely nested structures (campaign events, form fields, segment
   filters).
2. **Field aliases are per-instance. Look them up.** `mautic fields list` prints
   what this Mautic actually has. Guessing an alias gives you a 422, or worse, a
   silently ignored property.
3. **Parse `--format json`, `jsonl` or `csv` — never the table.** The table is
   for humans: it truncates cells at 48 display columns and shows a curated
   subset of properties. `--fields id,fields.all.email` picks columns by dotted
   path.
4. **`total` is not `count`.** A list returns the first 30 records by default.
   `--format json` gives `{"total":4318,"count":30,"items":[…]}` — compare the
   two before concluding anything about "how many". Use `--limit`, or `--all` to
   page through everything.
5. **`--order-by` takes database column names, not response property names.**
   `dateIdentified` is `date_identified`, `modifiedByUser` is
   `modified_by_user`. A wrong value is ignored silently rather than rejected,
   so the rows come back in some other order and look plausible.
6. **Dry-run anything that reaches people.** `emails send` mails every contact in
   the segments assigned to it, `contacts dnc remove` undoes a consent decision,
   and `contacts batch delete` is not reversible. Run with `--dry-run` first and
   read the `target:` line on stderr — it names the instance and identity.

## Authentication

Mautic accepts **basic auth** (a Mautic user's username and password, enabled
under Configuration → API Settings) and **OAuth2**. Basic auth is the simpler
choice for scripts and agents; OAuth2 exists for delegated access.

Values resolve **flag > environment variable > config file**, and a complete set
in the environment beats every profile — that is what makes one code path serve
both a laptop with five profiles and a CI runner with none.

| Variable | Meaning | Aliases accepted |
| --- | --- | --- |
| `MAUTIC_BASE_URL` | instance URL, with or without a trailing `/api` | `MAUTIC_API_ENDPOINT`, `MAUTIC_URL` |
| `MAUTIC_API_USER` | username for basic auth | `MAUTIC_USERNAME`, `MAUTIC_USER` |
| `MAUTIC_API_KEY` | password for basic auth | `MAUTIC_PASSWORD` |
| `MAUTIC_CLIENT_ID` / `MAUTIC_CLIENT_SECRET` | OAuth2 client | — |
| `MAUTIC_ACCESS_TOKEN` | OAuth2 token, used as-is | — |
| `MAUTIC_PROFILE` | which profile to use | — |

Configuration lives in `~/.config/mautic-cli/` (honouring `XDG_CONFIG_HOME`):
`config.toml` holds profiles and is not secret, `credentials.json` holds
passwords and tokens at mode 0600.

```bash
mautic auth status --verify          # what would be used, and does it work
mautic configure --url https://mautic.example.com --username api-user --password s3cret
mautic profiles list                 # switch instances with --profile or `profiles use`
mautic auth login --client-id … --client-secret …   # OAuth2, opens a browser
mautic auth login --client-credentials …            # OAuth2 with no browser, for CI
```

`mautic auth status` works with nothing configured and names what is missing.
Run it first, every time. Never echo a password or token back into a
conversation.

## Which command answers which question

| The question | Command |
| --- | --- |
| who is in the database, and how many? | `mautic contacts list --limit 5 --format json` |
| find a contact by email | `mautic contacts list --search 'email:someone@example.com'` |
| what has this contact done? | `mautic contacts activity <id>` |
| which segments/campaigns is a contact in? | `mautic contacts segments <id>` / `contacts campaigns <id>` |
| what fields does this instance have? | `mautic fields list` |
| what pages were visited, and when? | `mautic stats query page_hits --where date_hit:gte:2026-08-01 --order date_hit:desc` |
| how did the last email campaign do? | `mautic data get emails.in.time --date-from … --date-to … --dataset sent --dataset opened` |
| what segments exist? | `mautic segments list` |
| add a contact to a segment | `mautic segments add-contact <segment-id> <contact-id>` |
| send an email to one person | `mautic emails send-to-contact <email-id> <contact-id>` |
| stop emailing someone | `mautic contacts dnc add <id> email` |
| what can a webhook subscribe to? | `mautic webhooks triggers` |
| anything not wrapped here | `mautic api GET /some/endpoint` |

## Reading this reference in pieces

The command catalog is long. Take a slice rather than the whole thing:

```bash
mautic llm | sed -n '1,/^# Pitfalls/p'                                        # the guide only
mautic llm --format json | jq -r '.[] | select(.file=="20-analytics.md") | .body'
mautic contacts list --help                                                   # one command
```

## Typical workflows

**Find someone and see what they did.**

```bash
mautic contacts list --search 'email:someone@example.com' --format json \
  | jq '.items[] | {id, email: .fields.all.email}'
mautic contacts activity 1234 --limit 20
```

**Count traffic without pulling every contact.** The stats tables are far
cheaper than paging the entity endpoints, and they are what the dashboard uses.

```bash
mautic stats query page_hits \
  --where date_hit:gte:2026-08-01 --order date_hit:desc --all --format csv > hits.csv
```

**Create a contact, then put it in a segment.**

```bash
id=$(mautic contacts create --email a@example.com --firstname A --format json | jq -r .id)
mautic segments add-contact 7 "$id"
```

**Bulk-load contacts from a file.** Batch bodies are JSON arrays, and a 200 does
not mean every record was accepted — check `statusCodes` in the response.

```bash
mautic contacts batch create --data @contacts.json --dry-run
mautic contacts batch create --data @contacts.json --yes
```

## Failure modes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `HTTP 401` | wrong credentials, or an expired OAuth2 token | `mautic auth status --verify`; for OAuth2, `mautic auth login` |
| `HTTP 403` | the Mautic user's role lacks the permission | check `mautic users self`; grant it under Settings → Roles |
| `HTTP 404` on everything | the API is disabled, or the URL is wrong | enable it under Configuration → API Settings; the URL is the instance root |
| `the response was not JSON` | the URL points at the site, not the API, or a login page came back | check `mautic auth status`; do not put `/api` in the path passed to `mautic api` |
| `HTTP 422` with `rejected fields` | wrong field alias or type | `mautic fields list` |
| `HTTP 429` | Mautic's API rate limiter | already retried with backoff — space the calls out |
| a write "succeeds" but nothing changed | the property name was not one Mautic knows | Mautic ignores unknown properties silently; verify with `get` |
| sort order looks wrong | `--order-by` was given a camelCase name | use the snake_case column: `date_identified`, not `dateIdentified` |
| empty email/name columns on contacts | anonymous visitors are contacts too | filter with `--where email:isNotNull` |

## What this CLI will not do

- **Build campaign logic, form fields or segment filters for you.** Those are
  nested structures the UI generates. Build one in Mautic, read it back with
  `--format json`, and use that as the shape for `--data`.
- **Send mail through anything but Mautic's own queue.** `emails send` hands the
  job to Mautic; whether it goes out immediately depends on that instance's
  queue settings.
- **Talk to `/api/v2`.** Mautic 7 ships a second, self-documenting API at
  `/api/v2`. This CLI targets the stable `/api` surface that every version from
  2.x onward exposes.
- **Configure the instance.** Plugins, mail transport, rate limiter and cron are
  all outside the REST API.
