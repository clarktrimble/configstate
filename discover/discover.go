package discover

import (
	"configstate/entity"
	"context"
	"encoding/json"
	"sync"

	"github.com/clarktrimble/hondo"
	"github.com/pkg/errors"
)

// Logger specifies a logger.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...interface{}) context.Context
}

// Poller specifies a poller, returning data every so often that might be updated.
type Poller interface {
	Poll(ctx context.Context) (data []byte, err error)
}

// Discover polls for available services.
type Discover struct {
	Logger   Logger
	Poller   Poller
	services []entity.Service
	mu       sync.RWMutex
}

// Services returns available services.
func (dsc *Discover) Services() (services []entity.Service) {

	services = make([]entity.Service, len(dsc.services))

	dsc.mu.RLock()
	copy(services, dsc.services)
	dsc.mu.RUnlock()

	return services
}

// Start starts the polling worker.
func (dsc *Discover) Start(ctx context.Context, wg *sync.WaitGroup) {

	ctx = dsc.Logger.WithFields(ctx, "worker_id", hondo.Rand(7))
	dsc.Logger.Info(ctx, "worker starting", "name", "discovery")

	wg.Add(1)
	go dsc.work(ctx, wg)
}

// unexported

func (dsc *Discover) work(ctx context.Context, wg *sync.WaitGroup) {

	for {

		data, err := dsc.Poller.Poll(ctx)
		if errors.Is(err, context.Canceled) {
			dsc.Logger.Info(ctx, "worker shutting down")
			break
		}
		if err != nil {
			dsc.Logger.Error(ctx, "failed to watch", err)
			continue
		}

		services := []entity.Service{}
		err = json.Unmarshal(data, &services)
		if err != nil {
			err = errors.Wrapf(err, "failed to unmarshal services given: %s", data)
			dsc.Logger.Error(ctx, "failed to watch", err)
			continue
		}

		// Todo: hash or sommat?
		dsc.Logger.Info(ctx, "updating services")

		// Todo: mutex hand wring
		dsc.mu.Lock()
		dsc.services = services
		dsc.mu.Unlock()
	}

	wg.Done()
	dsc.Logger.Info(ctx, "worker stopped")
}
