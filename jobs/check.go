package jobs

import (
	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/channels/telegram"
	"github.com/watain666/ptt-alertor/models/counter"
)

const workers = 300

var ckCh = make(chan check)

func init() {
	for i := 0; i < workers; i++ {
		go messageWorker(ckCh)
	}
}

func messageWorker(ckCh chan check) {
	for {
		ck := <-ckCh
		sendMessage(ck)
	}
}

type check interface {
	String() string
	Self() Checker
	Stop()
	Run()
}

func sendMessage(c check) {
	cr := c.Self()
	account := cr.Profile.Account
	if cr.Profile.Telegram == "" {
		return
	}
	sendTelegram(c)
	counter.IncrAlert()
	log.WithFields(log.Fields{
		"account":  account,
		"platform": "telegram",
		"board":    cr.board,
		"type":     cr.subType,
		"word":     cr.word,
	}).Info("Message Sent")
}

func sendTelegram(c check) {
	cr := c.Self()
	telegram.SendTextMessage(cr.Profile.TelegramChat, c.String())
}
