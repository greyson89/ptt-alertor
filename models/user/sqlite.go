package user

import (
	"database/sql"
	"encoding/json"
	"errors"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

type SQLite struct{}

var connectDB = connections.DB

func (SQLite) List() (accounts []string) {
	rows, err := connectDB().Query("SELECT account FROM users")
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

func (SQLite) Exist(account string) bool {
	var exists int
	err := connectDB().QueryRow("SELECT 1 FROM users WHERE account = ?", account).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return exists == 1
}

func (SQLite) Save(account string, data interface{}) error {
	uJSON, err := json.Marshal(data)
	if err != nil {
		myutil.LogJSONEncode(err, data)
		return err
	}

	_, err = connectDB().Exec("INSERT INTO users (account, data) VALUES (?, ?)", account, uJSON)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return err
	}
	return nil
}

func (SQLite) Update(account string, user interface{}) error {
	uJSON, err := json.Marshal(user)
	if err != nil {
		myutil.LogJSONEncode(err, user)
		return err
	}

	result, err := connectDB().Exec("UPDATE users SET data = ? WHERE account = ?", uJSON, account)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("user not exist")
	}
	return nil
}

func (SQLite) Find(account string, user *User) {
	var uJSON []byte
	err := connectDB().QueryRow("SELECT data FROM users WHERE account = ?", account).Scan(&uJSON)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}

	if uJSON != nil {
		if err := json.Unmarshal(uJSON, user); err != nil {
			myutil.LogJSONDecode(err, uJSON)
		}
	}
}
