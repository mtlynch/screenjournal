# ScreenJournal

[![CircleCI](https://circleci.com/gh/mtlynch/screenjournal.svg?style=svg)](https://circleci.com/gh/mtlynch/screenjournal)
[![Docker Pulls](https://img.shields.io/docker/pulls/mtlynch/screenjournal.svg?maxAge=604800)](https://hub.docker.com/r/mtlynch/screenjournal/)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/m/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/commits/master)
[![GitHub last commit](https://img.shields.io/github/last-commit/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/commits/master)
[![Contributors](https://img.shields.io/github/contributors/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/graphs/contributors)
[![License](http://img.shields.io/:license-agpl-blue.svg?style=flat-square)](LICENSE)

Like Goodreads but for couch potatoes

## Overview

ScreenJournal lets you rate movies and TV shows and share recommendations with friends.

## Getting started

ScreenJournal is easy to self-host.

### Pre-requisitite: TMDB API key

ScreenJournal uses TMDB for retrieving metadata about movies and TV shows.

To host a ScreenJournal server, register [a free API key](https://www.themoviedb.org/documentation/api) from TMDB.

### Running ScreenJournal with Docker (easiest)

To run ScreenJournal within a Docker container, run the following command:

```bash
SJ_TMDB_API='your-TMDB-api-key' # Replace with your own

docker run \
  --env 'PORT=4003' \
  --env "SJ_TMDB_API=${SJ_TMDB_API}" \
  --env 'SJ_REQUIRE_TLS=false' \
  --publish 4003:4003/tcp \
  --volume "${PWD}/data:/data" \
  --name screenjournal \
  mtlynch/screenjournal
```

ScreenJournal will be running at <http://localhost:4003>

### Alternative methods for installing ScreenJournal

See the [advanced installation instructions](docs/advanced-installation.md)

### Creating an admin account

After starting ScreenJournal, navigate to the web UI and click "Sign Up."

### Inviting users

Currently, ScreenJournal does not support open signups. The only way for new users to join your ScreenJournal server is if you invite them.

From the nav bar, go to Admin > Invites to create invitation URLs to share with new users.

### Adding reviews

Once you have ScreenJournal up and running, you're ready to add reviews. Click "Add Rating" from the homepage to begin writing reviews.

## Parameters

### Command-line flags

| Flag  | Meaning                 | Default Value     |
| ----- | ----------------------- | ----------------- |
| `-db` | Path to SQLite database | `"data/store.db"` |

### Environment variables

| Environment Variable | Meaning                                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------------------------- |
| `PORT`               | TCP port on which to listen for HTTP connections (defaults to 4003).                                            |
| `SJ_TMDB_API`        | (required) API key for TMDB. You can obtain a free key at [TMDB](https://www.themoviedb.org/documentation/api). |
| `SJ_BEHIND_PROXY`    | (optional) Set to `"true"` to improve logging when ScreenJournal is running behind a reverse proxy.             |
| `SJ_REQUIRE_TLS`     | (optional) Set to `"false"` to set session cookies without the Secure flag.                                     |
| `SJ_SMTP_HOST`       | (optional) Hostname of SMTP server to send notifications.                                                       |
| `SJ_SMTP_PORT`       | (optional) Port of SMTP server to send notifications.                                                           |
| `SJ_SMTP_USERNAME`   | (optional) Username for SMTP server to send notifications.                                                      |
| `SJ_SMTP_PASSWORD`   | (optional) Password for SMTP server to send notifications.                                                      |
| `SJ_BASE_URL`        | (optional) Base URL of ScreenJournal server (only used for notifications).                                      |

## Scope and future

ScreenJournal is maintained by [Michael Lynch](https://mtlynch.io) as a hobby project.

Due to time limitations, I keep ScreenJournal's scope limited to only the features that fit into my workflows. That unfortunately means that I sometimes reject proposals or contributions for perfectly good features. It's nothing against those features, but I only have bandwidth to maintain features that I use.
