package article

import (
	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

type Articles []Article

// List returns the codes of every article that still has subscribers.
func (as Articles) List() (codes []string) {
	rows, err := connections.DB().Query("SELECT DISTINCT code FROM article_subscribers")
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return codes
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
			continue
		}
		codes = append(codes, code)
	}
	return codes
}

func (as Articles) String() string {
	var content string
	for _, a := range as {
		content += "\r\n\r\n" + a.String()
	}
	return content
}

func (as Articles) StringWithPushSum() string {
	var content string
	for _, a := range as {
		content += "\r\n\r\n" + a.StringWithPushSum()
	}
	return content
}
