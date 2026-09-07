package article

import (
	"database/sql"
	"encoding/json"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

// SQLite stores article content as JSON, keyed by article code.
type SQLite struct{}

func (SQLite) Find(code string, a *Article) {
	var board string
	var content []byte
	err := connections.DB().QueryRow(
		"SELECT board, content FROM articles WHERE code = ?", code,
	).Scan(&board, &content)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		}
		return
	}

	a.Board = board
	if err := json.Unmarshal(content, a); err != nil {
		log.WithField("code", code).Error("Article Content Unmarshal Failed")
		myutil.LogJSONDecode(err, content)
	}
}

func (SQLite) Save(a Article) error {
	articleJSON, err := json.Marshal(a)
	if err != nil {
		myutil.LogJSONEncode(err, a)
		return err
	}

	_, err = connections.DB().Exec(
		`INSERT INTO articles (code, board, content) VALUES (?, ?, ?)
		 ON CONFLICT(code) DO UPDATE SET board = excluded.board, content = excluded.content`,
		a.Code, a.Board, articleJSON,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (SQLite) Delete(code string) error {
	_, err := connections.DB().Exec("DELETE FROM articles WHERE code = ?", code)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}
