package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func (s *Server) Static(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	d, f := p.ByName("dir"), p.ByName("file")
	if d == "" || f == "" {
		http.NotFound(w, r)
		return nil
	}
	fp := fmt.Sprintf("www/static/%s/%s", d, f)
	data, err := Asset(fp)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	split := strings.Split(fp, ".")
	switch split[len(split)-1] {
	case "css":
		w.Header().Add("Content-Type", "text/css")
	case "js":
		w.Header().Add("Content-Type", "application/javascript")
	}
	_, err = w.Write(data)
	return err
}
