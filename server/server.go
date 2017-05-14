package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	store store.Store
	user  *user.User
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
			case query.ErrInvalidQuery:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			log.Log.Warn("server", zap.Error(err))
		}
		log.Log.Info(
			"server",
			zap.Duration("duration", time.Since(start)),
			zap.String("url", r.URL.String()),
			zap.String("method", r.Method),
			zap.String("host", r.Header.Get("Host")),
			zap.String("user-agent", r.Header.Get("User-Agent")),
		)
	}
}

func Run(server *Server, pattern string) error {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/gaffer", 302)
	})
	router.POST("/1/query", HandleWrapper(server, server.Query))
	router.GET("/gaffer", HandleWrapper(server, server.HTML))
	router.GET("/gaffer/:host", HandleWrapper(server, server.HTML))
	router.GET("/gaffer/:host/:service", HandleWrapper(server, server.HTML))
	router.GET("/static/:dir/:file", HandleWrapper(server, server.Static))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", pattern)))
	return http.ListenAndServe(pattern, router)
}

func New(store store.Store, usr *user.User) *Server {
	return &Server{store, usr}
}
