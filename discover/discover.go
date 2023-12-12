package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"

	"github.com/clarktrimble/hondo"
	"github.com/pkg/errors"

	"configstate/entity"
	"configstate/respond"
)

//go:generate moq -pkg mock -out mock/mock.go . Logger Poller

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

// Router specifies a router.
type Router interface {
	Set(method, path string, handler http.HandlerFunc)
}

// Discover polls for available services.
type Discover struct {
	Logger   Logger
	Poller   Poller
	services []entity.Service
	mu       sync.RWMutex
	sum      string
}

// Services returns a copy of available services.
func (dsc *Discover) Services() (services []entity.Service) {

	services = make([]entity.Service, len(dsc.services))

	dsc.mu.RLock()
	copy(services, dsc.services)
	dsc.mu.RUnlock()

	return services
}

// Start starts the poll worker.
func (dsc *Discover) Start(ctx context.Context, wg *sync.WaitGroup) {

	ctx = dsc.Logger.WithFields(ctx, "worker_id", hondo.Rand(7))
	dsc.Logger.Info(ctx, "worker starting", "name", "discovery")

	wg.Add(1)
	go dsc.work(ctx, wg)
}

// Register registers routes with the router.
func (dsc *Discover) Register(rtr Router) {

	rtr.Set("GET", "/services", dsc.getServices)
}

// unexported

func (dsc *Discover) work(ctx context.Context, wg *sync.WaitGroup) {

	hsh := fnv.New64a()

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

		hsh.Write(data)
		newSum := fmt.Sprintf("%x", hsh.Sum(nil))
		hsh.Reset()

		if dsc.sum == newSum {
			continue
		}
		dsc.sum = newSum

		services := []entity.Service{}
		err = json.Unmarshal(data, &services)
		if err != nil {
			err = errors.Wrapf(err, "failed to unmarshal services given: %s", data)
			dsc.Logger.Error(ctx, "failed to watch", err)
			continue
		}

		dsc.Logger.Info(ctx, "updating services")

		dsc.mu.Lock()
		dsc.services = services
		dsc.mu.Unlock()
	}

	wg.Done()
	dsc.Logger.Info(ctx, "worker stopped")
}

func (dsc *Discover) getServices(writer http.ResponseWriter, request *http.Request) {

	rp := &respond.Respond{
		Writer: writer,
		Logger: dsc.Logger,
	}

	rp.WriteObjects(request.Context(), map[string]any{"services": dsc.Services()})
}
