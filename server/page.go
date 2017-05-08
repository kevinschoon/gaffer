package server

import (
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

type Page struct {
	Cluster        string
	State          cluster.State
	Hosts          []*host.Host
	HostDetails    *HostDetails
	ServiceDetails *ServiceDetails
}

func (p Page) Progress() int {
	progress := int(float64(p.State) / float64(3) * 100)
	if progress > 100 {
		return 100
	}
	return progress
}

func (p Page) HostID() string {
	if p.HostDetails != nil {
		return p.HostDetails.Host.ID
	}
	return ""
}

func (p Page) ServiceID() string {
	if p.ServiceDetails != nil {
		return p.ServiceDetails.Service.ID
	}
	return ""
}

func (_ Page) Recent(d time.Duration) bool {
	return d < 20*time.Second
}

type HostDetails struct {
	Host     *host.Host
	Services []*service.Service
}

type ServiceDetails struct {
	Service *service.Service
}

func (s *Server) Overview(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	data, err := Asset("www/index.html")
	if err != nil {
		return err
	}
	tmpl, err := template.New("index").Parse(string(data))
	if err != nil {
		return err
	}
	resp, err := s.store.Query(&query.Query{Read: &query.Read{}})
	if err != nil {
		return err
	}
	cluster := resp.Read.Cluster
	page := &Page{
		Cluster: cluster.ID,
		State:   cluster.State(),
		Hosts:   cluster.Hosts,
	}
	hostID := p.ByName("host")
	if hostID != "" {
		page.HostDetails = &HostDetails{}
		for _, host := range cluster.Hosts {
			if host.ID == hostID {
				page.HostDetails.Host = host
			}
			if services, ok := cluster.Services[host.ID]; ok {
				page.HostDetails.Services = services
			}
		}
	}
	serviceID := p.ByName("service")
	if serviceID != "" {
		page.ServiceDetails = &ServiceDetails{}
		if services, ok := cluster.Services[hostID]; ok {
			for _, svc := range services {
				if svc.ID == serviceID {
					page.ServiceDetails.Service = svc
				}
			}
		}
	}
	return tmpl.Execute(w, page)
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
