package supervisor

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/client"
	"github.com/vektorlab/gaffer/cluster"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store/query"
	"go.uber.org/zap"
	"time"
)

const PollTime = 2000 * time.Millisecond

func maybeLog(fn func() error) error {
	err := fn()
	if err != nil {
		log.Log.Warn("supervisor", zap.Error(err))
	}
	return err
}

type Opts struct {
	Client    *client.Client
	ClusterID string
	Service   string
}

func Run(opts Opts) error {

	var (
		config *cluster.Cluster
		host   *cluster.Host
	)

	fn := func() error {
		// Request cluster information
		resp, err := opts.Client.Query(
			&query.Query{
				Type: query.READ,
				Read: &query.Read{
					ID: opts.ClusterID,
				},
			},
		)

		if err != nil {
			return err
		}

		config = resp.One()

		if config == nil {
			return fmt.Errorf("cannot find cluster")
		}

		for _, h := range config.Hosts {
			if err := h.Register(); err == nil {
				host = h
				break
			}
		}

		if host == nil {
			return fmt.Errorf("Could not register self with cluster")
		}

		service, ok := host.Services[opts.Service]
		if !ok {
			return fmt.Errorf("Invalid service %s", opts.Service)
		}

		run := func() error {

			maybeLog(func() error {
				return service.Start()
			})

			for {
				service.Update()
				host.Update()
				_, err := opts.Client.Query(&query.Query{
					Type: query.UPDATE,
					Update: &query.Update{
						Clusters: []*cluster.Cluster{
							config,
						},
					},
				})
				if err != nil {
					log.Log.Info("supervisor", zap.String("message", "failed to update remote server"), zap.Error(err))
				}
				if !service.Running() {
					return fmt.Errorf(service.Cmd.ProcessState.String())
				}
				time.Sleep(PollTime)
			}

		}
		exp := backoff.NewExponentialBackOff()
		exp.MaxElapsedTime = 30000 * time.Millisecond
		return backoff.RetryNotify(
			run,
			exp,
			func(err error, d time.Duration) {
				log.Log.Info(
					"supervisor",
					zap.String("message", fmt.Sprintf("service %s has failed", opts.Service)),
					zap.Duration("duration", d),
					zap.Error(err),
				)
			},
		)
	}

	return backoff.RetryNotify(
		fn,
		backoff.NewConstantBackOff(5000*time.Millisecond),
		func(err error, d time.Duration) {
			log.Log.Info(
				"supervisor",
				zap.String("message", "supervisor process timed out"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		},
	)
}
