package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/supervisor"
	"github.com/mesanine/gaffer/user"
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
	stop    chan bool
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

func (s *Server) Name() string { return "gaffer.http-server" }

func (s *Server) Run(_ *event.EventBus) error {
	router := httprouter.New()
	router.GET("/", HandleWrapper(s, s.HTML))
	router.GET("/get", HandleWrapper(s, s.Get))
	router.POST("/set", HandleWrapper(s, s.Set))
	router.GET("/static/:dir/:file", HandleWrapper(s, s.Static))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", s.pattern)))
	errCh := make(chan error)
	go func(errCh chan error) {
		errCh <- http.ListenAndServe(s.pattern, router)
	}(errCh)
	select {
	case err := <-errCh:
		return err
	case <-s.stop:
	}
	return nil
}

func (s *Server) Stop() error {
	s.stop <- true
	return nil
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
		stop:    make(chan bool, 1),
	}, nil
}
