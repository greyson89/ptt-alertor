package controllers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/models/counter"
)

var tpls = []string{
	"public/telegram.html",
	"public/tpls/head.tpl",
	"public/tpls/header.tpl",
	"public/tpls/slogan.tpl",
	"public/tpls/command.tpl",
	"public/tpls/counter.tpl",
	"public/tpls/footer.tpl",
	"public/tpls/script.tpl",
}

var templates = template.Must(template.ParseFiles(tpls...))

// Index Handles router "/" request
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	TelegramIndex(w, r, nil)
}

// TelegramIndex Handles router "/telegram" request
func TelegramIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := templates.ExecuteTemplate(w, "telegram.html", struct {
		URI   string
		Count []string
	}{"telegram", count()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func count() (counterStrs []string) {
	count, err := counter.Alert()
	if err != nil {
		return nil
	}
	countStrs := strings.Split((strconv.Itoa(count)), "")
	for index, num := range countStrs {
		counterStrs = append(counterStrs, num)
		if backIndex := len(countStrs) - index; backIndex != 1 && backIndex%3 == 1 {
			counterStrs = append(counterStrs, ",")
		}
	}
	return counterStrs
}
