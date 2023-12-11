package respond

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// copied out of delish because I did not want to depend on it
// feels like Respond has been awkward all along
// b-but code is fiddly and reasonably well tested
// Todo: maybe pull Respond into it's own mini-mod? (and stop wringing hands over it)

// Logger specifies a logger.
type Logger interface {
	Error(ctx context.Context, msg string, err error, kv ...any)
}

// Respond provides convinience methods when responding to a request
type Respond struct {
	Writer http.ResponseWriter
	Logger Logger
}

// Ok responds with 200
func (rp *Respond) Ok(ctx context.Context) {

	header(rp.Writer, 200)
	rp.Write(ctx, []byte(`{"status":"ok"}`))
}

// NotOk logs an error and responds with it
func (rp *Respond) NotOk(ctx context.Context, code int, err error) {

	header(rp.Writer, code)

	rp.Logger.Error(ctx, "returning error to client", err)
	rp.WriteObjects(ctx, map[string]any{"error": fmt.Sprintf("%v", err)})
}

// NotFound responds with 404
func (rp *Respond) NotFound(ctx context.Context) {

	header(rp.Writer, 404)
	rp.Write(ctx, []byte(`{"not":"found"}`))
}

// WriteObjects responds with marshalled objects by key
func (rp *Respond) WriteObjects(ctx context.Context, objects map[string]any) {

	header(rp.Writer, 0)

	data, err := json.Marshal(objects)
	if err != nil {
		err = errors.Wrapf(err, "somehow failed to encode: %#v", objects)
		rp.Logger.Error(ctx, "failed to encode response", err)

		header(rp.Writer, 500)
		rp.Write(ctx, []byte(`{"error": "failed to encode response"}`))
		return
	}

	rp.Write(ctx, data)
}

// Write respondes with arbitrary data, logging if error
func (rp *Respond) Write(ctx context.Context, data []byte) {

	// leaving content-type as exercise for handler

	_, err := rp.Writer.Write(data)
	if err != nil {
		err = errors.Wrapf(err, "failed to write response")
		rp.Logger.Error(ctx, "failed to write response", err)
	}
}

// unexported

func header(writer http.ResponseWriter, code int) {

	// Todo: yeah let rp receive here and stop passing writer arg

	writer.Header().Set("content-type", "application/json")
	if code != 0 {
		writer.WriteHeader(code)
	}
}
