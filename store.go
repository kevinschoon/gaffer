package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

const (
	CREATE    QueryType = "CREATE"
	READ      QueryType = "READ"
	UPDATE    QueryType = "UPDATE"
	DELETE    QueryType = "DELETE"
	READ_USER QueryType = "READ_USER"
)

type QueryType string

type Query struct {
	Type    QueryType `json:"type"`
	Cluster *Cluster  `json:"cluster"`
	User    *User     `json:"user"`
}

type Response struct {
	Clusters []*Cluster `json:"clusters"`
	User     *User      `json:"user"`
}

type Store interface {
	Query(*Query) (*Response, error)
	Close() error
}

type SQLStore struct {
	db  *sql.DB
	log *zap.Logger
}

func (s *SQLStore) Query(q *Query) (*Response, error) {
	response := &Response{
		Clusters: []*Cluster{},
	}
	switch q.Type {
	case CREATE:
		raw, err := json.Marshal(q.Cluster)
		if err != nil {
			return nil, err
		}
		tx, err := s.db.Begin()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		stmt, err := tx.Prepare("INSERT INTO clusters(id, user, data) values(?, ?, ?)")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(q.Cluster.ID, q.User.ID, string(raw))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
	case READ:
		rows, err := s.db.Query("SELECT id, data FROM clusters WHERE user == ?", q.User.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var (
				id   string
				data string
			)
			err := rows.Scan(&id, &data)
			if err != nil {
				return nil, err
			}
			cluster := &Cluster{}
			err = json.Unmarshal([]byte(data), cluster)
			if err != nil {
				return nil, err
			}
			cluster.ID = id
			response.Clusters = append(response.Clusters, cluster)
		}
	case UPDATE:
		raw, err := json.Marshal(q.Cluster)
		if err != nil {
			return nil, err
		}
		tx, err := s.db.Begin()
		if err != nil {
			return nil, err
		}
		stmt, err := tx.Prepare("UPDATE clusters SET data = ? WHERE id == ? AND user == ?")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(string(raw), q.Cluster.ID, q.User.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
	case DELETE:
		tx, err := s.db.Begin()
		if err != nil {
			return nil, err
		}
		stmt, err := tx.Prepare("DELETE FROM clusters WHERE id == ? AND user == ?")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(q.Cluster.ID, q.User.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
	case READ_USER:
		rows, err := s.db.Query("SELECT id, token FROM users WHERE token == ? LIMIT 1", q.User.Token)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var (
				id    int
				token string
			)
			err := rows.Scan(&id, &token)
			if err != nil {
				return nil, err
			}
			response.User = &User{id, token}
		}
	}
	return response, nil
}

func (s *SQLStore) Close() error { return s.db.Close() }

func NewSQLStore(connect string, logger *zap.Logger) (Store, error) {
	db, err := sql.Open("sqlite3", connect)
	if err != nil {
		return nil, err
	}
	return &SQLStore{db, logger}, nil
}
