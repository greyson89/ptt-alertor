package connections

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/myutil"
	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// sqlitePath returns the database file location. Defaults to ptt-alertor.db
// under myutil.StoragePath() (next to the binary, or the project root during
// `go run`/tests), overridable with SQLITE_PATH (":memory:" is valid for tests).
func sqlitePath() string {
	if path := os.Getenv("SQLITE_PATH"); path != "" {
		return path
	}
	return filepath.Join(myutil.StoragePath(), "ptt-alertor.db")
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	account TEXT PRIMARY KEY,
	data    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS boards (
	name     TEXT PRIMARY KEY,
	articles TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS articles (
	code    TEXT PRIMARY KEY,
	board   TEXT NOT NULL DEFAULT '',
	content TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS article_subscribers (
	code    TEXT NOT NULL,
	account TEXT NOT NULL,
	PRIMARY KEY (code, account)
);

CREATE TABLE IF NOT EXISTS keyword_subscribers (
	board   TEXT NOT NULL,
	account TEXT NOT NULL,
	PRIMARY KEY (board, account)
);

CREATE TABLE IF NOT EXISTS author_subscribers (
	board   TEXT NOT NULL,
	account TEXT NOT NULL,
	PRIMARY KEY (board, account)
);

CREATE TABLE IF NOT EXISTS pushsum_boards (
	board TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS pushsum_subscribers (
	board   TEXT NOT NULL,
	account TEXT NOT NULL,
	PRIMARY KEY (board, account)
);

CREATE TABLE IF NOT EXISTS pushsum_diff_ids (
	account    TEXT NOT NULL,
	board      TEXT NOT NULL,
	kind       TEXT NOT NULL,
	generation TEXT NOT NULL,
	article_id INTEGER NOT NULL,
	PRIMARY KEY (account, board, kind, generation, article_id)
);

CREATE TABLE IF NOT EXISTS counters (
	name  TEXT PRIMARY KEY,
	value INTEGER NOT NULL DEFAULT 0
);
`

func newDB() *sql.DB {
	path := sqlitePath()
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	d, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	// SQLite only supports a single writer at a time; keep one connection so
	// database/sql's pooling doesn't hand out concurrent writers that would
	// otherwise collide with SQLITE_BUSY.
	d.SetMaxOpenConns(1)
	if _, err := d.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;"); err != nil {
		log.Fatal(err)
	}
	if _, err := d.Exec(schema); err != nil {
		log.Fatal(err)
	}
	return d
}

// DB returns the shared SQLite connection, creating and migrating it on first use.
func DB() *sql.DB {
	dbOnce.Do(func() {
		db = newDB()
	})
	return db
}
