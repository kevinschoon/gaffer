package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HandleFunc func(http.ResponseWriter, *http.Request, *User, httprouter.Params) error

type Server struct {
	store     Store
	router    *httprouter.Router
	log       *zap.Logger
	Anonymous bool
}

func (s Server) Handler(fn HandleFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()
		var (
			u   *User
			err error
		)
		if s.Anonymous {
			u = &User{1, ""}
		} else {
			_, token, ok := r.BasicAuth()
			if ok {
				resp, err := s.store.Query(&Query{Type: READ_USER, User: &User{Token: token}})
				if err != nil {
					s.log.Warn("server", zap.String("cannot authenticate user", err.Error()))
					http.Error(w, err.Error(), 500)
					return
				}
				u = resp.User
			}
		}
		if u != nil {
			err = fn(w, r, u, p)
			if err != nil {
				s.log.Warn("server", zap.Error(err))
				http.Error(w, err.Error(), 500)
			}
		} else {
			s.log.Warn("server", zap.String("error", "user unauthorized"))
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		s.log.Info(
			"server",
			zap.Time("ts", start),
			zap.Duration("duration", time.Since(start)),
			zap.String("url", r.URL.String()),
			zap.String("method", r.Method),
			zap.String("host", r.Header.Get("Host")),
			zap.String("user-agent", r.Header.Get("User-Agent")),
		)
	}
}
func (s *Server) Cluster(w http.ResponseWriter, r *http.Request, u *User, p httprouter.Params) error {
	query := &Query{}
	err := json.NewDecoder(r.Body).Decode(query)
	if err != nil {
		return err
	}
	query.User = u
	if query.Type == "" {
		return fmt.Errorf("must specify Type")
	}
	if query.Type == CREATE {
		if query.Cluster == nil {
			return fmt.Errorf("must specify cluster parameters")
		}
	}
	resp, err := s.store.Query(query)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s *Server) Serve() error {
	s.router.POST("/1/cluster", s.Handler(s.Cluster))
	s.log.Info("server", zap.String("msg", "Listening @0.0.0.0:8080"))
	return http.ListenAndServe(":8080", s.router)
}

func NewServer(store Store, logger *zap.Logger) *Server {
	return &Server{
		store:  store,
		log:    logger,
		router: httprouter.New(),
	}
}
