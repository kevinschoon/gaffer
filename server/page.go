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
)

const chartScript = `
var ctx = document.getElementById("%s");
var myPieChart = new Chart(ctx,{
		type: '%s',
		data: %s,
		options: {animation: false}
});
`

var colors = []string{"#2C3E50", "#18BC9C", "#E74C3C", "#F39C12", "#3498DB"}

type Data struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

func (d Data) JS(chartType, elementID string) template.JS {
	raw, err := json.Marshal(d)
	if err != nil {
		return template.JS(fmt.Sprintf("console.log(\"ERROR: %s\")", err.Error()))
	}
	return template.JS(fmt.Sprintf(chartScript, elementID, chartType, string(raw)))
}

type Dataset struct {
	Data                 []int    `json:"data"`
	BackgroundColor      []string `json:"backgroundColor"`
	HoverBackgroundColor []string `json:"hoverBackgroundColor"`
}

func helpers(c *cluster.Cluster, pl cluster.ProcessList, p httprouter.Params) template.FuncMap {
	stats := c.Stats(pl)
	return template.FuncMap{
		"title": func() string { return c.ID },
		"home":  func() bool { return p.ByName("host") == "" },
		"param": func(key string) string { return p.ByName(key) },
		"state": func() string { return c.State(pl).String() },
		"progress": func() int {
			progress := int(float64(c.State(pl)) / float64(3) * 100)
			if progress > 100 {
				return 100
			}
			return progress
		},
		"chart": func() template.JS {
			if p.ByName("host") != "" {
				return Data{
					Labels: []string{"Started", "Stopped"},
					Datasets: []Dataset{
						Dataset{
							Data:                 []int{stats.Hosts[p.ByName("host")].Started, stats.Hosts[p.ByName("host")].Stopped},
							BackgroundColor:      []string{"#18BC9C", "#E74C3C"},
							HoverBackgroundColor: []string{"#18BC9C", "#E74C3C"},
						},
					},
				}.JS("pie", "chart")
			} else {
				points := []int{}
				labels := []string{}
				for name, host := range stats.Hosts {
					labels = append(labels, name)
					points = append(points, host.Started)
				}
				return Data{
					Labels: labels,
					Datasets: []Dataset{
						Dataset{
							Data:                 points,
							BackgroundColor:      colors,
							HoverBackgroundColor: colors,
						},
					},
				}.JS("polarArea", "chart")
			}
		},
		"started":  func(hostID, serviceID string) bool { return pl.Started(hostID, serviceID) },
		"host":     func() *host.Host { return c.Host(p.ByName("host")) },
		"hosts":    func() []*host.Host { return c.Hosts },
		"service":  func() *service.Service { return c.Service(p.ByName("host"), p.ByName("service")) },
		"services": func() map[string][]*service.Service { return c.Services },
		"selected": func(hostID, serviceID string) bool {
			return p.ByName("host") == hostID && p.ByName("service") == serviceID
		},
	}
}

func (s *Server) HTML(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	resp, err := s.store.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return err
	}
	pl, err := s.client.Processes()
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
			tmpl, err = template.New(name).Funcs(helpers(resp.Read.Cluster, pl, p)).Parse(string(raw))
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
