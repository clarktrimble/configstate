package svc

import (
	"configstate/entity"
	"context"
	"net/http"

	"github.com/clarktrimble/delish"
)

// Logger specifies a logger.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...interface{}) context.Context
}

// Discoverer specifies a discovery interface.
type Dicsoverer interface {
	Services() (services []entity.Service)
}

// Svc is a service-layer.
type Svc struct {
	Logger     Logger
	Discoverer Dicsoverer
}

// GetServices is a handler returning currently known services.
func (svc *Svc) GetServices(writer http.ResponseWriter, request *http.Request) {

	rp := &delish.Respond{
		Writer: writer,
		Logger: svc.Logger,
	}

	rp.WriteObjects(request.Context(), map[string]any{"services": svc.Discoverer.Services()})
}
