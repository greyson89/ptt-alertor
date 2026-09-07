package author

import (
	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

func Subscribers(board string) (accounts []string) {
	rows, err := connections.DB().Query(
		"SELECT account FROM author_subscribers WHERE board = ?", board,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return accounts
	}
	defer rows.Close()

	for rows.Next() {
		var account string
		if err := rows.Scan(&account); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		accounts = append(accounts, account)
	}
	return accounts
}

func AddSubscriber(board, account string) error {
	_, err := connections.DB().Exec(
		"INSERT OR IGNORE INTO author_subscribers (board, account) VALUES (?, ?)", board, account,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func RemoveSubscriber(board, account string) error {
	_, err := connections.DB().Exec(
		"DELETE FROM author_subscribers WHERE board = ? AND account = ?", board, account,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func Destroy(board string) error {
	_, err := connections.DB().Exec("DELETE FROM author_subscribers WHERE board = ?", board)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}
