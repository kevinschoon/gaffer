package server

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vektorlab/gaffer/config"
	"github.com/vektorlab/gaffer/store"
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
	Response *store.Response
	Cluster  *config.Cluster
	// TODO change to interface
	Node struct {
		IP      string
		Options []*config.Option
	}
}

func (_ ClusterPage) Upper(s string) string { return strings.ToUpper(s) }
func (c ClusterPage) Progress() int {
	return ((int(c.Cluster.State()) + 1) / 6) * 100
}

type HandleFunc func(http.ResponseWriter, *http.Request, *user.User, httprouter.Params) error

type Server struct {
	store     store.Store
	router    *httprouter.Router
	log       *zap.Logger
	Anonymous bool
}

func (s Server) Handler(fn HandleFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()
		var (
			u   *user.User
			err error
		)
		if s.Anonymous {
			u = &user.User{1, ""}
		} else {
			_, token, ok := r.BasicAuth()
			if ok {
				resp, err := s.store.Query(&store.Query{Type: store.READ_USER, User: &user.User{Token: token}})
				if err != nil {
					s.log.Warn("server", zap.String("cannot authenticate user", err.Error()))
					http.Error(w, err.Error(), 500)
					return
				}
				u = resp.User
			}
		}
		if u != nil {
			err = fn(w, r, u, p)
			if err != nil {
				s.log.Warn("server", zap.Error(err))
				http.Error(w, err.Error(), 500)
			}
		} else {
			s.log.Warn("server", zap.String("error", "user unauthorized"))
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		s.log.Info(
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
	query := &store.Query{}
	err := json.NewDecoder(r.Body).Decode(query)
	if err != nil {
		return err
	}
	query.User = u
	if query.Type == "" {
		return fmt.Errorf("must specify Type")
	}
	if query.Type == store.CREATE {
		if query.Cluster == nil {
			return fmt.Errorf("must specify cluster parameters")
		}
	}
	resp, err := s.store.Query(query)
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
	resp, err := s.store.Query(&store.Query{User: u, Type: store.READ})
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

func (s *Server) Serve() error {
	s.router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/clusters", 302)
	})
	s.router.GET("/clusters", s.Handler(s.ClusterHTML))
	s.router.GET("/clusters/:cluster", s.Handler(s.ClusterHTML))
	s.router.GET("/clusters/:cluster/:hostname", s.Handler(s.ClusterHTML))
	s.router.GET("/static/:dir/:file", s.Handler(s.Static))
	s.router.POST("/1/cluster", s.Handler(s.Cluster))
	s.log.Info("server", zap.String("msg", "Listening @0.0.0.0:8080"))
	return http.ListenAndServe(":8080", s.router)
}

func NewServer(store store.Store, logger *zap.Logger) *Server {
	return &Server{
		store:  store,
		log:    logger,
		router: httprouter.New(),
	}
}
