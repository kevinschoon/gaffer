package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/user"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type Status struct {
	Events int
}

type Server struct {
	user   *user.User
	port   int
	stop   chan bool
	status Status
	config config.Config
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

func (s *Server) Run(eb *event.EventBus) error {
	router := httprouter.New()
	router.GET("/status", HandleWrapper(s, s.Status))
	router.GET("/static/:dir/:file", HandleWrapper(s, s.Static))
	log.Log.Info(fmt.Sprintf("HTTP server listening @0.0.0.0:%d", s.port))
	errCh := make(chan error)
	go func(errCh chan error) {
		errCh <- http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.port), router)
	}(errCh)
	sub := event.NewSubscriber()
	eb.Subscribe(sub)
	evtCh := sub.Chan()
	for {
		select {
		case <-evtCh:
			// TODO: Basically pointless right now. Will
			// eventually populate a web UI with the status
			// of each service.
			s.status.Events++
		case err := <-errCh:
			return err
		case <-s.stop:
			return nil
		}
	}
}

func (s *Server) Stop() error {
	s.stop <- true
	return nil
}

func (s *Server) Configure(cfg config.Config) error {
	if cfg.User.User != "" {
		usr, err := user.FromString(cfg.User.User)
		if err != nil {
			return err
		}
		s.user = usr
	}
	s.port = cfg.Plugins.HTTPServer.Port
	s.status = Status{}
	s.stop = make(chan bool, 1)
	return nil
}
