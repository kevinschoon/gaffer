package server

/*
import (
	"encoding/json"
	"github.com/containerd/go-runc"
)

func (s *StatusResponse) UnmarshalStats() (map[string]*runc.Stats, error) {
	statsMap := map[string]*runc.Stats{}
	for id, stats := range s.Stats {
		statsMap[id] = &runc.Stats{}
		err := json.Unmarshal(stats.Value, statsMap[id])
		if err != nil {
			return nil, err
		}
	}
	return statsMap, nil
}
*/
