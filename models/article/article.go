package article

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"

	"time"

	"fmt"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/models/pushsum"
	"github.com/watain666/ptt-alertor/myutil"
)

type Article struct {
	ID               int    `json:"ID,omitempty"`
	Code             string `json:"code,omitempty"`
	Title            string
	Link             string
	Date             string    `json:"Date,omitempty"`
	Author           string    `json:"Author,omitempty"`
	Comments         Comments  `json:"comments,omitempty"`
	LastPushDateTime time.Time `json:"lastPushDateTime,omitempty"`
	Board            string    `json:"board,omitempty"`
	PushSum          int       `json:"pushSum,omitempty"`
	drive            Driver
}

type Driver interface {
	Find(code string, article *Article)
	Save(a Article) error
	Delete(code string) error
}

func NewArticle(drive Driver) *Article {
	return &Article{
		drive: drive,
	}
}

func (a Article) ParseID(Link string) (id int) {
	reg, err := regexp.Compile("https?://www.ptt.cc/bbs/.*/[GM]\\.(\\d+)\\..*")
	if err != nil {
		log.Fatal(err)
	}
	strs := reg.FindStringSubmatch(Link)
	if len(strs) < 2 {
		return 0
	}
	id, err = strconv.Atoi(strs[1])
	if err != nil {
		return 0
	}
	return id
}

func (a Article) MatchKeyword(keyword string) bool {
	if strings.Contains(keyword, "&") {
		keywords := strings.Split(keyword, "&")
		for _, keyword := range keywords {
			if !matchKeyword(a.Title, keyword) {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(keyword, "regexp:") {
		return matchRegex(a.Title, keyword)
	}
	return matchKeyword(a.Title, keyword)
}

// Exist check article exist or not
func (a Article) Exist() (bool, error) {
	var exists int
	err := connections.DB().QueryRow(
		"SELECT 1 FROM article_subscribers WHERE code = ? LIMIT 1", a.Code,
	).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return false, err
	}
	return exists == 1, nil
}

func (a Article) Find(code string) Article {
	a.drive.Find(code, &a)
	return a
}

func (a Article) Save() error {
	return a.drive.Save(a)
}

func (a Article) Destroy() error {
	if err := a.drive.Delete(a.Code); err != nil {
		return err
	}

	_, err := connections.DB().Exec("DELETE FROM article_subscribers WHERE code = ?", a.Code)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) AddSubscriber(account string) error {
	_, err := connections.DB().Exec(
		"INSERT OR IGNORE INTO article_subscribers (code, account) VALUES (?, ?)", a.Code, account,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) Subscribers() (accounts []string, err error) {
	rows, err := connections.DB().Query(
		"SELECT account FROM article_subscribers WHERE code = ?", a.Code,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
		return accounts, err
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
	return accounts, nil
}

func (a Article) RemoveSubscriber(sub string) error {
	_, err := connections.DB().Exec(
		"DELETE FROM article_subscribers WHERE code = ? AND account = ?", a.Code, sub,
	)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) String() string {
	return a.Title + "\r\n" + a.Link
}

func (a Article) StringWithPushSum() string {
	sumStr := strconv.Itoa(a.PushSum)
	if text, ok := pushsum.NumTextMap[a.PushSum]; ok {
		sumStr = text
	}
	return fmt.Sprintf("%s %s\r\n%s", sumStr, a.Title, a.Link)
}

func matchRegex(title string, regex string) bool {
	pattern := strings.TrimPrefix(regex, "regexp:")
	b, err := regexp.MatchString(pattern, title)
	if err != nil {
		return false
	}
	return b
}

func matchKeyword(title string, keyword string) bool {
	if strings.HasPrefix(keyword, "!") {
		excludeKeyword := strings.Trim(keyword, "!")
		return !containKeyword(title, excludeKeyword)
	}
	return containKeyword(title, keyword)
}

func containKeyword(title string, keyword string) bool {
	return strings.Contains(strings.ToLower(title), strings.ToLower(keyword))
}
