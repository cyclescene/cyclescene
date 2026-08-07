# Google Calendar importer

`g-cal` reads public Google Calendars. Calendar IDs are configured at runtime,
so one job can read rides for any number of Cycle Scene cities.

Required environment variable:

```bash
GOOGLE_CALENDARS='[
  {"id":"nci0dn513lqb1rcbvfb72drqq0@group.calendar.google.com","city":"la"},
  {"id":"another-city-rides@group.calendar.google.com","city":"another-city"}
]'
```

Every entry needs a unique Calendar ID and a Cycle Scene city code. The job
fetches, geocodes, and stores events in Turso's shared rides tables. It loads
the persistent geocode cache before calling Google, so a location is only
geocoded once unless its cache record is removed.

`GOOGLE_CALENDAR_LOOKAHEAD_DAYS` controls the future import window. It defaults
to `14` days and must be between 1 and 365. The deployed job receives this from
the `calendar_lookahead_days` Terraform variable, which also defaults to `14`.

After a complete successful fetch, a future event that is no longer returned
by its Google Calendar is marked cancelled in Turso only when it falls inside
the configured import window. A Calendar API failure stops the job before this
reconciliation can run.

Geocoding uses Application Default Credentials, just like `scraperv2`. Locally,
run `gcloud auth application-default login`; the deployed job's service account
needs permission to call the Google Geocoding API.

The Calendar reader also uses Application Default Credentials; it does not use
a Google Calendar API key. Share every configured calendar with the Cloud Run
job service account's email address as a reader. In Google Workspace, domain-
wide delegation is an alternative when an administrator manages the calendars.

## Deployment

The `infra/` directory creates a Cloud Run Job and a Cloud Scheduler trigger
that runs at 3:00 AM and 3:00 PM in the Los Angeles timezone. It also outputs
the job service account email; share each calendar with that address before the
first run. Set `turso_db_url` and `turso_db_rw_token` in the uncommitted
`infra/terraform.tfvars` file before applying.

```bash
make build
make push
cd infra
cp terraform.tfvars.example terraform.tfvars
tofu init
tofu apply
```
