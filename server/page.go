package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	//"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/supervisor"
	"html/template"
	"net/http"
	"strings"
)

func helpers(statuses []supervisor.Response) template.FuncMap {
	return template.FuncMap{
		"title":    func() string { return "gaffer" },
		"statuses": func() []supervisor.Response { return statuses },
	}
}

func (s *Server) HTML(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	config, err := s.source.Get()
	if err != nil {
		return err
	}
	mux := supervisor.ClientMux{supervisor.Clients(config.Hosts)}
	statuses := []supervisor.Response{}
	for status := range mux.Status() {
		statuses = append(statuses, status)
	}
	var tmpl *template.Template
	for _, name := range []string{
		"www/index.html",
	} {
		raw, err := Asset(name)
		if err != nil {
			return err
		}
		if tmpl == nil {
			tmpl, err = template.New(name).Funcs(helpers(statuses)).Parse(string(raw))
			if err != nil {
				return err
			}
		} else {
			tmpl, err = template.Must(tmpl.Clone()).Parse(string(raw))
			if err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}

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
