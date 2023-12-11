package discover

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/clarktrimble/delish"
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

// Capability represents something a service can do.
type Capability struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// Service is a service on the network.
type Service struct {
	Uri  string       `json:"uri"`
	Caps []Capability `json:"capabilities"`
}

// Discover polls for available services.
type Discover struct {
	Logger   Logger
	Poller   Poller
	services []Service
	mu       sync.RWMutex
}

// Services returns available services.
func (dsc *Discover) Services() (services []Service) {

	services = make([]Service, len(dsc.services))

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

		services := []Service{}
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

// Todo: prolly move this handler away -- delish dep is yuch?
// orrr: store bytes and skip hand-wrings seen in commented delish code below??
func (dsc *Discover) GetServices(writer http.ResponseWriter, request *http.Request) {

	rp := &delish.Respond{
		Writer: writer,
		Logger: dsc.Logger,
	}

	rp.WriteObjects(request.Context(), map[string]any{"services": dsc.Services()})
}

/*
func (rp *Respond) WriteObjects(ctx context.Context, objects map[string]any) {

	header(rp.Writer, 0)

	data, err := json.Marshal(objects)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", objects)
		rp.Logger.Error(ctx, "failed to encode response", err)

		rp.Writer.WriteHeader(http.StatusInternalServerError)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
	}

	rp.Write(ctx, data)
}
func (rp *Respond) Write(ctx context.Context, data []byte) {

	// leaving content-type as exercise for handler

	_, err := rp.Writer.Write(data)
	if err != nil {
		err = errors.Wrapf(err, "failed to write response")
		rp.Logger.Error(ctx, "failed to write response", err)
	}
}
func header(writer http.ResponseWriter, code int) {

	writer.Header().Set("content-type", "application/json")
	if code != 0 {
		writer.WriteHeader(code)
	}
}
*/
