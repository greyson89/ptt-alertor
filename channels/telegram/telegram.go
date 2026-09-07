package telegram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	log "github.com/Ptt-Alertor/logrus"

	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/command"
	"github.com/watain666/ptt-alertor/myutil"
)

var (
	bot   *tgbotapi.BotAPI
	err   error
	token = os.Getenv("TELEGRAM_TOKEN")
	host  = os.Getenv("APP_HOST")
)

func init() {
	if token == "" {
		log.Warn("Telegram token is empty, telegram channel disabled")
		return
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.WithError(err).Warn("Telegram Bot Initialize Failed, telegram channel disabled")
		return
	}
	// bot.Debug = true
	log.Info("Telegram Authorized on " + bot.Self.UserName)

	if strings.HasPrefix(strings.ToLower(host), "https://") {
		startWebhook()
		return
	}

	log.WithField("APP_HOST", host).Info("APP_HOST is not https, using long polling instead of webhook")
	startPolling()
}

// startWebhook has Telegram push updates to POST /telegram/:token. It needs
// a publicly reachable https APP_HOST (a real domain, or a tunnel such as
// Cloudflare Tunnel/ngrok) — see startPolling for the alternative that
// works without one.
func startWebhook() {
	webhookConfig := tgbotapi.NewWebhook(host + "/telegram/" + token)
	webhookConfig.MaxConnections = 100
	if _, err := bot.SetWebhook(webhookConfig); err != nil {
		log.WithError(err).Warn("Telegram Bot Set Webhook Failed, falling back to long polling")
		startPolling()
		return
	}
	log.Info("Telegram Bot Sets Webhook Success")
}

// startPolling has the bot repeatedly ask Telegram for new updates instead
// of receiving them pushed in - no public URL, port forwarding, or TLS
// certificate needed, which makes it the default when APP_HOST isn't https
// (e.g. running locally on a machine with no domain).
func startPolling() {
	if _, err := bot.RemoveWebhook(); err != nil {
		log.WithError(err).Warn("Telegram Bot Remove Webhook Failed")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.WithError(err).Error("Telegram Bot Start Long Polling Failed, telegram channel disabled")
		return
	}

	go func() {
		for update := range updates {
			processUpdate(update)
		}
	}()
	log.Info("Telegram Bot Long Polling Started")
}

// HandleRequest handles a webhook request from Telegram (startWebhook mode).
func HandleRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if bot == nil {
		http.Error(w, "telegram channel disabled", http.StatusServiceUnavailable)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Telegram Read Request Body Failed")
	}

	var update tgbotapi.Update
	json.Unmarshal(bytes, &update)
	processUpdate(update)
}

// processUpdate handles one update regardless of how it arrived (webhook
// push or long-polling pull).
func processUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		handleCallbackQuery(update)
		return
	}

	if update.Message != nil {
		if update.Message.IsCommand() {
			handleCommand(update)
			return
		}
		if update.Message.Text != "" {
			handleText(update)
			return
		}
	}
}

func handleCallbackQuery(update tgbotapi.Update) {
	var responseText string
	userID := strconv.Itoa(update.CallbackQuery.From.ID)
	switch update.CallbackQuery.Data {
	case "CANCEL":
		responseText = "取消"
	default:
		responseText = command.HandleCommand(update.CallbackQuery.Data, userID, true)
	}
	SendTextMessage(update.CallbackQuery.Message.Chat.ID, responseText)
}

// help - 所有指令清單
// list - 設定清單
// add - 新增看板關鍵字、作者、推文數
// del - 刪除看板關鍵字、作者、推文數
// showkeyboard - 顯示快捷小鍵盤
// hidekeyboard - 隱藏快捷小鍵盤
func handleCommand(update tgbotapi.Update) {
	var responseText string
	userID := strconv.Itoa(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "add", "del":
		text := update.Message.Command() + " " + update.Message.CommandArguments()
		responseText = command.HandleCommand(text, userID, true)
	case "start":
		command.HandleTelegramFollow(userID, chatID)
		responseText = "歡迎使用 Ptt Alertor\n輸入「指令」查看相關功能。\n\n觀看Demo:\nhttps://media.giphy.com/media/3ohzdF6vidM6I49lQs/giphy.gif"
	case "help":
		responseText = command.HandleCommand("help", userID, true)
	case "list":
		responseText = command.HandleCommand("list", userID, true)
	case "showkeyboard":
		showReplyKeyboard(chatID)
		return
	case "hidekeyboard":
		hideReplyKeyboard(chatID)
		return
	default:
		responseText = "I don't know the command"
	}
	SendTextMessage(chatID, responseText)
}

func handleText(update tgbotapi.Update) {
	var responseText string
	userID := strconv.Itoa(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	if match, _ := regexp.MatchString("^(刪除|刪除作者)+\\s.*\\*+", text); match {
		sendConfirmation(chatID, text)
		return
	}
	responseText = command.HandleCommand(text, userID, true)
	SendTextMessage(chatID, responseText)
}

func sendConfirmation(chatID int64, cmd string) {
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("是", cmd),
			tgbotapi.NewInlineKeyboardButtonData("否", "CANCEL"),
		))
	msg := tgbotapi.NewMessage(chatID, "確定"+cmd+"？")
	msg.ReplyMarkup = markup
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Send Confirmation Failed")
	}
}

const maxCharacters = 4096

// SendTextMessage sends text message to chatID
func SendTextMessage(chatID int64, text string) {
	for _, msg := range myutil.SplitTextByLineBreak(text, maxCharacters) {
		sendTextMessage(chatID, msg)
	}
}

func sendTextMessage(chatID int64, text string) {
	if bot == nil {
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Send Message Failed")
	}
}

func showReplyKeyboard(chatID int64) {
	if bot == nil {
		return
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("清單"),
			tgbotapi.NewKeyboardButton("推文清單"),
			tgbotapi.NewKeyboardButton("指令"),
		))
	msg := tgbotapi.NewMessage(chatID, "顯示小鍵盤")
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Show Reply Keyboard Failed")
	}
}

func hideReplyKeyboard(chatID int64) {
	if bot == nil {
		return
	}

	msg := tgbotapi.NewMessage(chatID, "隱藏小鍵盤")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Hide Reply Keyboard Failed")
	}
}
