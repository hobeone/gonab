// Package api implements the Newznab/nZEDb api described here:
// https://github.com/nZEDb/nZEDb/blob/dev/docs/newznab_api_specification.txt
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hobeone/gonab/config"
	"github.com/hobeone/gonab/db"
	"github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
)

const (
	webroot = "/gonab"
)

// RunAPIServer sets up and starts a server to provide the NewzNab API
func RunAPIServer(cfg *config.Config) {
	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)
	n := configRoutes(dbh)
	fmt.Println("Starting server on :8078")
	n.Run(":8078")
}

func configRoutes(dbh *db.Handle) *negroni.Negroni {
	r := mux.NewRouter().PathPrefix(webroot).Subrouter()
	r.HandleFunc("/api", capsHandler).Queries("t", "caps")
	r.HandleFunc("/api", searchHandler).Queries("t", "search")
	r.HandleFunc("/api", tvSearchHandler).Queries("t", "tvsearch")
	r.HandleFunc("/getnzb", nzbDownloadHandler)
	r.HandleFunc("/", homeHandler)
	n := negroni.Classic()
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	n.Use(negronilogrus.NewCustomMiddleware(logrus.DebugLevel, &logrus.JSONFormatter{}, "web"))
	n.Use(c)
	n.Use(dbMiddleware(dbh))
	n.UseHandler(r)
	return n
}

func getDB(r *http.Request) *db.Handle {
	spew.Dump(context.GetAll)
	if rv := context.Get(r, "db"); rv != nil {
		return rv.(*db.Handle)
	}
	panic("No database in context")
}

func dbMiddleware(dbh *db.Handle) negroni.Handler {
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		context.Set(r, "db", dbh)
		next(rw, r)
	})
}

func homeHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "home")
}

func getLink(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	port := r.Header.Get("X-Forwarded-Port")
	proto := r.Header.Get("X-Forwarded-Proto")
	if host == "" {
		host, port = getHostAndPort(r)
	}
	if proto == "" {
		proto = "http"
	}
	if port != "" {
		host = fmt.Sprintf("%s:%s", host, port)
	}
	return fmt.Sprintf("%s://%s%s", proto, host, webroot)
}

// getHost tries its best to return the request host.
// From github.com/gorilla/reverse
func getHostAndPort(r *http.Request) (string, string) {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	if host != "" && strings.Contains(host, ":") {
		parts := strings.SplitN(host, ":", 2)
		return parts[0], parts[1]
	}
	return "", ""
}
