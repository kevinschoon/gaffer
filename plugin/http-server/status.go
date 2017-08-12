package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (s *Server) Status(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	return json.NewEncoder(w).Encode(s.status)
}
