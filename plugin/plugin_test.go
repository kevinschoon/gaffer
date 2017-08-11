package plugin

import (
	"github.com/mesanine/gaffer/event"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type MockPlugin struct {
	stop chan bool
}

func (mp MockPlugin) Name() string { return "gaffer.MockPlugin" }

func (mp MockPlugin) Run(*event.EventBus) error {
	<-mp.stop
	return nil
}

func (mp MockPlugin) Stop() error {
	mp.stop <- true
	return nil
}

func NewMockPlugin() MockPlugin {
	return MockPlugin{stop: make(chan bool, 1)}
}

func TestRegistry(t *testing.T) {
	reg := Registry{}
	assert.NoError(t, reg.Register(NewMockPlugin()))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, Run(reg))
	}()
	assert.NoError(t, reg["gaffer.MockPlugin"].Stop())
	wg.Wait()
}
