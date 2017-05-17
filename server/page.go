package server

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/store/query"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type Data struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Data                 []int    `json:"data"`
	BackgroundColor      []string `json:"backgroundColor"`
	HoverBackgroundColor []string `json:"hoverBackgroundColor"`
}

func helpers(c *cluster.Cluster, p httprouter.Params) template.FuncMap {
	return template.FuncMap{
		"title": func() string { return c.ID },
		"param": func(key string) string { return p.ByName(key) },
		"state": func() string { return c.State().String() },
		"progress": func() int {
			progress := int(float64(c.State()) / float64(3) * 100)
			if progress > 100 {
				return 100
			}
			return progress
		},
		"data": func() template.JS {
			var (
				running  int
				degraded int
			)
			data := &Data{
				Labels: []string{"Running", "Faulted"},
				Datasets: []Dataset{
					Dataset{
						Data:                 []int{running, degraded},
						BackgroundColor:      []string{"#27ba4d", "#d9534f"},
						HoverBackgroundColor: []string{"#27ba4d", "#d9534f"},
					},
				},
			}
			raw, err := json.Marshal(data)
			if err != nil {
				return template.JS("")
			}
			return template.JS(raw)
		},
		"degraded": func(i interface{}) bool {
			switch t := i.(type) {
			case *host.Host:
				return t.TimeSinceLastContacted() > 20*time.Second
			}
			return true
		},
		"host": func() *host.Host {
			for _, host := range c.Hosts {
				if host.ID == p.ByName("host") {
					return host
				}
			}
			return nil
		},
		"hosts": func() []*host.Host { return c.Hosts },
		"service": func() *service.Service {
			for _, host := range c.Hosts {
				if p.ByName("host") == host.ID {
					for _, service := range c.Services[host.ID] {
						if service.ID == p.ByName("service") {
							return service
						}
					}
				}
			}
			return nil
		},
		"services": func() map[string][]*service.Service { return c.Services },
	}
}

func (s *Server) HTML(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	resp, err := s.store.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return err
	}
	var tmpl *template.Template
	for _, name := range []string{
		"www/index.html",
		"www/overview.html",
		"www/host.html",
		"www/service.html",
	} {
		raw, err := Asset(name)
		if err != nil {
			return err
		}
		if tmpl == nil {
			tmpl, err = template.New(name).Funcs(helpers(resp.Read.Cluster, p)).Parse(string(raw))
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
