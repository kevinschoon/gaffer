package store

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/vektorlab/gaffer/cluster"
	"time"
)

type Store struct {
	db *bolt.DB
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) GetService() (*cluster.Service, error) {
	var svc *cluster.Service
	err := s.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket([]byte("store")).Get([]byte("service"))
		if len(raw) == 0 {
			return nil
		}
		s := &cluster.Service{}
		err := json.Unmarshal(raw, s)
		svc = s
		return err
	})
	return svc, err
}

func (s *Store) SetService(svc *cluster.Service) error {
	raw, err := json.Marshal(svc)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("store")).Put([]byte("service"), raw)
	})
}

func New(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	return store, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("store"))
		return err
	})
}
