package supervisor

import (
	"fmt"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/runc"
	"github.com/mesanine/gaffer/store"
)

// Once launches on-boot services sequentially
// TODO: Add retry / backoff
func Once(cfg config.Config) error {
	services, err := store.NewFSStore(cfg).Services()
	if err != nil {
		return err
	}
	for _, svc := range services {
		ro, err := svc.ReadOnly()
		if err != nil {
			return err
		}
		log.Log.Info(fmt.Sprintf("starting on-boot service %s", svc.Id))
		code, err := runc.New(svc.Id, svc.Bundle, ro, cfg).Run()
		log.Log.Info(fmt.Sprintf("on-boot service %s exited with code %d", svc.Id, code))
		if code != 0 || err != nil {
			if err == nil {
				return fmt.Errorf("service %s returned a non-zero exit code %d", code)
			}
			return err
		}
	}
	return nil
}
