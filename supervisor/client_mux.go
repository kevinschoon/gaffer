package supervisor

import (
	"fmt"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"sync"
)

func Clients(hosts cluster.Hosts) map[cluster.Host]Client {
	clients := map[cluster.Host]Client{}
	for _, host := range hosts {
		clients[*host] = Client{Hostname: host.Name, Port: host.Port}
	}
	return clients
}

type Response struct {
	Host    cluster.Host
	Error   error
	Status  *StatusResponse
	Update  *UpdateResponse
	Restart *RestartResponse
}

type ClientMux struct {
	Clients map[cluster.Host]Client
}

func (cm ClientMux) Status() chan Response {
	respCh := make(chan Response)
	var wg sync.WaitGroup
	for host, cli := range cm.Clients {
		wg.Add(1)
		go func(host cluster.Host, cli Client) {
			defer wg.Done()
			resp, err := cli.Status()
			if err != nil {
				log.Log.Debug(fmt.Sprintf("could not refresh service from %s:%d", host.Name, host.Port))
				respCh <- Response{Host: host, Error: err}
			} else {
				respCh <- Response{Host: host, Status: resp}
			}
		}(host, cli)
	}
	go func() {
		wg.Wait()
		close(respCh)
	}()
	return respCh
}

func (cm ClientMux) Apply(service *cluster.Service) chan Response {
	respCh := make(chan Response)
	var wg sync.WaitGroup
	for host, cli := range cm.Clients {
		wg.Add(1)
		go func(cli Client) {
			defer wg.Done()
			resp, err := cli.Update(service)
			if err != nil {
				log.Log.Info(fmt.Sprintf("could not apply service to %s:%d", cli.Hostname, cli.Port))
				respCh <- Response{Host: host, Error: err}
			} else {
				respCh <- Response{Host: host, Update: resp}
			}
		}(cli)
	}
	go func() {
		wg.Wait()
		close(respCh)
	}()
	return respCh
}

func (cm ClientMux) Restart() chan Response {
	respCh := make(chan Response)
	var wg sync.WaitGroup
	for host, cli := range cm.Clients {
		wg.Add(1)
		go func(cli Client) {
			defer wg.Done()
			resp, err := cli.Restart()
			if err != nil {
				log.Log.Debug(fmt.Sprintf("could not restart service @ %s:%d", cli.Hostname, cli.Port))
				respCh <- Response{Host: host, Error: err}
			} else {
				respCh <- Response{Host: host, Restart: resp}
			}
		}(cli)
	}
	go func() {
		wg.Wait()
		close(respCh)
	}()
	return respCh
}
