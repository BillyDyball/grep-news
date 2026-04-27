package db

import (
	"database/sql"
	"time"
)

type Feed struct {
	URL           string
	Title         string
	SiteURL       string
	LastFetchedAt *time.Time
	ETag          string
	LastModified  string
	ErrorCount    int
	CreatedAt     time.Time
}

type Article struct {
	ID           int64
	FeedURL      string
	GUID         string
	Link         string
	Title        string
	Author       *string
	ContentHTML  string
	PublishedAt  *time.Time
	FetchedAt    time.Time
	ReadAt       *time.Time
	BookmarkedAt *time.Time
}

type FetchLogEntry struct {
	ID        int64
	FeedURL   string
	FetchedAt time.Time
	Status    string
	NewCount  int
	Error     *string
}

type FeedWithStats struct {
	Feed
	ArticleCount int
}

func (d *DB) InsertFeed(f *Feed) error {
	_, err := d.Exec(
		`INSERT INTO feeds (url, title, site_url) VALUES (?, ?, ?)`,
		f.URL, f.Title, f.SiteURL,
	)
	return err
}

func (d *DB) FeedExists(url string) (bool, error) {
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM feeds WHERE url = ?", url).Scan(&count)
	return count > 0, err
}

func (d *DB) DeleteFeed(url string) error {
	res, err := d.Exec("DELETE FROM feeds WHERE url = ?", url)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) DeleteArticlesByFeed(url string) error {
	_, err := d.Exec("DELETE FROM articles WHERE feed_url = ?", url)
	return err
}

func (d *DB) ListFeeds() ([]Feed, error) {
	rows, err := d.Query("SELECT url, title, site_url, last_fetched_at, etag, last_modified, error_count, created_at FROM feeds ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(&f.URL, &f.Title, &f.SiteURL, &f.LastFetchedAt, &f.ETag, &f.LastModified, &f.ErrorCount, &f.CreatedAt); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (d *DB) ListFeedsWithStats() ([]FeedWithStats, error) {
	rows, err := d.Query(`
		SELECT f.url, f.title, f.site_url, f.last_fetched_at, f.etag, f.last_modified, f.error_count, f.created_at,
		       COALESCE((SELECT COUNT(*) FROM articles a WHERE a.feed_url = f.url), 0)
		FROM feeds f
		ORDER BY f.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []FeedWithStats
	for rows.Next() {
		var f FeedWithStats
		if err := rows.Scan(&f.URL, &f.Title, &f.SiteURL, &f.LastFetchedAt, &f.ETag, &f.LastModified, &f.ErrorCount, &f.CreatedAt, &f.ArticleCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (d *DB) UpsertArticle(a *Article) (isNew bool, err error) {
	res, err := d.Exec(`
		INSERT INTO articles (feed_url, guid, link, title, author, content_html, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING`,
		a.FeedURL, a.GUID, a.Link, a.Title, a.Author, a.ContentHTML, a.PublishedAt,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (d *DB) UpdateFeedAfterFetch(url string, etag, lastModified string) error {
	_, err := d.Exec(
		`UPDATE feeds SET last_fetched_at = CURRENT_TIMESTAMP, etag = ?, last_modified = ? WHERE url = ?`,
		etag, lastModified, url,
	)
	return err
}

func (d *DB) IncrementFeedErrorCount(url string) error {
	_, err := d.Exec("UPDATE feeds SET error_count = error_count + 1 WHERE url = ?", url)
	return err
}

func (d *DB) InsertFetchLog(entry *FetchLogEntry) error {
	_, err := d.Exec(
		`INSERT INTO fetch_log (feed_url, status, new_count, error) VALUES (?, ?, ?, ?)`,
		entry.FeedURL, entry.Status, entry.NewCount, entry.Error,
	)
	return err
}

type ListArticlesParams struct {
	PageSize   int
	Page       int
	SortField  string
	SortDir    string
	FeedURL    string
	Since      string
	Search     string
	Unread     bool
	Bookmarked bool
}

func (d *DB) ListArticles(p ListArticlesParams) ([]Article, error) {
	query := `SELECT id, feed_url, guid, link, title, author, content_html, published_at, fetched_at, read_at, bookmarked_at FROM articles WHERE 1=1`
	var args []interface{}

	if p.FeedURL != "" {
		query += " AND feed_url = ?"
		args = append(args, p.FeedURL)
	}
	if p.Since != "" {
		query += " AND COALESCE(published_at, fetched_at) >= ?"
		args = append(args, p.Since)
	}
	if p.Search != "" {
		query += " AND title LIKE ?"
		args = append(args, "%"+p.Search+"%")
	}
	if p.Unread {
		query += " AND read_at IS NULL"
	}
	if p.Bookmarked {
		query += " AND bookmarked_at IS NOT NULL"
	}

	sortCol := "COALESCE(published_at, fetched_at)"
	switch p.SortField {
	case "title":
		sortCol = "title"
	case "author":
		sortCol = "COALESCE(author, 'zzz')"
	case "feed":
		sortCol = "feed_url"
	}

	dir := "DESC"
	if p.SortDir == "asc" {
		dir = "ASC"
	}

	query += " ORDER BY " + sortCol + " " + dir + ", id DESC"
	query += " LIMIT ? OFFSET ?"
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.FeedURL, &a.GUID, &a.Link, &a.Title, &a.Author, &a.ContentHTML, &a.PublishedAt, &a.FetchedAt, &a.ReadAt, &a.BookmarkedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func (d *DB) GetArticle(id int64) (*Article, error) {
	var a Article
	err := d.QueryRow(
		`SELECT id, feed_url, guid, link, title, author, content_html, published_at, fetched_at, read_at, bookmarked_at FROM articles WHERE id = ?`, id,
	).Scan(&a.ID, &a.FeedURL, &a.GUID, &a.Link, &a.Title, &a.Author, &a.ContentHTML, &a.PublishedAt, &a.FetchedAt, &a.ReadAt, &a.BookmarkedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *DB) MarkArticleRead(id int64) error {
	_, err := d.Exec("UPDATE articles SET read_at = CURRENT_TIMESTAMP WHERE id = ? AND read_at IS NULL", id)
	return err
}

func (d *DB) SetBookmark(id int64, bookmarked bool) error {
	if bookmarked {
		_, err := d.Exec("UPDATE articles SET bookmarked_at = CURRENT_TIMESTAMP WHERE id = ? AND bookmarked_at IS NULL", id)
		return err
	}
	_, err := d.Exec("UPDATE articles SET bookmarked_at = NULL WHERE id = ?", id)
	return err
}

func (d *DB) PruneArticles(olderThan time.Time, dryRun bool) (int64, error) {
	if dryRun {
		var count int64
		err := d.QueryRow(
			"SELECT COUNT(*) FROM articles WHERE bookmarked_at IS NULL AND COALESCE(published_at, fetched_at) < ?",
			olderThan,
		).Scan(&count)
		return count, err
	}

	res, err := d.Exec(
		"DELETE FROM articles WHERE bookmarked_at IS NULL AND COALESCE(published_at, fetched_at) < ?",
		olderThan,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) GetConfig(key string) (string, error) {
	var val string
	err := d.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&val)
	return val, err
}

func (d *DB) SetConfig(key, value string) error {
	_, err := d.Exec("INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?", key, value, value)
	return err
}

type Stats struct {
	FeedCount      int
	ArticleCount   int
	DBSizeBytes    int64
	FetchSuccesses int
	FetchFailures  int
	LastFetchAt    *time.Time
}

func (d *DB) GetStats() (*Stats, error) {
	s := &Stats{}
	d.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&s.FeedCount)
	d.QueryRow("SELECT COUNT(*) FROM articles").Scan(&s.ArticleCount)
	d.QueryRow("SELECT COUNT(*) FROM fetch_log WHERE status = 'success'").Scan(&s.FetchSuccesses)
	d.QueryRow("SELECT COUNT(*) FROM fetch_log WHERE status != 'success'").Scan(&s.FetchFailures)
	d.QueryRow("SELECT MAX(fetched_at) FROM fetch_log").Scan(&s.LastFetchAt)
	return s, nil
}

type FeedArticleCount struct {
	FeedURL      string
	FeedTitle    string
	ArticleCount int
}

func (d *DB) GetArticleCountPerFeed() ([]FeedArticleCount, error) {
	rows, err := d.Query(`
		SELECT f.url, f.title, COUNT(a.id) 
		FROM feeds f LEFT JOIN articles a ON a.feed_url = f.url 
		GROUP BY f.url ORDER BY COUNT(a.id) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FeedArticleCount
	for rows.Next() {
		var f FeedArticleCount
		if err := rows.Scan(&f.FeedURL, &f.FeedTitle, &f.ArticleCount); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

func (d *DB) GetStaleFeedsSince(d2 time.Time) ([]Feed, error) {
	rows, err := d.Query(
		"SELECT url, title, site_url, last_fetched_at, etag, last_modified, error_count, created_at FROM feeds WHERE last_fetched_at IS NULL OR last_fetched_at < ?",
		d2,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(&f.URL, &f.Title, &f.SiteURL, &f.LastFetchedAt, &f.ETag, &f.LastModified, &f.ErrorCount, &f.CreatedAt); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (d *DB) IntegrityCheck() (string, error) {
	var result string
	err := d.QueryRow("PRAGMA integrity_check").Scan(&result)
	return result, err
}
