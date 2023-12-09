package consul

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const (
	kvPath      string = "/v1/kv/%s"
	blockPath   string = "%s?index=%d&wait=%ds"
	limitFactor int    = 3
)

// Client specifies an http client.
type Client interface {
	SendObject(ctx context.Context, method, path string, snd, rcv any) (err error)
}

// Config is Consul configuration.
type Config struct {
	PollInterval time.Duration `json:"poll_interval" desc:"long polling duration" default:"1m"`
}

// Consul is a Consul client.
type Consul struct {
	Client       Client
	Limiter      *rate.Limiter
	PollInterval time.Duration
	idx          uint64
	// Todo: move Key here??
}

// New creates a Consul from Config.
func (cfg *Config) New(client Client) *Consul {

	rateLimit := rate.Every(cfg.PollInterval / time.Duration(limitFactor))

	return &Consul{
		Client:       client,
		Limiter:      rate.NewLimiter(rateLimit, limitFactor),
		PollInterval: cfg.PollInterval,
	}
}

// GetKv gets a key/value with a long poll if idx is not zero.
func (csl *Consul) GetKv(ctx context.Context, key string, idx uint64) (value []byte, latest uint64, err error) {

	path := fmt.Sprintf(kvPath, key)
	if idx != 0 {
		path = fmt.Sprintf(blockPath, path, idx, int(csl.PollInterval.Seconds()))
	}

	results := []kvResult{}
	err = csl.Client.SendObject(ctx, "GET", path, nil, &results)
	if err != nil {
		return
	}

	if len(results) != 1 {
		err = errors.Errorf("non-singular kv results for key: %s", key)
		return
	}

	value, err = base64.StdEncoding.DecodeString(results[0].Value)
	if err != nil {
		err = errors.Wrapf(err, "failed to decode value from: %#v", results)
		return
	}

	latest = results[0].ModifyIndex
	return
}

func (csl *Consul) Watch(ctx context.Context, key string) (data []byte, err error) {

	delay := csl.Limiter.Reserve().Delay()
	//fmt.Printf(">>> delay: %s\n\n", delay)
	time.Sleep(delay)

	var newIdx uint64
	data, newIdx, err = csl.GetKv(ctx, key, csl.idx)
	if err != nil {
		return
	}

	if newIdx < csl.idx {
		newIdx = 0
	}
	csl.idx = newIdx

	return
}

// unexported

type kvResult struct {
	CreateIndex uint64
	ModifyIndex uint64
	LockIndex   uint64
	Session     string
	Key         string
	Value       string
	Flags       uint64
}
