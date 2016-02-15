package api

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hobeone/gonab/config"
	"github.com/hobeone/gonab/db"
	"github.com/meatballhat/negroni-logrus"
	"github.com/rs/cors"
)

var searchResponseTemplate = template.Must(template.ParseFiles("api/response.tmpl"))

// RunAPIServer sets up and starts a server to provide the NewzNab API
func RunAPIServer(cfg *config.Config) {
	r := mux.NewRouter()
	r.HandleFunc("/api", capsHandler).Queries("t", "caps")
	r.HandleFunc("/api", searchHandler).Queries("t", "search")
	r.HandleFunc("/api", tvSearchHandler).Queries("t", "tvsearch")
	r.HandleFunc("/", homeHandler)
	fmt.Println("Starting server on :8078")
	n := negroni.Classic()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	},
	)

	n.Use(negronilogrus.NewCustomMiddleware(logrus.DebugLevel, &logrus.JSONFormatter{}, "web"))
	n.Use(dbMiddleware(cfg))
	n.Use(c)
	n.UseHandler(r)
	n.Run(":8078")
}

func getDB(r *http.Request) *db.Handle {
	if rv := context.Get(r, "db"); rv != nil {
		return rv.(*db.Handle)
	}
	return nil
}

func dbMiddleware(cfg *config.Config) negroni.Handler {
	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		context.Set(r, "db", dbh)
		next(rw, r)
	})
}

func homeHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "home")
}
