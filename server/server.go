package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	source cluster.Source
	user   *user.User
}

type HandleFunc func(http.ResponseWriter, *http.Request, httprouter.Params) error

func HandleWrapper(s *Server, fn HandleFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()
		if s.user != nil {
			id, token, ok := r.BasicAuth()
			if !(s.user.ID == id && s.user.Token == token) || !ok {
				w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		err := fn(w, r, p)
		if err != nil {
			switch err.(type) {
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			log.Log.Warn("server", zap.Error(err))
		}
		if !strings.Contains(r.URL.String(), "/static") {
			log.Log.Info(
				"request",
				zap.Duration("duration", time.Since(start)),
				zap.String("url", r.URL.String()),
				zap.String("method", r.Method),
				zap.String("host", r.Header.Get("Host")),
				zap.String("user-agent", r.Header.Get("User-Agent")),
			)
		}
	}
}

func Run(server *Server, pattern string) error {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/gaffer", 302)
	})
	router.GET("/get", HandleWrapper(server, server.Get))
	router.POST("/set", HandleWrapper(server, server.Set))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", pattern)))
	return http.ListenAndServe(pattern, router)
}

func New(source cluster.Source, u *user.User) *Server {
	return &Server{source: source, user: u}
}
