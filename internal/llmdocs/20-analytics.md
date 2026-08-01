# Analytics: stats tables and dashboard datasets

Most questions of the form "how many, and when" are answered far better by
`mautic stats` or `mautic data` than by paging entity endpoints. Both read the
same underlying tables the Mautic dashboard uses.

Pick between them like this:

- **`mautic data get <type>`** when you want an aggregate over time — it returns
  what a dashboard widget would show, already bucketed.
- **`mautic stats query <table>`** when you want the rows themselves, to filter,
  join or export.

## `mautic stats` — raw tables

`mautic stats tables` prints the queryable tables and their columns for the
instance in hand. The set is stable across Mautic versions:

```
asset_downloads             audit_log                   campaign_lead_event_log
campaign_leads              channel_url_trackables      companies_leads
dynamic_content_lead_data   dynamic_content_stats       email_stat_replies
email_stats                 email_stats_devices         focus_stats
form_submissions            ip_addresses                lead_categories
lead_companies_change_log   lead_devices                lead_donotcontact
lead_event_log              lead_frequencyrules         lead_lists_leads
lead_points_change_log      lead_stages_change_log      lead_utmtags
page_hits                   page_redirects              point_lead_action_log
point_lead_event_log        push_notification_stats     sms_message_stats
stage_lead_action_log       tweet_stats                 video_hits
webhook_logs
```

Column names are database columns: `date_hit`, `lead_id`, `email_id`,
`url_title`. Contacts are `lead_*` here, because that is what Mautic called them
before 2.0 and the schema never changed.

The ones that answer most questions:

| Table | What it holds | Key columns |
| --- | --- | --- |
| `page_hits` | every tracked page view, including anonymous ones | `date_hit`, `lead_id`, `url`, `url_title`, `referer`, `code` |
| `email_stats` | one row per email sent | `date_sent`, `date_read`, `is_read`, `is_failed`, `email_id`, `lead_id`, `open_count` |
| `form_submissions` | form submissions | `date_submitted`, `form_id`, `lead_id` |
| `lead_points_change_log` | every point change, with its reason | `date_added`, `lead_id`, `delta`, `event_name` |
| `lead_donotcontact` | consent state and why | `lead_id`, `channel`, `reason`, `date_added` |
| `lead_utmtags` | attribution per contact | `lead_id`, `utm_source`, `utm_medium`, `utm_campaign` |
| `campaign_lead_event_log` | what a campaign did to whom | `date_triggered`, `campaign_id`, `lead_id`, `event_id` |
| `lead_lists_leads` | segment membership | `lead_id`, `leadlist_id`, `date_added`, `manually_removed` |

`page_hits.code` is the HTTP status of the tracked page. Filtering it to 200
removes the 404s that otherwise inflate every traffic count.

```bash
# yesterday's successful page views, newest first
mautic stats query page_hits \
  --where date_hit:gte:2026-07-31 --where date_hit:lt:2026-08-01 --where code:eq:200 \
  --order date_hit:desc --all --format csv > hits.csv

# how one email performed, row by row
mautic stats query email_stats --where email_id:eq:42 --all --format jsonl
```

Timestamps in these tables are **UTC**. A "yesterday" computed in local time
will not line up with a `date_hit` window unless you convert first.

## `mautic data` — dashboard datasets

`mautic data types` lists what the instance offers, grouped as the dashboard
groups them. The core set:

| Group | Types |
| --- | --- |
| Contact | `created.leads`, `created.leads.in.time`, `anonymous.vs.identified.leads`, `lead.lifetime`, `map.of.leads`, `top.lists`, `top.owners`, `top.creators`, `segments.build.time` |
| Email | `emails.in.time`, `most.sent.emails`, `most.read.emails`, `ignored.vs.read.emails`, `emails.best.hours`, `most.hit.email.redirects`, `sent.email.to.contacts`, `device.granularity.email`, `created.emails`, `upcoming.emails` |
| Page | `page.hits.in.time`, `popular.pages`, `unique.vs.returning.leads`, `dwell.times`, `device.granularity`, `created.pages` |
| Form | `submissions.in.time`, `top.submitters`, `top.submission.referrers`, `created.forms` |
| Asset | `asset.downloads.in.time`, `popular.assets`, `unique.vs.repetitive.downloads`, `created.assets` |
| Campaign | `events.in.time`, `leads.added.in.time` |
| Other | `points.in.time`, `stages.in.time`, `recent.activity`, `report` |

Common parameters:

- `--date-from` / `--date-to` as `YYYY-MM-DD`
- `--time-unit` — `Y`, `m`, `W`, `d` or `H`, deciding the bucket size
- `--dataset` — repeatable, selecting series (`sent`, `opened`, `clicked`,
  `unsubscribed`, …) for the types that support several
- `--filter name=value` — repeatable, scoping to `campaignId`, `companyId` or
  `segmentId`
- `--with-counts` — add totals alongside the series
- `--data-format raw` — unformatted numbers, which is what you want when the
  output is going into arithmetic rather than onto a chart

```bash
# daily send/open/click for July, scoped to one segment
mautic data get emails.in.time \
  --date-from 2026-07-01 --date-to 2026-07-31 --time-unit d \
  --dataset sent --dataset opened --dataset clicked \
  --filter segmentId=7 --with-counts

# the ten most visited landing pages this month
mautic data get popular.pages --date-from 2026-07-01 --date-to 2026-07-31 --limit 10
```

These always return the raw chart envelope (`labels`, `datasets`), so
`--format json` is the only sensible way to read them — a chart is not a table.

## Reports

`mautic reports list` shows reports defined in the UI; `mautic reports results
<id>` runs one. The rows have whatever columns the report definition declares,
so `--fields` is usually needed to get a readable table. Reports cannot be
created or edited through the API — build them in Mautic.
