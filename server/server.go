package server

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/cluster"
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
	Name     string
	Hostname string
	Response *query.Response
	Cluster  *cluster.Cluster
}

func (_ ClusterPage) Upper(s string) string { return strings.ToUpper(s) }
func (c ClusterPage) Progress() int {
	return ((int(c.Cluster.State()) + 1) / 6) * 100
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
				q := &query.Query{Type: query.READ_USER}
				q.ReadUser.User = &user.User{ID: id, Token: token}
				resp, err := s.store.Query(q)
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
	page.Name = p.ByName("cluster")
	page.Hostname = p.ByName("hostname")
	if page.Name != "" {
		for _, cluster := range resp.Clusters {
			if cluster.ID == page.Name {
				page.Cluster = cluster
			}
		}
		if page.Cluster == nil {
			http.NotFound(w, r)
			return nil
		}
	} else {
		page.Name = "clusters"
	}
	// TODO: Need a unique identifier, not hostname
	/*
		if page.Hostname != "" {
			var found bool
			for _, zk := range page.Cluster.Zookeepers {
				if zk.Hostname == page.Hostname {
					found = true
					page.Node.IP = zk.IP
					page.Node.Options = zk.Options
				}
			}
			for _, master := range page.Cluster.Masters {
				if master.Hostname == page.Hostname {
					found = true
					page.Node.IP = master.IP
					page.Node.Options = master.Options
				}
			}
			if !found {
				http.NotFound(w, r)
				return nil
			}
		}
	*/
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
	router.GET("/static/:dir/:file", HandleWrapper(server, server.Static))
	router.POST("/1/cluster", HandleWrapper(server, server.Cluster))
	log.Log.Info("server", zap.String("msg", fmt.Sprintf("Listening @%s", pattern)))
	return http.ListenAndServe(pattern, router)
}

func New(store store.Store, anon bool) *Server {
	return &Server{store, anon}
}
