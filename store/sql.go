package store

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store/query"
	"github.com/vektorlab/gaffer/user"
	"go.uber.org/zap"
)

const initStmt = `
CREATE TABLE clusters (id STRING NOT NULL PRIMARY KEY, user STRING, data STRING);
CREATE TABLE users (id STRING NOT NULL PRIMARY KEY, token STRING NOT NULL);
INSERT INTO users(id, token) values("root", "root");
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
	for _, cluster := range q.Create.Clusters {
		raw, err := json.Marshal(cluster)
		if err != nil {
			return nil, err
		}
		stmt, err := tx.Prepare("INSERT INTO clusters(id, user, data) values(?, ?, ?)")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(cluster.ID, q.User.ID, string(raw))
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &query.Response{Clusters: q.Create.Clusters}, nil
}

func (s *SQLStore) read(q *query.Query) (*query.Response, error) {
	clusters := []*cluster.Cluster{}
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
		cluster := &cluster.Cluster{}
		err = json.Unmarshal([]byte(data), cluster)
		if err != nil {
			return nil, err
		}
		cluster.ID = id
		clusters = append(clusters, cluster)
	}
	return &query.Response{Clusters: clusters}, nil
}

func (s *SQLStore) readUser(q *query.Query) (*query.Response, error) {
	var usr *user.User
	rows, err := s.db.Query("SELECT id, token FROM users WHERE id == ? AND token == ? LIMIT 1", q.User.ID, q.User.Token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id    string
			token string
		)
		err := rows.Scan(&id, &token)
		if err != nil {
			return nil, err
		}
		usr = &user.User{id, token}
	}
	if usr == nil {
		return nil, ErrUserNotFound{q.User.ID}
	}
	return &query.Response{User: usr}, nil
}

func (s *SQLStore) update(q *query.Query) (*query.Response, error) {
	var (
		err error
		tx  *sql.Tx
	)
	defer maybeRollback(tx, err)
	tx, err = s.db.Begin()
	if err != nil {
		return nil, err
	}
	for _, cluster := range q.Update.Clusters {
		raw, err := json.Marshal(cluster)
		if err != nil {
			return nil, err
		}
		stmt, err := tx.Prepare("UPDATE clusters SET data = ? WHERE id == ? AND user == ?")
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(string(raw), cluster.ID, q.User.ID)
		if err != nil {
			return nil, err
		}
		err = stmt.Close()
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &query.Response{}, nil
}

func (s *SQLStore) delete(q *query.Query) (*query.Response, error) {
	var (
		err error
		tx  *sql.Tx
	)
	defer maybeRollback(tx, err)
	tx, err = s.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("DELETE FROM clusters WHERE id == ? AND user == ?")
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(q.Delete.ID, q.User.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &query.Response{}, nil
}

func (s *SQLStore) Query(q *query.Query) (*query.Response, error) {
	switch q.Type {
	case query.CREATE:
		return s.create(q)
	case query.READ:
		return s.read(q)
	case query.READ_USER:
		return s.readUser(q)
	case query.UPDATE:
		return s.update(q)
	case query.DELETE:
		return s.delete(q)
	}
	panic(q.Type)
}

func (s *SQLStore) Close() error { return s.db.Close() }

func NewSQLStore(connect string, init bool) (Store, error) {
	db, err := sql.Open("sqlite3", connect)
	if err != nil {
		return nil, err
	}
	if init {
		_, err = db.Exec(initStmt)
		if err != nil {
			return nil, err
		}
	}
	return &SQLStore{db}, nil
}
