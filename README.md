# grep-news

A terminal RSS/Atom feed reader.

## Set up

Clone this repo and build the project.

## Supported formats

- RSS 2.0
- Atom 1.0

## Examples

### Add feed

Feeds are validated and added to a local SQLite database. Multiple URLs can be added at once.

```sh
grep-news add https://somenews.com/rss.xml https://anothernews.com/feed
```

Behaviour:

- Duplicate URLs are rejected with a warning.
- The feed is fetched on add to validate it is reachable and parseable.
- Redirected URLs are followed; the final URL is stored.
- Invalid or unreachable feeds are rejected with an error message.

### Remove feed

Feeds are removed from the SQLite database. By default, previously fetched articles are kept.

```sh
grep-news remove https://somenews.com/rss.xml
```

Use `--purge` to also delete all articles associated with the feed:

```sh
grep-news remove --purge https://somenews.com/rss.xml
```

### Fetch

The fetch command fetches all feeds and upserts articles into the SQLite database. Articles are deduplicated using the feed item GUID when available, falling back to the canonical link, then to `(feed_id, title, published_at)`.

HTTP caching headers (`ETag`, `Last-Modified`) are stored per feed and used for conditional requests on subsequent fetches to reduce bandwidth and be polite to servers.

In an interactive terminal, fetch displays a live TUI-style progress view:

```sh
grep-news fetch
● [FETCHED]   somenews.com        (3 new articles)
? [FETCHING]  anothernews.com
✗ [ERROR]     problem.org         - HTTP 404 Not Found
```

In non-interactive environments (pipes, cron, CI), fetch outputs plain line-oriented logs instead.

Feeds that error do not stop the overall fetch. Errors are categorised as:

- **Network timeout** — server did not respond in time.
- **HTTP error** — non-2xx status code.
- **Parse error** — response is not valid RSS/Atom XML.

#### Exit codes

| Code | Meaning                          |
|------|----------------------------------|
| 0    | All feeds fetched successfully   |
| 1    | Some feeds failed                |
| 2    | Invalid usage or database error  |

### List articles

The list command presents the 10 most recent articles from the database, sorted by `published_at DESC, id DESC`. When `published_at` is missing from a feed item, `fetched_at` is used as a fallback.

Listing articles is **read-only** — it does not change any state.

```sh
grep-news list
```

Default output format:

```
YYYY-MM-DD  <LINK>  <TITLE>
```

#### Output formats

| Flag      | Description                         |
|-----------|-------------------------------------|
| `--table` | Human-readable aligned columns      |
| `--csv`   | CSV with proper quoting/escaping    |
| `--json`  | JSON array for scripting            |

#### Filtering and pagination

```sh
grep-news list --size 20 --page 2 --sort published_at
```

| Flag              | Description                                      |
|-------------------|--------------------------------------------------|
| `--size N`        | Number of articles per page (default: 10)        |
| `--page N`        | Page number (default: 1)                         |
| `--sort FIELD`    | Sort by field: `published_at` (default), `title`, `author`, `feed` |
| `--feed URL`      | Filter to a specific feed                        |
| `--since DATE`    | Show articles published after DATE               |
| `--search TERM`   | Search article titles                            |

Null values (e.g. missing author) sort last.

### List feeds

The feeds command lists all subscribed feeds with health stats: last successful fetch time, total error count, and article count.

```sh
grep-news feeds
```

Default output format:

```
<URL>  <TITLE>  <LAST_FETCHED_AT>  <ARTICLE_COUNT>  <ERROR_COUNT>
```

Output format flags (`--table`, `--csv`, `--json`) behave the same as in `list`.

### Import and export

Feeds can be imported from and exported to OPML files for portability between feed readers.

```sh
grep-news import feeds.opml
grep-news export > feeds.opml
```

Behaviour:

- On import, feeds that already exist are skipped with a warning.
- On import, each feed is fetched to validate it is reachable and parseable, just like `add`.
- On export, all subscribed feeds are written as a valid OPML 2.0 document to stdout.

### Read article

The read command renders an article's content in the terminal. Content is converted from HTML to a readable plain-text format.

```sh
grep-news read 42
```

If the article has no inline content, the link is displayed instead.

### Open article

The open command opens an article's link in the default browser.

```sh
grep-news open 42
```

### Bookmark article

The bookmark command marks an article as bookmarked. Bookmarked articles can be listed with `--bookmarked`.

```sh
grep-news bookmark 42
grep-news list --bookmarked
```

Use `--remove` to remove a bookmark:

```sh
grep-news bookmark --remove 42
```

### Read/unread tracking

Articles are marked as read when viewed with `read` or `open`. Unread articles can be listed with `--unread`.

```sh
grep-news list --unread
```

#### Additional list filters

| Flag              | Description                                      |
|-------------------|--------------------------------------------------|
| `--unread`        | Show only unread articles                        |
| `--bookmarked`    | Show only bookmarked articles                    |
| `--asc`           | Sort in ascending order                          |
| `--desc`          | Sort in descending order (default)               |

### Shell completions

Shell completions can be generated for bash, zsh, and fish.

```sh
grep-news completions bash > /etc/bash_completion.d/grep-news
grep-news completions zsh > ~/.zfunc/_grep-news
grep-news completions fish > ~/.config/fish/completions/grep-news.fish
```

### Configuration

The config command manages default settings such as page size, output format, and fetch timeout.

```sh
grep-news config set page_size 20
grep-news config set format table
grep-news config get page_size
```

The database path can be overridden with the `$GREP_NEWS_DB` environment variable or the `--db` flag:

```sh
grep-news --db /path/to/feeds.db list
GREP_NEWS_DB=/path/to/feeds.db grep-news list
```

### Colour output

Output is coloured automatically when connected to an interactive terminal. This can be overridden:

```sh
grep-news list --color
grep-news list --no-color
```

### Doctor

The doctor command checks the health of the local database and subscribed feeds.

```sh
grep-news doctor
```

Checks performed:

- Database integrity (`PRAGMA integrity_check`).
- Feeds that have not been successfully fetched in over 7 days.
- Feeds whose URLs are no longer reachable.

### Stats

The stats command prints a summary of the local database.

```sh
grep-news stats
```

Output includes:

- Total number of feeds and articles.
- Article count per feed.
- Database file size.
- Fetch history summary (successes, failures, last run).

### Prune

The prune command removes old articles from the database. Bookmarked articles are never pruned.

```sh
grep-news prune --older-than 90d
```

Use `--dry-run` to preview what would be removed without deleting anything:

```sh
grep-news prune --older-than 90d --dry-run
```

### Man page

A man page can be generated from the CLI definition:

```sh
grep-news man > grep-news.1
```

## Data model

### Feeds

| Column           | Description                        |
|------------------|------------------------------------|
| `url`            | Canonical feed URL (primary key)   |
| `title`          | Feed title                         |
| `site_url`       | Feed's website URL                 |
| `last_fetched_at`| Timestamp of last successful fetch |
| `etag`           | HTTP ETag for conditional requests |
| `last_modified`  | HTTP Last-Modified header value    |

### Articles

| Column         | Description                                      |
|----------------|--------------------------------------------------|
| `id`           | Auto-incrementing ID                             |
| `feed_url`     | Foreign key to feeds                             |
| `guid`         | Feed item GUID (used for dedup)                  |
| `link`         | Article URL (canonicalised, tracking params stripped) |
| `title`        | Article title (HTML stripped)                    |
| `author`       | Author name (nullable)                           |
| `published_at` | Publication timestamp (nullable)                 |
| `fetched_at`   | When the article was first stored                |
| `read_at`      | When the article was read (nullable)             |
| `bookmarked_at`| When the article was bookmarked (nullable)       |

### Tags

| Column         | Description                                      |
|----------------|--------------------------------------------------|
| `id`           | Auto-incrementing ID                             |
| `name`         | Tag name (unique)                                |

### Feed tags

| Column         | Description                                      |
|----------------|--------------------------------------------------|
| `feed_url`     | Foreign key to feeds                             |
| `tag_id`       | Foreign key to tags                              |

### Fetch log

| Column         | Description                                      |
|----------------|--------------------------------------------------|
| `id`           | Auto-incrementing ID                             |
| `feed_url`     | Foreign key to feeds                             |
| `fetched_at`   | Timestamp of the fetch attempt                   |
| `status`       | Result status (success, timeout, http_error, parse_error) |
| `new_count`    | Number of new articles stored                    |
| `error`        | Error message (nullable)                         |
