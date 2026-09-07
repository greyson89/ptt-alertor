package counter

import (
	"database/sql"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

func Alert() (count int, err error) {
	err = connections.DB().QueryRow(
		"SELECT value FROM counters WHERE name = 'alert'",
	).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return count, err
}

func IncrAlert() error {
	_, err := connections.DB().Exec(
		`INSERT INTO counters (name, value) VALUES ('alert', 1)
		 ON CONFLICT(name) DO UPDATE SET value = value + 1`,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}
