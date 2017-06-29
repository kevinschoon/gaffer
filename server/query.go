package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/host"
	"net/http"
)

func (s *Server) Set(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	cfg := &host.Config{}
	err := json.NewDecoder(r.Body).Decode(cfg)
	if err != nil {
		return err
	}
	return s.source.Set(cfg)
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	cfg, err := s.source.Get()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(cfg)
}
