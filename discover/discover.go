package discover

import (
	"context"
	"fmt"
	"hash"
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
// Poll is expected to rate-limit itself in some reasonable way.
//
// Todo: consider moving interval and rate limit here.
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
	hash     hash.Hash
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

	dsc.hash = fnv.New64a()

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
		if dsc.unchanged(data) {
			continue
		}

		services, err := entity.DecodeServices(data)
		if err != nil {
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

func (dsc *Discover) unchanged(data []byte) bool {

	dsc.hash.Write(data)
	newSum := fmt.Sprintf("%x", dsc.hash.Sum(nil))
	dsc.hash.Reset()

	if dsc.sum == newSum {
		return true
	}

	dsc.sum = newSum
	return false
}

func (dsc *Discover) getServices(writer http.ResponseWriter, request *http.Request) {

	rp := &respond.Respond{
		Writer: writer,
		Logger: dsc.Logger,
	}

	rp.WriteObjects(request.Context(), map[string]any{"services": dsc.Services()})
}
