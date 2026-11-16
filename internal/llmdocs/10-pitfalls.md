# Pitfalls

Things about Mautic's API that produce wrong answers rather than errors. None of
these are visible from the command help.

## Segments are called "lists" inside Mautic

`/segments` returns `{"lists": [...]}` and a single segment arrives as
`{"list": {...}}`. The `segments` commands hide this, but it matters the moment
you use `mautic api` or read a raw response. The same rename shows up in
`contacts segments <id>` (key `lists`) and in the `--lists` flag on emails and
text messages, which takes **segment** IDs.

## Response envelopes are not uniform

Single records are wrapped in a singular key, collections in a plural one, and
the two are not derivable from the path:

| Path | Single | Collection |
| --- | --- | --- |
| `/contacts` | `contact` | `contacts` |
| `/segments` | `list` | `lists` |
| `/hooks` | `hook` | `hooks` |
| `/points/triggers` | `trigger` | `triggers` |
| `/dynamiccontents` | `dynamicContent` | `dynamicContents` |
| `/smses` | `sms` | `smses` |
| `/users/self` | *(no wrapper at all)* | — |

The wrapped commands unwrap these for you. `mautic api` does not — that is the
point of it.

## Collections come back as objects keyed by ID, not arrays

Most list endpoints return `{"total":"12","contacts":{"47":{…},"51":{…}}}`. Some
return an array instead, `total` is sometimes a string and sometimes a number,
and a few — `/contacts/list/fields` is the one that bites — return the
collection as the *entire body* with no wrapper and no `total` at all. This CLI
normalises all of it: `--format json` always gives
`{"total":N,"count":M,"items":[…]}` with items in ascending ID order.

## Some documented endpoints are gone from Mautic 7

The API documentation still describes endpoints that newer releases dropped.
`GET /contacts/<id>/events` is the one you will hit: deprecated in favour of the
activity endpoint, then removed in 7.0, so `mautic contacts events` returns 404
on a current instance. Use `mautic contacts activity` instead.

If a documented endpoint 404s while everything else works, suspect this before
suspecting credentials — check the instance version with
`mautic auth status --verify`.

## Anonymous visitors are contacts

Mautic creates a contact record for every tracked visitor, identified or not. On
a busy instance the overwhelming majority have no email and no name — which is
why `contacts list` looks mostly empty and why `total` is enormous.

```bash
mautic contacts list --where email:isNotNull --limit 20          # identified only
mautic api GET /data/anonymous.vs.identified.leads               # the ratio
```

## `--where` speaks Doctrine, not SQL

Expressions are Doctrine's: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `like`,
`notLike`, `in`, `notIn`, `isNull`, `isNotNull`, `between`. Values for `in` are
comma-separated in a single `--where`:

```bash
mautic contacts list --where 'phone:in:444444444,888888888'
mautic contacts list --where 'email:like:%@example.com'
mautic stats query page_hits --where date_hit:gte:2026-08-01 --where date_hit:lt:2026-09-01
```

Columns in `--where` are **database columns**, not response properties — the
same snake_case rule as `--order-by`.

## `--search` is a different language from `--where`

`--search` takes Mautic's own search commands, the ones the UI search box
accepts: `email:someone@example.com`, `is:anonymous`, `segment:my-alias`,
`tag:vip`. It is not a substring match over everything, and an unrecognised
command silently matches nothing rather than erroring.

## `--minimal` is worth using

Reading contacts without `--minimal` returns every field grouped four ways plus
IP history and tags. On a list of 100 that is megabytes. `--minimal` drops the
nested lists and is dramatically faster; add `--fields` to pick what you need.

## PATCH versus PUT

`update` sends PATCH and changes only what you pass. `--replace` sends PUT,
which **blanks every field you did not supply**. That is why `--replace`
requires `--yes`.

## A form PATCH resets `postAction` unless it is in the body

Mautic's `PATCH /forms/ID/edit` does not leave every absent property alone: when
the body has no `postAction`, the form's post-submit behaviour goes back to the
default `return` (reload the page). `postActionProperty` — the redirect URL or
the message — survives, so the form *looks* intact in a `get` until you notice
`postAction`. This was hit by adding a single hidden field with
`forms update ID --form-fields '[...]'`: the field was added and a redirecting
form stopped redirecting.

`mautic forms update` now guards against it. Unless you pass `--post-action`
(or `postAction` through `--set` / `--data`), it first reads the form and sends
the current `postAction` and `postActionProperty` along, printing
`note: kept the form's current postAction=...` on stderr. That read is a GET, so
it also happens under `--dry-run`, and the dry-run body shows exactly what will
be sent. `--replace` (PUT) is not guarded — it blanks everything you omit anyway.

After any form update, check the result:

```bash
mautic forms get ID --format json | jq '{postAction, postActionProperty, fields: [.fields[] | .alias]}'
```

The same API, called directly with `mautic api PATCH /forms/ID/edit`, has no
such guard.

## Unknown properties are ignored, not rejected

Sending a property Mautic does not recognise usually returns 200 with the record
unchanged. There is no error to catch, so a typo'd alias looks like a successful
write. Read the record back after a write you care about.

## Dates

Mautic accepts and returns ISO 8601, and interprets bare dates in the
instance's timezone rather than UTC. The stats tables store UTC. When a daily
figure is one day off, this is why — compare `mautic stats query` output against
`mautic data get … --time-unit d` before trusting either.

## Rate limiting is configured per instance

There is an API rate limiter, but its thresholds live in the instance's
`local.php` and are not discoverable through the API. 429s are retried
automatically with backoff, honouring `Retry-After`; past that, space the calls
out. Bulk work belongs in `contacts batch` and in `--all`, not in a loop of
single requests.

## A 500 on a write may still have saved the record

Mautic sometimes saves the record and only then fails while building the
response (seen with `forms create`: HTTP 500, yet the form exists). For that
reason the CLI retries a POST or PATCH only on 429, never on 5xx or a dropped
connection; GET, PUT and DELETE are still retried. After a failed write, look
before you try again: `mautic forms list --search name:...` (or the matching
`list`) shows whether the record was created. `--retries 0` turns retries off
completely.

## Form fields from plugins cannot be created over the API

`plugin.recaptcha`, `plugin.captcha` and other `plugin.*` field types are stored
when sent through the API, but the template that renders them is part of the
field's `customParameters`, which only the form builder in the UI sets. The form
then returns HTTP 500 wherever it is shown, including `/form/generate.js`. The
CLI refuses such fields in `forms create` and `forms update`. Create the form
without them and add the field in Mautic (Components → Forms → the form →
Fields).

## A batch 200 does not mean every record succeeded

`contacts batch create` and `batch update` return a `statusCodes` array with one
entry per record. Check it. This CLI prints a one-line summary to stderr when
any record failed, but the detail is in the response.

## `smses send` is a GET

Mautic defines sending a text message to a contact as `GET /smses/ID/contact/ID/send`.
Because it is a GET, `--dry-run` does not suppress it — the flag only guards
non-GET requests. Check the arguments before pressing enter.
