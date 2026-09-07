package pushsum

import (
	"database/sql"
	"strconv"
	"strings"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

var NumTextMap = map[int]string{
	100:  "爆",
	-10:  "X1",
	-20:  "X2",
	-30:  "X3",
	-40:  "X4",
	-50:  "X5",
	-60:  "X6",
	-70:  "X7",
	-80:  "X8",
	-90:  "X9",
	-100: "XX",
}

func ConvertPushCount(str string) int {
	for num, text := range NumTextMap {
		if strings.EqualFold(str, text) {
			return num
		}
	}
	cnt, err := strconv.Atoi(str)
	if err != nil {
		cnt = 0
	}
	return cnt
}

func List() (boards []string) {
	rows, err := connections.DB().Query("SELECT board FROM pushsum_boards")
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return boards
	}
	defer rows.Close()

	for rows.Next() {
		var board string
		if err := rows.Scan(&board); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		boards = append(boards, board)
	}
	return boards
}

func Exist(board string) bool {
	var exists int
	err := connections.DB().QueryRow("SELECT 1 FROM pushsum_boards WHERE board = ?", board).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return exists == 1
}

func Add(board string) error {
	_, err := connections.DB().Exec("INSERT OR IGNORE INTO pushsum_boards (board) VALUES (?)", board)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func Remove(board string) error {
	_, err := connections.DB().Exec("DELETE FROM pushsum_boards WHERE board = ?", board)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func AddSubscriber(board, account string) error {
	_, err := connections.DB().Exec(
		"INSERT OR IGNORE INTO pushsum_subscribers (board, account) VALUES (?, ?)", board, account,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func RemoveSubscriber(board, account string) error {
	_, err := connections.DB().Exec(
		"DELETE FROM pushsum_subscribers WHERE board = ? AND account = ?", board, account,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func ListSubscribers(board string) (subs []string) {
	rows, err := connections.DB().Query(
		"SELECT account FROM pushsum_subscribers WHERE board = ?", board,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return subs
	}
	defer rows.Close()

	for rows.Next() {
		var account string
		if err := rows.Scan(&account); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		subs = append(subs, account)
	}
	return subs
}

func Destroy(board string) error {
	_, err := connections.DB().Exec("DELETE FROM pushsum_subscribers WHERE board = ?", board)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

// DiffList reports which of the given article ids have not yet been notified
// about for this account/board/kind, then remembers them as notified.
//
// A "generation" is either the current accumulating set ("base") or the
// previous one kept around to avoid re-notifying right after a reset
// ("bench", see ReplaceBenchKeys). The very first call for a combination
// only establishes the baseline and reports nothing, matching the previous
// Redis-backed behaviour.
func DiffList(account, board, kind string, ids ...int) []int {
	if len(ids) == 0 {
		return []int{}
	}

	db := connections.DB()

	var baseExists int
	err := db.QueryRow(
		`SELECT 1 FROM pushsum_diff_ids
		 WHERE account = ? AND board = ? AND kind = ? AND generation = 'base' LIMIT 1`,
		account, board, kind,
	).Scan(&baseExists)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}

	known := make(map[int]bool)
	rows, err := db.Query(
		`SELECT article_id FROM pushsum_diff_ids
		 WHERE account = ? AND board = ? AND kind = ? AND generation IN ('base', 'bench')`,
		account, board, kind,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return []int{}
	}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		known[id] = true
	}
	rows.Close()

	newIDs := make([]int, 0)
	for _, id := range ids {
		if !known[id] {
			newIDs = append(newIDs, id)
		}
	}

	for _, id := range newIDs {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO pushsum_diff_ids (account, board, kind, generation, article_id)
			 VALUES (?, ?, ?, 'base', ?)`,
			account, board, kind, id,
		); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		}
	}

	if baseExists != 1 {
		return []int{}
	}
	return newIDs
}

func DelDiffList(account, board, kind string) error {
	_, err := connections.DB().Exec(
		"DELETE FROM pushsum_diff_ids WHERE account = ? AND board = ? AND kind = ?",
		account, board, kind,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

// ReplaceBenchKeys periodically rotates every account/board/kind's "base"
// generation into "bench", discarding the previous bench. This is what lets
// still-high articles be re-notified after a while instead of being
// suppressed forever.
func ReplaceBenchKeys() error {
	db := connections.DB()
	if _, err := db.Exec("DELETE FROM pushsum_diff_ids WHERE generation = 'bench'"); err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return err
	}
	if _, err := db.Exec("UPDATE pushsum_diff_ids SET generation = 'bench' WHERE generation = 'base'"); err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return err
	}
	return nil
}
