package plugin

import (
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type MockPlugin struct {
	stop chan bool
}

func (mp MockPlugin) Name() string { return "gaffer.MockPlugin" }

func (mp *MockPlugin) Configure(config.Config) error {
	mp.stop = make(chan bool, 1)
	return nil
}

func (mp *MockPlugin) Run(*event.EventBus) error {
	<-mp.stop
	return nil
}

func (mp *MockPlugin) Stop() error {
	mp.stop <- true
	return nil
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	assert.NoError(t, reg.Register(&MockPlugin{}))
	assert.NoError(t, reg.Configure(config.Config{}))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, reg.Run())
	}()
	assert.NoError(t, reg.plugins["gaffer.MockPlugin"].Stop())
	wg.Wait()
}
