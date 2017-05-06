package server

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type ClusterPage struct {
	ClusterName string
	Hostname    string
	ServiceName string
	Response    *query.Response
}

func (c ClusterPage) Service() *service.Service {
	if host := c.Host(); host != nil {
		return host.Services[c.ServiceName]
	}
	return &service.Service{}
}

func (c ClusterPage) Host() *cluster.Host {
	for _, host := range c.Cluster().Hosts {
		if host.Hostname == c.Hostname {
			return host
		}
	}
	return &cluster.Host{}
}

func (c ClusterPage) Cluster() *cluster.Cluster {
	for _, cluster := range c.Response.Clusters {
		if cluster.ID == c.ClusterName {
			return cluster
		}
	}
	return &cluster.Cluster{}
}

func (c ClusterPage) Recent(d time.Duration) bool {
	return d < 1*time.Minute
}

func (_ ClusterPage) Upper(s string) string { return strings.ToUpper(s) }

func (c ClusterPage) ServicesRunning() int {
	if host := c.Host(); host != nil {
		var running int
		for _, service := range host.Services {
			if service.Process != nil {
				running++
			}
		}
		return (running / len(host.Services)) * 100
	}
	return 0
}

func (c ClusterPage) Progress() int {
	progress := int(float64(c.Cluster().State()) / float64(3) * 100)
	if progress > 100 {
		return 100
	}
	return progress
}

type Server struct {
	store     store.Store
	anonymous bool
}

type HandleFunc func(http.ResponseWriter, *http.Request, *user.User, httprouter.Params) error

func HandleWrapper(s *Server, fn HandleFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()
		var (
			u   *user.User
			err error
		)
		if s.anonymous {
			u = &user.User{"root", ""}
		} else {
			id, token, ok := r.BasicAuth()
			if ok {
				resp, err := s.store.Query(&query.Query{
					Type: query.READ_USER,
					User: &user.User{ID: id, Token: token},
				})
				if err != nil {
					log.Log.Warn("server", zap.String("cannot authenticate user", err.Error()))
					http.Error(w, err.Error(), 500)
					return
				}
				u = resp.User
			}
		}
		if u != nil {
			err = fn(w, r, u, p)
			if err != nil {
				log.Log.Warn("server", zap.Error(err))
				http.Error(w, err.Error(), 500)
			}
		} else {
			log.Log.Warn("server", zap.String("error", "user unauthorized"))
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		log.Log.Info(
			"server",
			zap.Duration("duration", time.Since(start)),
			zap.String("url", r.URL.String()),
			zap.String("method", r.Method),
			zap.String("host", r.Header.Get("Host")),
			zap.String("user-agent", r.Header.Get("User-Agent")),
		)
	}

}

func (s *Server) Cluster(w http.ResponseWriter, r *http.Request, u *user.User, p httprouter.Params) error {
	q := &query.Query{}
	err := json.NewDecoder(r.Body).Decode(q)
	if err != nil {
		return err
	}
	q.User = u
	if q.Type == "" {
		return fmt.Errorf("must specify Type")
	}
	if q.Type == query.CREATE {
		if q.Create.Clusters == nil {
			return fmt.Errorf("must specify cluster parameters")
		}
	}
	resp, err := s.store.Query(q)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(resp)
}

func (s *Server) ClusterHTML(w http.ResponseWriter, r *http.Request, u *user.User, p httprouter.Params) error {
	data, err := Asset("www/index.html")
	if err != nil {
		return err
	}
	tmpl, err := template.New("index").Parse(string(data))
	if err != nil {
		return err
	}
	resp, err := s.store.Query(&query.Query{User: u, Type: query.READ})
	if err != nil {
		return err
	}
	page := &ClusterPage{Response: resp}
	page.ClusterName = p.ByName("cluster")
	page.Hostname = p.ByName("hostname")
	page.ServiceName = p.ByName("service")

	return tmpl.Execute(w, page)
}

func (s *Server) Static(w http.ResponseWriter, r *http.Request, u *user.User, p httprouter.Params) error {
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

func Run(server *Server, pattern string) error {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/clusters", 302)
	})
	router.GET("/clusters", HandleWrapper(server, server.ClusterHTML))
	router.GET("/clusters/:cluster", HandleWrapper(server, server.ClusterHTML))
	router.GET("/clusters/:cluster/:hostname", HandleWrapper(server, server.ClusterHTML))
	router.GET("/clusters/:cluster/:hostname/:service", HandleWrapper(server, server.ClusterHTML))
	router.GET("/static/:dir/:file", HandleWrapper(server, server.Static))
	router.POST("/1/cluster", HandleWrapper(server, server.Cluster))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", pattern)))
	return http.ListenAndServe(pattern, router)
}

func New(store store.Store, anon bool) *Server {
	return &Server{store, anon}
}
