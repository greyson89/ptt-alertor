package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"

	"github.com/watain666/ptt-alertor/channels/telegram"
	ctrlr "github.com/watain666/ptt-alertor/controllers"
	"github.com/watain666/ptt-alertor/jobs"
)

var (
	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	authUser      = os.Getenv("AUTH_USER")
	authPassword  = os.Getenv("AUTH_PW")
)

type myRouter struct {
	httprouter.Router
}

func (mr myRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"method": r.Method,
		"IP":     r.RemoteAddr,
		"URI":    r.URL.Path,
	}).Info("visit")
	mr.Router.ServeHTTP(w, r)
}

func newRouter() *myRouter {
	r := &myRouter{
		Router: *httprouter.New(),
	}
	r.NotFound = http.FileServer(http.Dir("public"))
	return r
}

func basicAuth(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		user, password, hasAuth := r.BasicAuth()
		if hasAuth && user == authUser && password == authPassword {
			handle(w, r, params)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func main() {
	log.Info("Start Jobs")
	startJobs()

	router := newRouter()

	router.GET("/", ctrlr.Index)
	router.GET("/telegram", ctrlr.TelegramIndex)

	router.POST("/broadcast", basicAuth(ctrlr.Broadcast))

	// boards apis
	router.GET("/boards/:boardName/articles/:code", ctrlr.BoardArticle)
	router.GET("/boards/:boardName/articles", ctrlr.BoardArticleIndex)
	router.GET("/boards", ctrlr.BoardIndex)

	// keyword apis
	router.GET("/keyword/boards", ctrlr.KeywordBoards)

	// author apis
	router.GET("/author/boards", ctrlr.AuthorBoards)

	// pushsum apis
	router.GET("/pushsum/boards", ctrlr.PushSumBoards)

	// articles apis
	router.GET("/articles", ctrlr.ArticleIndex)

	// users apis
	router.GET("/users/:account", basicAuth(ctrlr.UserFind))
	router.GET("/users", basicAuth(ctrlr.UserAll))
	router.POST("/users", basicAuth(ctrlr.UserCreate))
	router.PUT("/users/:account", basicAuth(ctrlr.UserModify))

	// telegram
	router.POST("/telegram/"+telegramToken, telegram.HandleRequest)

	// Web Server
	log.Info("Web Server Start on Port 9090")
	srv := http.Server{
		Addr:    ":9090",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServer", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Shutdown Web Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Web Server Showdown Failed")
	}
	log.Info("Web Server Was Been Shutdown")
}

func startJobs() {
	go jobs.NewChecker().Run()
	go jobs.NewPushSumChecker().Run()
	go jobs.NewCommentChecker().Run()
	go jobs.NewPttMonitor().Run()
	c := cron.New()
	c.AddJob("@every 48h", jobs.NewPushSumKeyReplacer())
	c.Start()
}

func init() {
	// for initial app
	jobs.NewPushSumKeyReplacer().Run()
	jobs.NewGenerator().Run()
	jobs.NewFetcher().Run()
	jobs.NewCategoryCleaner().Run()
}
