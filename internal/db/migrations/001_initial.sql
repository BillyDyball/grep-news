CREATE TABLE feeds (
    url TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    site_url TEXT NOT NULL DEFAULT '',
    last_fetched_at DATETIME,
    etag TEXT NOT NULL DEFAULT '',
    last_modified TEXT NOT NULL DEFAULT '',
    error_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_url TEXT NOT NULL REFERENCES feeds(url) ON DELETE CASCADE,
    guid TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    author TEXT,
    content_html TEXT NOT NULL DEFAULT '',
    published_at DATETIME,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME,
    bookmarked_at DATETIME
);

-- Dedup indexes: GUID takes precedence, then link, then composite
CREATE UNIQUE INDEX idx_articles_guid ON articles(feed_url, guid) WHERE guid != '';
CREATE UNIQUE INDEX idx_articles_link ON articles(feed_url, link) WHERE guid = '' AND link != '';
CREATE UNIQUE INDEX idx_articles_composite ON articles(feed_url, title, published_at) WHERE guid = '' AND link = '' AND published_at IS NOT NULL;

CREATE INDEX idx_articles_published ON articles(published_at DESC, id DESC);
CREATE INDEX idx_articles_feed ON articles(feed_url);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE feed_tags (
    feed_url TEXT NOT NULL REFERENCES feeds(url) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (feed_url, tag_id)
);

CREATE TABLE fetch_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_url TEXT NOT NULL REFERENCES feeds(url) ON DELETE CASCADE,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL CHECK(status IN ('success', 'timeout', 'http_error', 'parse_error')),
    new_count INTEGER NOT NULL DEFAULT 0,
    error TEXT
);

CREATE INDEX idx_fetch_log_feed ON fetch_log(feed_url, fetched_at DESC);

CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
