package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/host"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/supervisor"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	source  host.Source
	user    *user.User
	client  *supervisor.ClientMux
	pattern string
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

func Run(server *Server) error {
	router := httprouter.New()
	router.GET("/", HandleWrapper(server, server.HTML))
	router.GET("/get", HandleWrapper(server, server.Get))
	router.POST("/set", HandleWrapper(server, server.Set))
	router.GET("/static/:dir/:file", HandleWrapper(server, server.Static))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", server.pattern)))
	return http.ListenAndServe(server.pattern, router)
}

func New(source host.Source, cfg config.Config) (*Server, error) {
	var usr *user.User
	if cfg.User.User != "" {
		u, err := user.FromString(cfg.User.User)
		if err != nil {
			return nil, err
		}
		usr = u
	}
	return &Server{
		source:  source,
		user:    usr,
		client:  supervisor.NewClientMux(source, host.Any()),
		pattern: cfg.Server.Pattern,
	}, nil
}
