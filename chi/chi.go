// Package chi adapts the exuberant chi router to the much smaller interface used here.
package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Chi represents the wrapped chi router/mux instance.
type Chi struct {
	Mux *chi.Mux
}

// New creates a new Chi.
func New() *Chi {

	return &Chi{
		Mux: chi.NewMux(),
	}
}

// Set sets a route given http method, path/pattern and a handler.
func (chi *Chi) Set(method, path string, handler http.HandlerFunc) {

	chi.Mux.MethodFunc(method, path, handler)
}

// ServeHTTP handles http requests.
func (chi *Chi) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	chi.Mux.ServeHTTP(writer, request)
}
