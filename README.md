# ScreenJournal

[![CircleCI](https://circleci.com/gh/mtlynch/screenjournal.svg?style=svg)](https://circleci.com/gh/mtlynch/screenjournal)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/m/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/commits/master)
[![GitHub last commit](https://img.shields.io/github/last-commit/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/commits/master)
[![Contributors](https://img.shields.io/github/contributors/mtlynch/screenjournal)](https://github.com/mtlynch/screenjournal/graphs/contributors)
[![License](http://img.shields.io/:license-agpl-blue.svg?style=flat-square)](LICENSE)

Like Goodreads but for couch potatoes

## Overview

ScreenJournal lets you rate movies you've seen and share movie recommendations with friends.

## Development status

ScreenJournal is in pre-alpha state and is not yet documented for other people to use it. If you can figure out how to use it, you're welcome to play around. I'm planning to get it to the point where it's useful to others soon.

## Parameters

### Command-line flags

| Flag  | Meaning                 | Default Value     |
| ----- | ----------------------- | ----------------- |
| `-db` | Path to SQLite database | `"data/store.db"` |

### Environment variables

You can adjust behavior of the Docker container by passing these parameters with `docker run -e`:

| Environment Variable | Meaning                                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------------------------- |
| `PORT`               | TCP port on which to listen for HTTP connections (defaults to 3001).                                            |
| `SJ_TMDB_API`        | (required) API key for TMDB. You can obtain a free key at [TMDB](https://www.themoviedb.org/documentation/api). |
| `SJ_ADMIN_USERNAME`  | (required) Username for admin user. (soon to be removed)                                                        |
| `SJ_ADMIN_PASSWORD`  | (required) Password for admin user. (soon to be removed)                                                        |
