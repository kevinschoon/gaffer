package supervisor

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/vektorlab/gaffer/log"
	"github.com/vektorlab/gaffer/store"
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
	Store   store.Store
	Service string
}

func Run(opts Opts) error {

	fn := func() error {
		// Request cluster information
		self, svc, err := store.Register(opts.Store, opts.Service)

		if err != nil {
			return err
		}

		run := func() error {

			maybeLog(func() error {
				return svc.Start()
			})

			for {

				maybeLog(func() error {
					return store.Update(opts.Store, self, svc)
				})

				if !svc.Running() {
					return fmt.Errorf(svc.Cmd.ProcessState.String())
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
					opts.Service,
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
				opts.Service,
				zap.String("message", "supervisor process timed out"),
				zap.Duration("duration", d),
				zap.Error(err),
			)
		},
	)
}
