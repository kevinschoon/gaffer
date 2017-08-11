package server

import (
	"fmt"
	"github.com/containerd/go-runc"
	"github.com/julienschmidt/httprouter"
	"github.com/mesanine/gaffer/host"
	"github.com/mesanine/gaffer/service"
	"github.com/mesanine/gaffer/supervisor"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

type hostContainer struct {
	Host    host.Host
	Entries entries
}

type entry struct {
	Service service.Service
	Stats   runc.Stats
}

type entries []entry

func (e entries) Len() int           { return len(e) }
func (e entries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e entries) Less(i, j int) bool { return e[i].Service.Id < e[j].Service.Id }

func helpers(statuses []*supervisor.StatusResponse) template.FuncMap {
	hosts := []hostContainer{}
	for _, status := range statuses {
		h := hostContainer{
			Host:    *status.Host,
			Entries: entries{},
		}
		stats, _ := status.UnmarshalStats()
		for _, service := range status.Services {
			e := entry{
				Service: *service,
				Stats:   *stats[service.Id],
			}
			h.Entries = append(h.Entries, e)
		}
		hosts = append(hosts, h)
		sort.Sort(h.Entries)
	}
	return template.FuncMap{
		"title": func() string { return "gaffer" },
		"hosts": func() []hostContainer { return hosts },
	}
}

func (s *Server) HTML(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	statuses := []*supervisor.StatusResponse{}
	ch, err := s.client.Status(&supervisor.StatusRequest{})
	if err != nil {
		return err
	}
	for status := range ch {
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
