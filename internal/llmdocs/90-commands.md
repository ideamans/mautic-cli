# Command catalog

Generated from the cobra command tree by `go generate ./...`. Do not edit by hand — edit the command definitions instead.

Global flags apply to every command and are listed once, under `mautic`.

## Global flags

| flag | type | default | description |
| --- | --- | --- | --- |
| `--access-token` | string | — | OAuth2 access token to use as-is (env: MAUTIC_ACCESS_TOKEN) |
| `--auth` | string | — | authentication method: basic or oauth2 (default: whatever the profile or environment provides) |
| `--client-id` | string | — | OAuth2 client ID (env: MAUTIC_CLIENT_ID) |
| `--client-secret` | string | — | OAuth2 client secret (env: MAUTIC_CLIENT_SECRET) |
| `--dry-run` | bool | `false` | print the request that would be sent and send nothing |
| `--fields` | string | — | comma-separated dotted paths selecting columns, e.g. id,fields.all.email |
| `--format` | string | `table` | output format: table, json, jsonl or csv |
| `--password` | string | — | Mautic password for basic auth (env: MAUTIC_API_KEY) |
| `--profile` | string | — | named profile to use (env: MAUTIC_PROFILE) |
| `--retries` | int | `3` | extra attempts for HTTP 429 and 5xx responses (0 = none; POST and PATCH are retried only on 429) |
| `--timeout` | duration | `1m0s` | timeout for a single HTTP request |
| `--url` | string | — | Mautic instance URL, with or without a trailing /api (env: MAUTIC_BASE_URL, MAUTIC_API_ENDPOINT) |
| `--username` | string | — | Mautic username for basic auth (env: MAUTIC_API_USER) |
| `--verbose` | bool | `false` | log each request to stderr |
| `--yes` | bool | `false` | skip the confirmation gate on destructive commands |

## `mautic api`

Call any Mautic API endpoint directly

Call any Mautic API endpoint directly.

PATH is relative to the instance's /api root and must start with a slash, so
/contacts becomes https://<instance>/api/contacts. The response is printed
exactly as Mautic returned it, with no unwrapping — which is also the way to see
the envelope keys the wrapped commands hide.

Use this for endpoints this CLI does not wrap, for plugin endpoints, and for
checking what an endpoint really returns before trusting a summary of it.

```
mautic api <METHOD> <PATH>
```

Example:

```
mautic api GET /contacts --query limit=5 --query 'search=email:*@example.com'
  mautic api GET /stats/page_hits --query 'where[0][col]=date_hit' --query 'where[0][expr]=gte' --query 'where[0][val]=2026-08-01'
  mautic api POST /contacts/new --data '{"email":"a@example.com","firstname":"A"}'
  mautic api PATCH /contacts/42/edit --data @patch.json --dry-run
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin |
| `--query` | stringArray | `[]` | query parameter as name=value, repeatable |

## `mautic assets`

Downloadable files tracked by Mautic

Downloadable files tracked by Mautic

Wraps the Mautic /assets endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic assets create`

Create a asset

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | URL alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | asset description |
| `--file` | string | — | file name in Mautic's asset directory (upload it first with `mautic files upload`) |
| `--language` | string | — | language locale, e.g. en |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--remote-path` | string | — | path on the remote storage when storage-location is remote |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--storage-location` | string | — | where the file lives: local or remote |
| `--title` | string | — | asset title |

### `mautic assets delete`

Delete a asset

Delete a asset.

This cannot be undone. The command refuses to run without --yes.

```
mautic assets delete <id>
```

### `mautic assets get`

Get one record by ID

```
mautic assets get <id>
```

### `mautic assets list`

List assets

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic assets update`

Update a asset

Update a asset.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic assets update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | URL alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | asset description |
| `--file` | string | — | file name in Mautic's asset directory (upload it first with `mautic files upload`) |
| `--language` | string | — | language locale, e.g. en |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--remote-path` | string | — | path on the remote storage when storage-location is remote |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--storage-location` | string | — | where the file lives: local or remote |
| `--title` | string | — | asset title |

## `mautic auth`

Inspect and obtain credentials

### `mautic auth login`

Obtain OAuth2 tokens and store them in a profile

Obtain OAuth2 tokens and store them in a profile.

Two grants are supported, because Mautic supports two:

  authorization_code (default)  opens a browser, needs a person, and returns a
                                refresh token valid for 14 days by default
  --client-credentials          no browser and no person, so it is the one that
                                works from CI — but not every instance enables it

The redirect URI must be registered in Mautic under Settings -> API Credentials
exactly as it is passed here, including the port and path.

Basic auth needs none of this: "mautic configure --username <user> --password
<key>" is enough, and is the simpler choice for scripts.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--client-credentials` | bool | `false` | use the client_credentials grant: no browser, no user, works from CI |
| `--login-timeout` | duration | `5m0s` | how long to wait for the browser callback |
| `--no-browser` | bool | `false` | print the authorize URL instead of opening a browser |
| `--redirect-uri` | string | — | callback URL, registered verbatim in Mautic (default: http://localhost:8080/oauth/callback) |
| `--save-profile` | string | — | profile to store the tokens in (default: the selected profile, or "default") |

### `mautic auth logout`

Forget the stored secrets for a profile

Forget the stored secrets for a profile.

This clears the entry in credentials.json. The profile itself stays in
config.toml, so "mautic auth login" or "mautic configure" can refill it.

### `mautic auth status`

Show which credentials would be used, and whether they work

Show which credentials would be used, and whether they work.

Run this before anything that writes. It never prints a secret; it prints where
each value came from, so that "which instance am I about to change" has an
answer.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--verify` | bool | `false` | call GET /users/self to prove the credentials actually work |

## `mautic campaigns`

Campaigns and their contact membership

Campaigns and their contact membership

Wraps the Mautic /campaigns endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic campaigns add-contact`

Add a contact to a campaign

```
mautic campaigns add-contact <campaign-id> <contact-id>
```

### `mautic campaigns clone`

Clone a campaign, including its events

```
mautic campaigns clone <campaign-id>
```

### `mautic campaigns contacts`

List the contacts in a campaign

```
mautic campaigns contacts <campaign-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic campaigns create`

Create a campaign

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | campaign alias |
| `--canvas-settings` | string | — | canvas layout as JSON |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | campaign description |
| `--events` | string | — | campaign events as a JSON array — build these in the UI and copy them out, they are not practical to write by hand |
| `--name` | string | — | campaign name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

### `mautic campaigns delete`

Delete a campaign

Delete a campaign.

This cannot be undone. The command refuses to run without --yes.

```
mautic campaigns delete <id>
```

### `mautic campaigns get`

Get one record by ID

```
mautic campaigns get <id>
```

### `mautic campaigns list`

List campaigns

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic campaigns remove-contact`

Remove a contact from a campaign

```
mautic campaigns remove-contact <campaign-id> <contact-id>
```

### `mautic campaigns update`

Update a campaign

Update a campaign.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic campaigns update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | campaign alias |
| `--canvas-settings` | string | — | canvas layout as JSON |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | campaign description |
| `--events` | string | — | campaign events as a JSON array — build these in the UI and copy them out, they are not practical to write by hand |
| `--name` | string | — | campaign name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

## `mautic categories`

Categories, shared across every bundle

Categories, shared across every bundle

Wraps the Mautic /categories endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic categories create`

Create a category

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | category alias |
| `--bundle` | string | — | which bundle the category belongs to: asset, campaign, category, email, form, page, plugin:<name>, point, sms, stage, global |
| `--color` | string | — | hex colour without the leading # |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | category description |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--title` | string | — | category title |

### `mautic categories delete`

Delete a category

Delete a category.

This cannot be undone. The command refuses to run without --yes.

```
mautic categories delete <id>
```

### `mautic categories get`

Get one record by ID

```
mautic categories get <id>
```

### `mautic categories list`

List categories

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic categories update`

Update a category

Update a category.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic categories update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | category alias |
| `--bundle` | string | — | which bundle the category belongs to: asset, campaign, category, email, form, page, plugin:<name>, point, sms, stage, global |
| `--color` | string | — | hex colour without the leading # |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | category description |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--title` | string | — | category title |

## `mautic companies`

Companies and their contact membership

Companies and their contact membership

Wraps the Mautic /companies endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic companies add-contact`

Attach a contact to a company

```
mautic companies add-contact <company-id> <contact-id>
```

### `mautic companies create`

Create a company

| flag | type | default | description |
| --- | --- | --- | --- |
| `--address1` | string | — | address line 1 |
| `--address2` | string | — | address line 2 |
| `--city` | string | — | city |
| `--country` | string | — | country |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--email` | string | — | company email address |
| `--industry` | string | — | industry |
| `--name` | string | — | company name (required) |
| `--owner` | int | `0` | ID of the Mautic user who owns the company |
| `--phone` | string | — | company phone number |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--state` | string | — | state or prefecture |
| `--website` | string | — | company website |
| `--zipcode` | string | — | postal code |

### `mautic companies delete`

Delete a company

Delete a company.

This cannot be undone. The command refuses to run without --yes.

```
mautic companies delete <id>
```

### `mautic companies get`

Get one record by ID

```
mautic companies get <id>
```

### `mautic companies list`

List companies

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic companies remove-contact`

Detach a contact from a company

```
mautic companies remove-contact <company-id> <contact-id>
```

### `mautic companies update`

Update a company

Update a company.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic companies update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--address1` | string | — | address line 1 |
| `--address2` | string | — | address line 2 |
| `--city` | string | — | city |
| `--country` | string | — | country |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--email` | string | — | company email address |
| `--industry` | string | — | industry |
| `--name` | string | — | company name (required) |
| `--owner` | int | `0` | ID of the Mautic user who owns the company |
| `--phone` | string | — | company phone number |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--state` | string | — | state or prefecture |
| `--website` | string | — | company website |
| `--zipcode` | string | — | postal code |

## `mautic configure`

Add or update a named profile

Add or update a named profile.

A profile is one Mautic instance plus the credentials for it. The non-secret
half goes to config.toml and the secrets to credentials.json, both under
~/.config/mautic-cli/ at 0600.

Basic auth is the quickest way to a working setup:

  mautic configure --url https://mautic.example.com \
    --username api-user --password s3cret

For OAuth2, register the client here and then run "mautic auth login":

  mautic configure --url https://mautic.example.com --auth oauth2 \
    --client-id 1_abc --client-secret def
  mautic auth login

Nothing here is required to use the CLI: setting MAUTIC_BASE_URL,
MAUTIC_API_USER and MAUTIC_API_KEY works with no config file at all, and the
environment beats every profile.

| flag | type | default | description |
| --- | --- | --- | --- |
| `--name` | string | — | profile name to write (default: --profile, then the current default, then "default") |

## `mautic contacts`

Contacts, their activity, points, segments and consent

Contacts, their activity, points, segments and consent.

Contact fields are configurable per instance. The flags below cover the core
fields that ship with Mautic; anything else is set with --set alias=value, and
"mautic fields list" prints the aliases this instance actually has.

Reading a contact returns fields grouped by category, with a flattened copy
under fields.all — so --fields id,fields.all.email,points is usually what you
want for a table.

Aliases: leads

### `mautic contacts activity`

Show one contact's activity timeline

Show one contact's activity timeline: emails read, pages hit, forms
submitted, points changed and so on.

Use "mautic contacts activity-all" for the instance-wide feed.

```
mautic contacts activity <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date-from` | string | — | earliest event date, as YYYY-MM-DD |
| `--date-to` | string | — | latest event date, as YYYY-MM-DD |
| `--exclude-type` | stringArray | `[]` | event type to exclude, repeatable |
| `--include-type` | stringArray | `[]` | event type to include, repeatable (e.g. email.read, page.hit, form.submitted) |
| `--limit` | int | `0` | events per page |
| `--order-dir` | string | — | sort direction for the timestamp: asc or desc |
| `--page` | int | `0` | page number, starting at 1 |
| `--search` | string | — | filter events by a search string |

### `mautic contacts activity-all`

Show the activity timeline across every contact

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date-from` | string | — | earliest event date, as YYYY-MM-DD |
| `--date-to` | string | — | latest event date, as YYYY-MM-DD |
| `--exclude-type` | stringArray | `[]` | event type to exclude, repeatable |
| `--include-type` | stringArray | `[]` | event type to include, repeatable (e.g. email.read, page.hit, form.submitted) |
| `--limit` | int | `0` | events per page |
| `--order-dir` | string | — | sort direction for the timestamp: asc or desc |
| `--page` | int | `0` | page number, starting at 1 |
| `--search` | string | — | filter events by a search string |

### `mautic contacts batch`

Create, update or delete many contacts in one request

Create, update or delete many contacts in one request.

The body is a JSON array of contact objects, so pass --data @contacts.json (or
@- for stdin). Mautic replies with one status per record: an overall HTTP 200
does not mean every record was accepted, so check the statusCodes in the
response before reporting success.

#### `mautic contacts batch create`

Create many contacts

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | JSON array of contact objects: inline, @file.json or @- for stdin |

#### `mautic contacts batch delete`

Delete many contacts by ID

| flag | type | default | description |
| --- | --- | --- | --- |
| `--id` | stringArray | `[]` | contact ID to delete, repeatable or comma-separated |

#### `mautic contacts batch update`

Update many contacts

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | JSON array of contact objects: inline, @file.json or @- for stdin |

### `mautic contacts campaigns`

List the campaigns a contact is a member of

```
mautic contacts campaigns <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts companies`

List the companies a contact belongs to

```
mautic contacts companies <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts create`

Create a contact

| flag | type | default | description |
| --- | --- | --- | --- |
| `--address1` | string | — | address line 1 |
| `--address2` | string | — | address line 2 |
| `--city` | string | — | city |
| `--company` | string | — | company name |
| `--country` | string | — | country |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email` | string | — | email address |
| `--firstname` | string | — | first name |
| `--ip` | string | — | IP address to attribute the contact to |
| `--last-active` | string | — | last active date/time |
| `--lastname` | string | — | last name |
| `--mobile` | string | — | mobile number |
| `--owner` | int | `0` | ID of the Mautic user who owns the contact |
| `--phone` | string | — | phone number |
| `--points` | int | `0` | absolute point total (use `contacts points add` to adjust it relatively) |
| `--position` | string | — | job title |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--state` | string | — | state or prefecture |
| `--tags` | stringArray | `[]` | tag to attach; prefix with - to remove it (repeatable) |
| `--website` | string | — | website |
| `--zipcode` | string | — | postal code |

### `mautic contacts delete`

Delete a contact

Delete a contact.

This cannot be undone. The command refuses to run without --yes.

```
mautic contacts delete <id>
```

### `mautic contacts devices`

List the devices seen for a contact

```
mautic contacts devices <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts dnc`

Manage a contact's do-not-contact flags

Manage a contact's do-not-contact flags.

Channel is usually "email" or "sms". Adding a flag suppresses that channel for
the contact; this is consent state, so treat removing one as something the
contact asked for, not as cleanup.

Aliases: do-not-contact

#### `mautic contacts dnc add`

Mark a contact as do-not-contact on a channel

```
mautic contacts dnc add <contact-id> <channel>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--comments` | string | — | free-text note stored with the flag |
| `--reason` | int | `0` | why: 1 unsubscribed, 2 bounced, 3 manual |

#### `mautic contacts dnc remove`

Clear a contact's do-not-contact flag on a channel

```
mautic contacts dnc remove <contact-id> <channel>
```

### `mautic contacts events`

Show one contact's events (removed in Mautic 7 — use `activity`)

Show one contact's events.

Deprecated in favour of "mautic contacts activity", and removed in Mautic 7.0:
on a 7.x instance this returns 404. Use "mautic contacts activity" instead.

It remains for instances still on 2.x–6.x. Note that it nests its filters under
filters[...], where the activity endpoint takes them as bare parameters.

```
mautic contacts events <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date-from` | string | — | earliest event date, as YYYY-MM-DD |
| `--date-to` | string | — | latest event date, as YYYY-MM-DD |
| `--exclude-type` | stringArray | `[]` | event type to exclude, repeatable |
| `--include-type` | stringArray | `[]` | event type to include, repeatable (e.g. email.read, page.hit, form.submitted) |
| `--limit` | int | `0` | events per page |
| `--order-dir` | string | — | sort direction for the timestamp: asc or desc |
| `--page` | int | `0` | page number, starting at 1 |
| `--search` | string | — | filter events by a search string |

### `mautic contacts fields`

List the contact field definitions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts get`

Get one record by ID

```
mautic contacts get <id>
```

### `mautic contacts list`

List contacts

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts notes`

List the notes attached to a contact

```
mautic contacts notes <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts owners`

List the Mautic users who can own contacts

### `mautic contacts points`

Adjust a contact's point total

Adjust a contact's point total relatively.

To set an absolute total instead, use "mautic contacts update <id> --points N".

#### `mautic contacts points add`

Add points to a contact

```
mautic contacts points add <contact-id> <points>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action-name` | string | — | action recorded in the contact's point history |
| `--event-name` | string | — | name recorded in the contact's point history |

#### `mautic contacts points subtract`

Subtract points from a contact

```
mautic contacts points subtract <contact-id> <points>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action-name` | string | — | action recorded in the contact's point history |
| `--event-name` | string | — | name recorded in the contact's point history |

### `mautic contacts segments`

List the segments a contact belongs to

```
mautic contacts segments <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic contacts update`

Update a contact

Update a contact.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic contacts update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--address1` | string | — | address line 1 |
| `--address2` | string | — | address line 2 |
| `--city` | string | — | city |
| `--company` | string | — | company name |
| `--country` | string | — | country |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email` | string | — | email address |
| `--firstname` | string | — | first name |
| `--ip` | string | — | IP address to attribute the contact to |
| `--last-active` | string | — | last active date/time |
| `--lastname` | string | — | last name |
| `--mobile` | string | — | mobile number |
| `--owner` | int | `0` | ID of the Mautic user who owns the contact |
| `--phone` | string | — | phone number |
| `--points` | int | `0` | absolute point total (use `contacts points add` to adjust it relatively) |
| `--position` | string | — | job title |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--state` | string | — | state or prefecture |
| `--tags` | stringArray | `[]` | tag to attach; prefix with - to remove it (repeatable) |
| `--website` | string | — | website |
| `--zipcode` | string | — | postal code |

### `mautic contacts utm`

Manage the UTM tags recorded against a contact

#### `mautic contacts utm add`

Attach a set of UTM tags to a contact

```
mautic contacts utm add <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--campaign` | string | — | utm_campaign |
| `--content` | string | — | utm_content |
| `--last-active` | string | — | last active date/time, as ISO 8601 |
| `--medium` | string | — | utm_medium |
| `--page-url` | string | — | the URL the tags were seen on |
| `--query` | string | — | raw query string |
| `--referer` | string | — | referring URL |
| `--remote-host` | string | — | remote host |
| `--source` | string | — | utm_source |
| `--term` | string | — | utm_term |
| `--user-agent` | string | — | user agent string |

#### `mautic contacts utm remove`

Remove one set of UTM tags from a contact

```
mautic contacts utm remove <contact-id> <utm-id>
```

## `mautic data`

Pre-built chart datasets, as used by Mautic's dashboard widgets

Pre-built chart datasets, as used by Mautic's dashboard widgets.

These return aggregates that would otherwise mean paging through the stats
tables and summing by hand — prefer them when the question is "how many, over
time".

### `mautic data get`

Fetch one dataset

```
mautic data get <type>
```

Example:

```
mautic data get emails.in.time --date-from 2026-07-01 --date-to 2026-07-31 --time-unit d \
    --dataset sent --dataset opened --with-counts
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data-format` | string | — | set to raw for unformatted numbers |
| `--dataset` | stringArray | `[]` | series to include, repeatable (e.g. sent, opened, clicked, unsubscribed) |
| `--date-from` | string | — | start of the window, as YYYY-MM-DD |
| `--date-to` | string | — | end of the window, as YYYY-MM-DD |
| `--filter` | stringArray | `[]` | scope as name=value, repeatable (campaignId, companyId, segmentId) |
| `--limit` | int | `0` | row limit, for datasets that return lists |
| `--offset` | int | `0` | row offset, for datasets that return lists |
| `--time-unit` | string | — | bucket size: Y, m, W, d or H |
| `--with-counts` | bool | `false` | include totals alongside the series |

### `mautic data types`

List the available datasets

## `mautic dynamic-contents`

Dynamic web content blocks

Dynamic web content blocks

Wraps the Mautic /dynamiccontents endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

Aliases: dynamiccontents

### `mautic dynamic-contents create`

Create a dynamic-content

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--content` | string | — | the HTML content |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--filters` | string | — | targeting filters as JSON |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--slot-name` | string | — | slot this content fills |

### `mautic dynamic-contents delete`

Delete a dynamic-content

Delete a dynamic-content.

This cannot be undone. The command refuses to run without --yes.

```
mautic dynamic-contents delete <id>
```

### `mautic dynamic-contents get`

Get one record by ID

```
mautic dynamic-contents get <id>
```

### `mautic dynamic-contents list`

List dynamic-contents

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic dynamic-contents update`

Update a dynamic-content

Update a dynamic-content.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic dynamic-contents update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--content` | string | — | the HTML content |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--filters` | string | — | targeting filters as JSON |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--slot-name` | string | — | slot this content fills |

## `mautic emails`

Emails, including sending them to segments and contacts

Emails, including sending them to segments and contacts

Wraps the Mautic /emails endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic emails create`

Create a email

| flag | type | default | description |
| --- | --- | --- | --- |
| `--assets` | string | — | asset IDs to attach, as a JSON array |
| `--bcc` | string | — | BCC address |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email-type` | string | — | list for a segment email, template for a template email |
| `--from-address` | string | — | from address, when it differs from the configured one |
| `--from-name` | string | — | from name |
| `--html` | string | — | HTML body — pass @file.html to avoid shell quoting |
| `--language` | string | — | language locale |
| `--lists` | string | — | segment IDs to send to, as a JSON array such as [1,2] |
| `--name` | string | — | internal name |
| `--plain-text` | string | — | plain text body |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--reply-to` | string | — | reply-to address |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--subject` | string | — | subject line |
| `--template` | string | — | theme used as the base |
| `--unsubscribe-form` | int | `0` | ID of the form shown on the unsubscribe page |

### `mautic emails delete`

Delete a email

Delete a email.

This cannot be undone. The command refuses to run without --yes.

```
mautic emails delete <id>
```

### `mautic emails get`

Get one record by ID

```
mautic emails get <id>
```

### `mautic emails list`

List emails

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic emails reply`

Record that a send received a reply

Record that a send received a reply.

The tracking hash is the unique hash Mautic generates for each email send, not
the email ID — read it from email_stats.tracking_hash:

  mautic stats query email_stats --where lead_id:eq:1234 --fields id,email_id,tracking_hash

Recording a reply also marks the send as read, so it changes the open statistics
for that email. It is meant for inbound mail processors feeding replies back into
Mautic.

```
mautic emails reply <tracking-hash>
```

### `mautic emails send`

Send a segment email to everyone in its segments

Send a segment email to every contact in the segments assigned to it.

This reaches real people and cannot be recalled. The command refuses to run
without --yes; use --dry-run first to see exactly what would be sent, and note
which instance the target line names.

Only emails with emailType "list" can be broadcast this way. A template email
has to be sent per contact with "mautic emails send-to-contact".

```
mautic emails send <email-id>
```

### `mautic emails send-to-contact`

Send an email to one contact

```
mautic emails send-to-contact <email-id> <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--tokens` | string | — | token replacements as a JSON object, inline or @file.json |

### `mautic emails update`

Update a email

Update a email.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic emails update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--assets` | string | — | asset IDs to attach, as a JSON array |
| `--bcc` | string | — | BCC address |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email-type` | string | — | list for a segment email, template for a template email |
| `--from-address` | string | — | from address, when it differs from the configured one |
| `--from-name` | string | — | from name |
| `--html` | string | — | HTML body — pass @file.html to avoid shell quoting |
| `--language` | string | — | language locale |
| `--lists` | string | — | segment IDs to send to, as a JSON array such as [1,2] |
| `--name` | string | — | internal name |
| `--plain-text` | string | — | plain text body |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--reply-to` | string | — | reply-to address |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--subject` | string | — | subject line |
| `--template` | string | — | theme used as the base |
| `--unsubscribe-form` | int | `0` | ID of the form shown on the unsubscribe page |

## `mautic fields`

Custom field definitions for contacts and companies

Custom field definitions for contacts and companies.

Run this first when a write fails validation: contact fields are configured per
instance, and "mautic contacts create --set <alias>=<value>" needs the alias
exactly as it appears here.

Use --object company for company fields.

### `mautic fields create`

Create a field definition

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | field alias used in the API and in tokens; derived from the label when omitted |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--default-value` | string | — | default value |
| `--group` | string | — | which tab the field appears on: core, social, personal or professional |
| `--label` | string | — | human readable label (required) |
| `--object` | string | — | the object the field belongs to: lead or company |
| `--order` | int | `0` | display order |
| `--properties` | string | — | type-specific options as JSON, e.g. {"list":[{"label":"A","value":"a"}]} |
| `--publicly-updatable` | bool | `false` | let tracked visitors update it themselves |
| `--published` | bool | `false` | published state |
| `--required` | bool | `false` | make the field required |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--type` | string | — | field type: text, textarea, email, number, tel, url, select, multiselect, boolean, date, datetime, time, lookup, country, region, timezone, locale |
| `--unique-identifier` | bool | `false` | use the field to deduplicate records |

### `mautic fields delete`

Delete a field definition

Delete a field definition.

This drops the column and every value stored in it across all records. It cannot
be undone, and the command refuses to run without --yes.

```
mautic fields delete <field-id>
```

### `mautic fields get`

Get one field definition

```
mautic fields get <field-id>
```

### `mautic fields list`

List field definitions

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic fields update`

Update a field definition

```
mautic fields update <field-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | field alias used in the API and in tokens; derived from the label when omitted |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--default-value` | string | — | default value |
| `--group` | string | — | which tab the field appears on: core, social, personal or professional |
| `--label` | string | — | human readable label (required) |
| `--object` | string | — | the object the field belongs to: lead or company |
| `--order` | int | `0` | display order |
| `--properties` | string | — | type-specific options as JSON, e.g. {"list":[{"label":"A","value":"a"}]} |
| `--publicly-updatable` | bool | `false` | let tracked visitors update it themselves |
| `--published` | bool | `false` | published state |
| `--required` | bool | `false` | make the field required |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--type` | string | — | field type: text, textarea, email, number, tel, url, select, multiselect, boolean, date, datetime, time, lookup, country, region, timezone, locale |
| `--unique-identifier` | bool | `false` | use the field to deduplicate records |

## `mautic files`

Files in Mautic's media directories

Files in Mautic's media directories.

--dir images covers media/images and supports subdirectories; --dir assets
covers media/files, where asset files live, and has no subdirectories.

Uploading a file does not create an Asset. Upload first, then point
"mautic assets create --file <name>" at the uploaded file name.

### `mautic files delete`

Delete a file from a media directory

```
mautic files delete <filename>
```

### `mautic files list`

List files in a media directory

### `mautic files upload`

Upload a file to a media directory

```
mautic files upload <path>
```

## `mautic focus`

Focus items: pop-ups, notices and banners

Focus items: pop-ups, notices and banners

Wraps the Mautic /focus endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic focus create`

Create a focu

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | internal name |
| `--properties` | string | — | appearance and behaviour as JSON |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--style` | string | — | presentation: modal, notification, page or bar |
| `--type` | string | — | focus type: form, link, collect_and_display or link |
| `--website` | string | — | website the item is used on |

### `mautic focus delete`

Delete a focu

Delete a focu.

This cannot be undone. The command refuses to run without --yes.

```
mautic focus delete <id>
```

### `mautic focus get`

Get one record by ID

```
mautic focus get <id>
```

### `mautic focus list`

List focus

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic focus update`

Update a focu

Update a focu.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic focus update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | internal name |
| `--properties` | string | — | appearance and behaviour as JSON |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--style` | string | — | presentation: modal, notification, page or bar |
| `--type` | string | — | focus type: form, link, collect_and_display or link |
| `--website` | string | — | website the item is used on |

## `mautic forms`

Forms, their fields, actions and submissions

Forms, their fields, actions and submissions

Wraps the Mautic /forms endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic forms contact-submissions`

List one contact's submissions of a form

```
mautic forms contact-submissions <form-id> <contact-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic forms create`

Create a form

| flag | type | default | description |
| --- | --- | --- | --- |
| `--actions` | string | — | form actions as a JSON array |
| `--alias` | string | — | form alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--form-fields` | string | — | form fields as a JSON array |
| `--form-type` | string | — | standalone or campaign |
| `--name` | string | — | form name |
| `--post-action` | string | — | what happens after submit: return, redirect or message |
| `--post-action-property` | string | — | the redirect URL or the message, depending on post-action |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--render-style` | bool | `false` | let Mautic render its own stylesheet |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

### `mautic forms delete`

Delete a form

Delete a form.

This cannot be undone. The command refuses to run without --yes.

```
mautic forms delete <id>
```

### `mautic forms delete-actions`

Delete actions from a form

```
mautic forms delete-actions <form-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--action` | stringArray | `[]` | ID to delete, repeatable |

### `mautic forms delete-fields`

Delete fields from a form

```
mautic forms delete-fields <form-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--field` | stringArray | `[]` | ID to delete, repeatable |

### `mautic forms get`

Get one record by ID

```
mautic forms get <id>
```

### `mautic forms list`

List forms

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic forms submission`

Get one form submission

```
mautic forms submission <form-id> <submission-id>
```

### `mautic forms submissions`

List submissions for a form

```
mautic forms submissions <form-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic forms update`

Update a form

Update a form.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

Mautic resets a form's postAction to "return" when a PATCH body omits it, so
without --post-action this command first reads the form (a GET, which runs
even under --dry-run) and sends its current postAction and postActionProperty
back, with a note on stderr. Verify afterwards with
`mautic forms get <id> --format json`.

```
mautic forms update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--actions` | string | — | form actions as a JSON array |
| `--alias` | string | — | form alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--form-fields` | string | — | form fields as a JSON array |
| `--form-type` | string | — | standalone or campaign |
| `--name` | string | — | form name |
| `--post-action` | string | — | what happens after submit: return, redirect or message |
| `--post-action-property` | string | — | the redirect URL or the message, depending on post-action |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--render-style` | bool | `false` | let Mautic render its own stylesheet |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

## `mautic messages`

Marketing messages that span several channels

Marketing messages that span several channels

Wraps the Mautic /messages endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic messages create`

Create a message

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--channels` | string | — | per-channel configuration as JSON |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | message name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

### `mautic messages delete`

Delete a message

Delete a message.

This cannot be undone. The command refuses to run without --yes.

```
mautic messages delete <id>
```

### `mautic messages get`

Get one record by ID

```
mautic messages get <id>
```

### `mautic messages list`

List messages

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic messages update`

Update a message

Update a message.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic messages update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--channels` | string | — | per-channel configuration as JSON |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | message name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

## `mautic notes`

Notes attached to contacts

Notes attached to contacts

Wraps the Mautic /notes endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic notes create`

Create a note

| flag | type | default | description |
| --- | --- | --- | --- |
| `--contact` | int | `0` | ID of the contact the note belongs to (required) |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--datetime` | string | — | date/time the note refers to, as ISO 8601 |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--text` | string | — | note body |
| `--type` | string | — | note type: general, email, call or meeting |

### `mautic notes delete`

Delete a note

Delete a note.

This cannot be undone. The command refuses to run without --yes.

```
mautic notes delete <id>
```

### `mautic notes get`

Get one record by ID

```
mautic notes get <id>
```

### `mautic notes list`

List notes

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic notes update`

Update a note

Update a note.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic notes update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--contact` | int | `0` | ID of the contact the note belongs to (required) |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--datetime` | string | — | date/time the note refers to, as ISO 8601 |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--text` | string | — | note body |
| `--type` | string | — | note type: general, email, call or meeting |

## `mautic notifications`

Web push notifications

Web push notifications

Wraps the Mautic /notifications endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic notifications create`

Create a notification

| flag | type | default | description |
| --- | --- | --- | --- |
| `--button` | string | — | button label |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--heading` | string | — | notification heading |
| `--language` | string | — | language locale |
| `--message` | string | — | notification body |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--target-url` | string | — | URL opened when the notification is clicked |

### `mautic notifications delete`

Delete a notification

Delete a notification.

This cannot be undone. The command refuses to run without --yes.

```
mautic notifications delete <id>
```

### `mautic notifications get`

Get one record by ID

```
mautic notifications get <id>
```

### `mautic notifications list`

List notifications

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic notifications update`

Update a notification

Update a notification.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic notifications update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--button` | string | — | button label |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--heading` | string | — | notification heading |
| `--language` | string | — | language locale |
| `--message` | string | — | notification body |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--target-url` | string | — | URL opened when the notification is clicked |

## `mautic pages`

Landing pages

Landing pages

Wraps the Mautic /pages endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic pages create`

Create a page

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | URL alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--html` | string | — | HTML body — pass @file.html to avoid shell quoting |
| `--language` | string | — | language locale |
| `--meta-description` | string | — | meta description for the page head |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--redirect-type` | int | `0` | HTTP status used when redirecting: 301 or 302 |
| `--redirect-url` | string | — | URL to redirect to |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--template` | string | — | theme used as the base |
| `--title` | string | — | page title |

### `mautic pages delete`

Delete a page

Delete a page.

This cannot be undone. The command refuses to run without --yes.

```
mautic pages delete <id>
```

### `mautic pages get`

Get one record by ID

```
mautic pages get <id>
```

### `mautic pages list`

List pages

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic pages update`

Update a page

Update a page.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic pages update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | URL alias |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--html` | string | — | HTML body — pass @file.html to avoid shell quoting |
| `--language` | string | — | language locale |
| `--meta-description` | string | — | meta description for the page head |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--redirect-type` | int | `0` | HTTP status used when redirecting: 301 or 302 |
| `--redirect-url` | string | — | URL to redirect to |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--template` | string | — | theme used as the base |
| `--title` | string | — | page title |

## `mautic point-triggers`

Point triggers: what happens when a contact reaches a score

Point triggers: what happens when a contact reaches a score

Wraps the Mautic /points/triggers endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

Aliases: triggers

### `mautic point-triggers create`

Create a point-trigger

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--color` | string | — | hex colour applied to contacts that reach it |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--events` | string | — | trigger events as a JSON array |
| `--name` | string | — | trigger name |
| `--points` | int | `0` | point total at which the trigger fires |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--trigger-existing-leads` | bool | `false` | also apply to contacts who already exceed the score |

### `mautic point-triggers delete`

Delete a point-trigger

Delete a point-trigger.

This cannot be undone. The command refuses to run without --yes.

```
mautic point-triggers delete <id>
```

### `mautic point-triggers delete-events`

Delete events from a point trigger

```
mautic point-triggers delete-events <trigger-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--event` | stringArray | `[]` | ID to delete, repeatable |

### `mautic point-triggers event-types`

List the available point trigger event types

### `mautic point-triggers get`

Get one record by ID

```
mautic point-triggers get <id>
```

### `mautic point-triggers list`

List point-triggers

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic point-triggers update`

Update a point-trigger

Update a point-trigger.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic point-triggers update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--color` | string | — | hex colour applied to contacts that reach it |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--events` | string | — | trigger events as a JSON array |
| `--name` | string | — | trigger name |
| `--points` | int | `0` | point total at which the trigger fires |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--trigger-existing-leads` | bool | `false` | also apply to contacts who already exceed the score |

## `mautic points`

Point actions: the rules that award points

Point actions: the rules that award points

Wraps the Mautic /points endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic points create`

Create a point

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--delta` | int | `0` | points awarded when the action fires |
| `--description` | string | — | description |
| `--name` | string | — | action name |
| `--properties` | string | — | type-specific configuration as JSON |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--type` | string | — | action type — list them with `mautic points types` |

### `mautic points delete`

Delete a point

Delete a point.

This cannot be undone. The command refuses to run without --yes.

```
mautic points delete <id>
```

### `mautic points get`

Get one record by ID

```
mautic points get <id>
```

### `mautic points list`

List points

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic points types`

List the available point action types

### `mautic points update`

Update a point

Update a point.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic points update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--delta` | int | `0` | points awarded when the action fires |
| `--description` | string | — | description |
| `--name` | string | — | action name |
| `--properties` | string | — | type-specific configuration as JSON |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--type` | string | — | action type — list them with `mautic points types` |

## `mautic profiles`

List, switch and remove named profiles

### `mautic profiles list`

List the configured profiles

### `mautic profiles remove`

Remove a profile and its stored secrets

```
mautic profiles remove <name>
```

### `mautic profiles use`

Set the default profile

```
mautic profiles use <name>
```

## `mautic reports`

Reports defined in Mautic, and their results

Reports defined in Mautic, and their results

Wraps the Mautic /reports endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic reports get`

Get one record by ID

```
mautic reports get <id>
```

### `mautic reports list`

List reports

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic reports results`

Run a report and return its rows

Run a report and return its rows.

The rows come back under "data" with their own column set, which depends on the
report definition rather than on any Mautic entity — so --fields is usually
needed to get a readable table.

```
mautic reports results <report-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--date-from` | string | — | start of the reporting window, as YYYY-MM-DD |
| `--date-to` | string | — | end of the reporting window, as YYYY-MM-DD |
| `--limit` | int | `0` | rows per page |
| `--page` | int | `0` | page number, starting at 1 |

## `mautic roles`

User roles and their permissions

User roles and their permissions

Wraps the Mautic /roles endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic roles create`

Create a role

| flag | type | default | description |
| --- | --- | --- | --- |
| `--admin` | bool | `false` | grant full administrative access |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | role name |
| `--permissions` | string | — | permissions as JSON, e.g. {"lead:leads":["viewown","editown"]} |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

### `mautic roles delete`

Delete a role

Delete a role.

This cannot be undone. The command refuses to run without --yes.

```
mautic roles delete <id>
```

### `mautic roles get`

Get one record by ID

```
mautic roles get <id>
```

### `mautic roles list`

List roles

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic roles update`

Update a role

Update a role.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic roles update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--admin` | bool | `false` | grant full administrative access |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | role name |
| `--permissions` | string | — | permissions as JSON, e.g. {"lead:leads":["viewown","editown"]} |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

## `mautic segments`

Segments and their contact membership

Segments and their contact membership.

Mautic calls these "lists" internally, so the API returns {"lists": [...]} and a
single segment arrives as {"list": {...}}. This command hides that, but it is
worth knowing when reading raw responses from "mautic api".

Filters are the hard part: build one segment in the UI, read it back with
"mautic segments get <id> --format json", and use its filters array as the shape
for --filters.

Aliases: lists

### `mautic segments add-contact`

Add one contact to a segment

```
mautic segments add-contact <segment-id> <contact-id>
```

### `mautic segments add-contacts`

Add many contacts to a segment in one request

```
mautic segments add-contacts <segment-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--contact` | stringArray | `[]` | contact ID to add, repeatable |

### `mautic segments create`

Create a segment

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | segment alias |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--filters` | string | — | membership filters as a JSON array |
| `--name` | string | — | segment name |
| `--public` | bool | `false` | make the segment visible to every user |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

### `mautic segments delete`

Delete a segment

Delete a segment.

This cannot be undone. The command refuses to run without --yes.

```
mautic segments delete <id>
```

### `mautic segments get`

Get one record by ID

```
mautic segments get <id>
```

### `mautic segments list`

List segments

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic segments remove-contact`

Remove one contact from a segment

```
mautic segments remove-contact <segment-id> <contact-id>
```

### `mautic segments update`

Update a segment

Update a segment.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic segments update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--alias` | string | — | segment alias |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--filters` | string | — | membership filters as a JSON array |
| `--name` | string | — | segment name |
| `--public` | bool | `false` | make the segment visible to every user |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |

## `mautic smses`

Text messages

Text messages

Wraps the Mautic /smses endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

Aliases: sms, text-messages

### `mautic smses create`

Create a sms

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--language` | string | — | language locale |
| `--lists` | string | — | segment IDs to send to, as a JSON array |
| `--message` | string | — | message body |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--sms-type` | string | — | list for a segment message, template for a template message |

### `mautic smses delete`

Delete a sms

Delete a sms.

This cannot be undone. The command refuses to run without --yes.

```
mautic smses delete <id>
```

### `mautic smses get`

Get one record by ID

```
mautic smses get <id>
```

### `mautic smses list`

List smses

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic smses send`

Send a text message to one contact

Send a text message to one contact.

Mautic exposes this as a GET, which is unusual for something that sends a
message — so --dry-run does not suppress it. Check the arguments before running.

```
mautic smses send <sms-id> <contact-id>
```

### `mautic smses update`

Update a sms

Update a sms.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic smses update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--language` | string | — | language locale |
| `--lists` | string | — | segment IDs to send to, as a JSON array |
| `--message` | string | — | message body |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--sms-type` | string | — | list for a segment message, template for a template message |

## `mautic stages`

Lifecycle stages and their contact membership

Lifecycle stages and their contact membership

Wraps the Mautic /stages endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic stages add-contact`

Move a contact into a stage

```
mautic stages add-contact <stage-id> <contact-id>
```

### `mautic stages create`

Create a stage

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | stage name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--weight` | int | `0` | ordering weight; lower comes first |

### `mautic stages delete`

Delete a stage

Delete a stage.

This cannot be undone. The command refuses to run without --yes.

```
mautic stages delete <id>
```

### `mautic stages get`

Get one record by ID

```
mautic stages get <id>
```

### `mautic stages list`

List stages

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic stages remove-contact`

Remove a contact from a stage

```
mautic stages remove-contact <stage-id> <contact-id>
```

### `mautic stages update`

Update a stage

Update a stage.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic stages update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--name` | string | — | stage name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--weight` | int | `0` | ordering weight; lower comes first |

## `mautic stats`

Query Mautic's statistics tables directly

Query Mautic's statistics tables directly.

"tables" lists what can be queried and the columns of each table; "query" reads
one of them with the same where/order/limit parameters as any other list.

Column names here are database columns in snake_case (date_hit, lead_id), not
the camelCase property names used elsewhere in the API.

### `mautic stats query`

Read rows from a statistics table

```
mautic stats query <table>
```

Example:

```
# page hits since a date, newest first
  mautic stats query page_hits --where date_hit:gte:2026-08-01 --order date_hit:desc --limit 50

  # every email send recorded for one contact
  mautic stats query email_stats --where lead_id:eq:1234 --all --format csv
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic stats tables`

List the queryable tables and their columns

## `mautic tags`

Contact tags

Contact tags

Wraps the Mautic /tags endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic tags create`

Create a tag

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--tag` | string | — | the tag itself |

### `mautic tags delete`

Delete a tag

Delete a tag.

This cannot be undone. The command refuses to run without --yes.

```
mautic tags delete <id>
```

### `mautic tags get`

Get one record by ID

```
mautic tags get <id>
```

### `mautic tags list`

List tags

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic tags update`

Update a tag

Update a tag.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic tags update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--tag` | string | — | the tag itself |

## `mautic themes`

Installed themes

Installed themes.

Themes are identified by name rather than by numeric ID. "get" returns the
theme's config; "download" fetches the theme as a zip, which is what the API
actually serves for that path.

### `mautic themes create`

Install a theme from a zip archive

```
mautic themes create <zip-path>
```

### `mautic themes delete`

Delete an installed theme

```
mautic themes delete <theme-name>
```

### `mautic themes download`

Download a theme as a zip archive

```
mautic themes download <theme-name>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--output` | string | — | where to write the zip (default: <theme-name>.zip) |

### `mautic themes get`

Show one theme's config

```
mautic themes get <theme-name>
```

### `mautic themes list`

List installed themes and their config

## `mautic tweets`

Tweets used by the Twitter/X integration

Tweets used by the Twitter/X integration

Wraps the Mautic /tweets endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic tweets create`

Create a tweet

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--language` | string | — | language locale |
| `--link` | string | — | URL appended to the tweet |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--text` | string | — | tweet text, which may contain Mautic tokens |

### `mautic tweets delete`

Delete a tweet

Delete a tweet.

This cannot be undone. The command refuses to run without --yes.

```
mautic tweets delete <id>
```

### `mautic tweets get`

Get one record by ID

```
mautic tweets get <id>
```

### `mautic tweets list`

List tweets

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic tweets update`

Update a tweet

Update a tweet.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic tweets update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--language` | string | — | language locale |
| `--link` | string | — | URL appended to the tweet |
| `--name` | string | — | internal name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--text` | string | — | tweet text, which may contain Mautic tokens |

## `mautic users`

Mautic users (the people who log in, not contacts)

Mautic users (the people who log in, not contacts)

Wraps the Mautic /users endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

### `mautic users create`

Create a user

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email` | string | — | email address (required) |
| `--firstname` | string | — | first name (required) |
| `--lastname` | string | — | last name (required) |
| `--locale` | string | — | interface locale |
| `--login` | string | — | login name, stored as username (required) |
| `--position` | string | — | job title |
| `--role` | int | `0` | role ID (required) — list them with `mautic roles list` |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--signature` | string | — | email signature |
| `--timezone` | string | — | timezone, e.g. Asia/Tokyo |
| `--user-password` | string | — | password as JSON: {"password":"…","confirm":"…"} |

### `mautic users delete`

Delete a user

Delete a user.

This cannot be undone. The command refuses to run without --yes.

```
mautic users delete <id>
```

### `mautic users get`

Get one record by ID

```
mautic users get <id>
```

### `mautic users list`

List users

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic users permission-check`

Check whether a user holds a permission

```
mautic users permission-check <user-id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--permission` | string | — | permission to test, as bundle:area:level |

### `mautic users self`

Show the user the current credentials authenticate as

### `mautic users update`

Update a user

Update a user.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic users update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--email` | string | — | email address (required) |
| `--firstname` | string | — | first name (required) |
| `--lastname` | string | — | last name (required) |
| `--locale` | string | — | interface locale |
| `--login` | string | — | login name, stored as username (required) |
| `--position` | string | — | job title |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--role` | int | `0` | role ID (required) — list them with `mautic roles list` |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--signature` | string | — | email signature |
| `--timezone` | string | — | timezone, e.g. Asia/Tokyo |
| `--user-password` | string | — | password as JSON: {"password":"…","confirm":"…"} |

## `mautic webhooks`

Webhooks Mautic calls when events happen

Webhooks Mautic calls when events happen

Wraps the Mautic /hooks endpoints. Fields not covered by a flag can be set with
--set name=value, and an entire body can be supplied with --data @file.json.

Aliases: hooks

### `mautic webhooks create`

Create a webhook

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--events-order-dir` | string | — | order of queued events: ASC or DESC |
| `--name` | string | — | webhook name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--secret` | string | — | secret used to sign the payload |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--triggers` | stringArray | `[]` | event to subscribe to — list them with `mautic webhooks triggers` (repeatable) |
| `--webhook-url` | string | — | URL Mautic posts to |

### `mautic webhooks delete`

Delete a webhook

Delete a webhook.

This cannot be undone. The command refuses to run without --yes.

```
mautic webhooks delete <id>
```

### `mautic webhooks get`

Get one record by ID

```
mautic webhooks get <id>
```

### `mautic webhooks list`

List webhooks

| flag | type | default | description |
| --- | --- | --- | --- |
| `--all` | bool | `false` | fetch every page, not just the first |
| `--limit` | int | `30` | maximum number of records to return |
| `--minimal` | bool | `false` | omit the nested lists from each record, which is much faster |
| `--order` | stringArray | `[]` | advanced sort as col:dir, repeatable |
| `--order-by` | string | — | column to sort by, in snake_case (dateIdentified is date_identified) |
| `--order-dir` | string | — | sort direction: asc or desc |
| `--page-size` | int | `100` | records per request when --all is set |
| `--published-only` | bool | `false` | return only published records |
| `--search` | string | — | search string or Mautic search command to filter by |
| `--start` | int | `0` | starting row |
| `--where` | stringArray | `[]` | advanced filter as col:expr:value, repeatable (expr: eq, neq, gt, gte, lt, lte, like, in, notIn, isNull, isNotNull) |

### `mautic webhooks triggers`

List the events a webhook can subscribe to

### `mautic webhooks update`

Update a webhook

Update a webhook.

By default this sends PATCH, which changes only the fields you pass. With
--replace it sends PUT, which replaces the whole record and blanks every field
you did not provide.

```
mautic webhooks update <id>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | int | `0` | category ID — list them with `mautic categories list` |
| `--data` | string | — | request body as JSON: inline, @file.json, or @- for stdin. Field flags override it |
| `--description` | string | — | description |
| `--events-order-dir` | string | — | order of queued events: ASC or DESC |
| `--name` | string | — | webhook name |
| `--publish-down` | string | — | date/time to unpublish, as ISO 8601 |
| `--publish-up` | string | — | date/time to publish, as ISO 8601 |
| `--published` | bool | `false` | published state |
| `--replace` | bool | `false` | send PUT instead of PATCH, replacing the whole record (needs --yes) |
| `--secret` | string | — | secret used to sign the payload |
| `--set` | stringArray | `[]` | set any other property as name=value, repeatable |
| `--set-json` | stringArray | `[]` | set any other property to a JSON value as name=<json>, repeatable |
| `--triggers` | stringArray | `[]` | event to subscribe to — list them with `mautic webhooks triggers` (repeatable) |
| `--webhook-url` | string | — | URL Mautic posts to |
