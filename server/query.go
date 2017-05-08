package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/store/query"
	"net/http"
)

func (s *Server) Query(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	q := &query.Query{}
	err := json.NewDecoder(r.Body).Decode(q)
	if err != nil {
		return err
	}
	resp, err := s.store.Query(q)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}
