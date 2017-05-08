package sql

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/host"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store/query"
	"go.uber.org/zap"
)

const initStmt = `
CREATE TABLE cluster (id STRING NOT NULL PRIMARY KEY);
CREATE TABLE hosts (id STRING NOT NULL PRIMARY KEY, data BLOB);
CREATE TABLE services (id STRING, host_id STRING NOT NULL, data BLOB);
`

func maybeRollback(tx *sql.Tx, err error) {
	if err != nil && tx != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			log.Log.Info("DB", zap.Error(txErr))
		}
	}
}

type SQLStore struct {
	db *sql.DB
}

func (s *SQLStore) cluster() (*cluster.Cluster, error) {
	var config *cluster.Cluster
	rows, err := s.db.Query("SELECT * FROM cluster LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		config = &cluster.Cluster{
			Hosts:    []*host.Host{},
			Services: map[string][]*service.Service{},
		}
		err = rows.Scan(&config.ID)
		if err != nil {
			return nil, err
		}
		break
	}
	return config, nil
}

func (s *SQLStore) create(q *query.Query) (*query.Response, error) {
	var (
		tx  *sql.Tx
		err error
	)
	defer maybeRollback(tx, err)
	tx, err = s.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO hosts(id, data) values(?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for _, host := range q.Create.Cluster.Hosts {
		raw, err := json.Marshal(host)
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(host.ID, raw)
		if err != nil {
			return nil, err
		}
	}
	stmt, err = tx.Prepare("INSERT INTO services(id, host_id, data) values(?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for hostID, services := range q.Create.Cluster.Services {
		for _, service := range services {
			raw, err := json.Marshal(service)
			if err != nil {
				return nil, err
			}
			_, err = stmt.Exec(service.ID, hostID, raw)
			if err != nil {
				return nil, err
			}
		}
	}
	return &query.Response{Create: &query.CreateResponse{Cluster: q.Create.Cluster}}, tx.Commit()
}

func (s *SQLStore) read(q *query.Query) (*query.Response, error) {
	config, err := s.cluster()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query("SELECT * FROM hosts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		var (
			id   string
			data string
		)
		err = rows.Scan(&id, &data)
		if err != nil {
			return nil, err
		}
		config.Hosts = append(config.Hosts, &host.Host{ID: id})
		err = json.Unmarshal([]byte(data), config.Hosts[i])
		if err != nil {
			return nil, err
		}
	}
	rows, err = s.db.Query("SELECT * FROM services")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id     string
			hostID string
			data   []byte
		)
		err = rows.Scan(&id, &hostID, &data)
		if err != nil {
			return nil, err
		}
		if _, ok := config.Services[hostID]; !ok {
			config.Services[hostID] = []*service.Service{}
		}
		svc := &service.Service{ID: id}
		config.Services[hostID] = append(config.Services[hostID], svc)
		err = json.Unmarshal(data, svc)
		if err != nil {
			return nil, err
		}
	}
	return &query.Response{Read: &query.ReadResponse{config}}, nil
}

func (s *SQLStore) update(q *query.Query) (*query.Response, error) {
	resp := &query.Response{Update: &query.UpdateResponse{}}
	if q.Update.Host != nil {
		raw, err := json.Marshal(q.Update.Host)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Exec("UPDATE hosts SET data = ? WHERE id = ?", raw, q.Update.Host.ID)
		if err != nil {
			return nil, err
		}
		resp.Update.Host = q.Update.Host
	}
	if q.Update.Service != nil {
		raw, err := json.Marshal(q.Update.Service)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Exec(
			"UPDATE services SET data = ? WHERE id = ? AND host_id = ?",
			raw,
			q.Update.Service.ID,
			q.Update.Host.ID,
		)
		if err != nil {
			return nil, err
		}
		resp.Update.Service = q.Update.Service
	}
	return resp, nil
}

func (s *SQLStore) delete(q *query.Query) (*query.Response, error) {
	switch {
	case q.Delete.HostID != "":
		_, err := s.db.Exec("DELETE FROM hosts WHERE id = ?", q.Delete.HostID)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Exec("DELETE FROM services where host_id = ?", q.Delete.HostID)
		if err != nil {
			return nil, err
		}
	case q.Delete.ServiceID != "" && q.Delete.HostID != "":
		_, err := s.db.Exec("DELETE FROM services WHERE id = ? AND host_id = ?", q.Delete.ServiceID, q.Delete.HostID)
		if err != nil {
			return nil, err
		}
	}
	return &query.Response{Delete: &query.DeleteResponse{}}, nil
}

func (s *SQLStore) Query(q *query.Query) (*query.Response, error) {
	err := query.Validate(q)
	if err != nil {
		return nil, err
	}
	switch {
	case q.Create != nil:
		return s.create(q)
	case q.Read != nil:
		return s.read(q)
	case q.Update != nil:
		return s.update(q)
	case q.Delete != nil:
		return s.delete(q)
	}
	panic(q)
}

func (s *SQLStore) Close() error { return s.db.Close() }

func New(name, connect string, init bool) (*SQLStore, error) {
	db, err := sql.Open("sqlite3", connect)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if init {
		_, err = db.Exec(initStmt)
		if err != nil {
			return nil, err
		}
		_, err = db.Exec("INSERT INTO cluster(id) values(?)", name)
		if err != nil {
			return nil, err
		}
	}
	return &SQLStore{db}, nil
}
