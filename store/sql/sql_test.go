package sql

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/cluster/service"
	"github.com/vektorlab/gaffer/store/query"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

const CLEAN = false

func tempDB(t *testing.T) (*SQLStore, func()) {
	path, _ := ioutil.TempDir("/tmp", "gaffer-test-")
	db, err := New("mock-cluster", fmt.Sprintf("%s/gaffer.db", path), true)
	assert.NoError(t, err)
	return db, func() {
		fmt.Printf("rm %s ", path)
		if CLEAN {
			fmt.Println(os.RemoveAll(path))
		}
	}
}

func stdout(i interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	enc.Encode(i)
}

func TestQuery(t *testing.T) {
	db, cleanup := tempDB(t)
	defer cleanup()
	config := cluster.New("", 5)
	for i := 0; i < 5; i++ {
		config.Services[config.Hosts[i].ID] = []*service.Service{
			&service.Service{ID: "mesos-master", Args: []string{"mesos-master"}},
			&service.Service{ID: "mesos-agent", Args: []string{"mesos-agent"}},
			&service.Service{ID: "zookeeper", Args: []string{"zookeeper"}},
		}
	}
	resp, err := db.Query(&query.Query{
		Create: &query.Create{
			Cluster: config,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Create.Cluster.Hosts, 5)
	assert.Len(t, resp.Create.Cluster.Services, 5)
	stdout(resp)
	resp, err = db.Query(
		&query.Query{Read: &query.Read{}},
	)
	assert.NoError(t, err)
	assert.Len(t, resp.Read.Cluster.Hosts, 5)
	assert.Len(t, resp.Read.Cluster.Services, 5)
	for _, services := range resp.Read.Cluster.Services {
		assert.Len(t, services, 3)
	}
	stdout(resp)
	host := resp.Read.Cluster.Hosts[0]
	svc := resp.Read.Cluster.Services[host.ID][0]
	host.Registered = true
	resp, err = db.Query(&query.Query{UpdateHost: &query.UpdateHost{host}})
	assert.NoError(t, err)
	assert.True(t, resp.UpdateHost.Host.Registered)
	svc.Environment = []*service.Env{&service.Env{"fuu", "bar"}}
	resp, err = db.Query(&query.Query{UpdateService: &query.UpdateService{HostID: host.ID, Service: svc}})
	assert.NoError(t, err)
	assert.Equal(t, resp.UpdateService.Service.Environment[0].Name, "fuu")
	_, err = db.Query(&query.Query{Delete: &query.Delete{HostID: host.ID, ServiceID: svc.ID}})
	assert.NoError(t, err)
	_, err = db.Query(&query.Query{Delete: &query.Delete{HostID: host.ID}})
	assert.NoError(t, err)
}

func allRegistered(t *testing.T, db *SQLStore) bool {
	resp, err := db.Query(&query.Query{Read: &query.Read{}})
	assert.NoError(t, err)
	for _, host := range resp.Read.Cluster.Hosts {
		if !host.Registered {
			return false
		}
	}
	return true
}

func maybeRegister(t *testing.T, db *SQLStore) {
	resp, err := db.Query(&query.Query{Read: &query.Read{}})
	assert.NoError(t, err)
	host := resp.Read.Cluster.Hosts[rand.Intn(len(resp.Read.Cluster.Hosts))]
	if rand.Intn(100) > 90 {
		host.Registered = true
		host.Hostname = "some-host"
		host.LastContacted = time.Now()
		host.LastRegistered = time.Now()
		_, err := db.Query(&query.Query{UpdateHost: &query.UpdateHost{Host: host}})
		assert.NoError(t, err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	db, cleanup := tempDB(t)
	defer cleanup()
	config := cluster.New("", 5)
	for i := 0; i < 5; i++ {
		config.Services[config.Hosts[i].ID] = []*service.Service{
			&service.Service{ID: "mesos-master", Args: []string{"mesos-master"}},
			&service.Service{ID: "mesos-agent", Args: []string{"mesos-agent"}},
			&service.Service{ID: "zookeeper", Args: []string{"zookeeper"}},
		}
	}
	resp, err := db.Query(&query.Query{
		Create: &query.Create{
			Cluster: config,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Create.Cluster.Hosts, 5)
	assert.Len(t, resp.Create.Cluster.Services, 5)
	stdout(resp)
	for !allRegistered(t, db) {
		for i := 0; i < 20; i++ {
			go maybeRegister(t, db)
		}
	}
}

func init() {
	rand.Seed(time.Now().Unix())
}
