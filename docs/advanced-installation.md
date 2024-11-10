# Advanced Installation

## Running ScreenJournal from a precompiled binary

Linux binaries are available for ScreenJournal.

ScreenJournal runs as a single-file binary, so installation is straightforward.

First, download the binary for your architecture from the [latest release](https://github.com/mtlynch/screenjournal/releases/latest). Extract the file from the archive, and run it with the following command:

```bash
SJ_TMDB_API='your-TMDB-api-key' # Replace with your own

SJ_REQUIRE_TLS=false \
  PORT=4003 \
  SJ_TMDB_API="${SJ_TMDB_API}" \
  ./screenjournal
```

ScreenJournal will be running at <http://localhost:4003>

## Running ScreenJournal from source

```bash
SJ_TMDB_API='your-TMDB-api-key' # Replace with your own

SJ_REQUIRE_TLS=false \
  PORT=4003 \
  SJ_TMDB_API="${SJ_TMDB_API}" \
  go run ./cmd/screenjournal
```

ScreenJournal will be running at <http://localhost:4003>
