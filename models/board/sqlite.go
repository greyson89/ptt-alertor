package board

import (
	"database/sql"
	"encoding/json"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/myutil"
)

// SQLite implements both the Driver (article storage) and Cacher (board
// registry) interfaces backed by a single "boards" table: a row's existence
// is the registry entry, its "articles" column the stored article list.
type SQLite struct{}

func (SQLite) List() (boards []string) {
	rows, err := connections.DB().Query("SELECT name FROM boards")
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return boards
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		boards = append(boards, name)
	}
	return boards
}

func (SQLite) Exist(boardName string) bool {
	var exists int
	err := connections.DB().QueryRow("SELECT 1 FROM boards WHERE name = ?", boardName).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return exists == 1
}

func (SQLite) Create(boardName string) error {
	_, err := connections.DB().Exec(
		"INSERT OR IGNORE INTO boards (name, articles) VALUES (?, '[]')", boardName,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (SQLite) Remove(boardName string) error {
	if _, err := connections.DB().Exec("DELETE FROM boards WHERE name = ?", boardName); err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return err
	}
	return nil
}

func (SQLite) GetArticles(boardName string) (articles article.Articles) {
	if boardName == "" {
		return
	}

	var articlesJSON []byte
	err := connections.DB().QueryRow(
		"SELECT articles FROM boards WHERE name = ?", boardName,
	).Scan(&articlesJSON)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DB Find Board Failed")
		}
		return
	}

	if len(articlesJSON) > 0 {
		if err := json.Unmarshal(articlesJSON, &articles); err != nil {
			myutil.LogJSONDecode(err, articlesJSON)
		}
	}
	return articles
}

func (SQLite) Save(boardName string, articles article.Articles) error {
	articlesJSON, err := json.Marshal(articles)
	if err != nil {
		myutil.LogJSONEncode(err, articles)
		return err
	}

	_, err = connections.DB().Exec(
		`INSERT INTO boards (name, articles) VALUES (?, ?)
		 ON CONFLICT(name) DO UPDATE SET articles = excluded.articles`,
		boardName, articlesJSON,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DB Save Board Failed")
	}
	return err
}

func (SQLite) Delete(boardName string) error {
	_, err := connections.DB().Exec("DELETE FROM boards WHERE name = ?", boardName)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DB Delete Board Failed")
	}
	return err
}
