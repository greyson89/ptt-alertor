package controllers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/models/counter"
	"github.com/watain666/ptt-alertor/myutil"
)

var tplNames = []string{
	"telegram.html",
	"tpls/head.tpl",
	"tpls/header.tpl",
	"tpls/slogan.tpl",
	"tpls/command.tpl",
	"tpls/counter.tpl",
	"tpls/footer.tpl",
	"tpls/script.tpl",
}

var templates = template.Must(template.ParseFiles(tplPaths()...))

// tplPaths resolves the template files against myutil.PublicPath() rather
// than a bare relative path, so the binary finds them next to itself even
// when launched with a different working directory.
func tplPaths() []string {
	paths := make([]string, len(tplNames))
	for i, name := range tplNames {
		paths[i] = filepath.Join(myutil.PublicPath(), name)
	}
	return paths
}

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
