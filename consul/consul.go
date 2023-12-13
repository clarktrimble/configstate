// Package consul is a Consul Key-Value client.
package consul

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

//go:generate moq -pkg mock -out mock/mock.go . Client

// Todo: consider adding recursive keys

const (
	kvPath      string = "/v1/kv/%s"
	blockPath   string = "%s?index=%d&wait=%ds"
	limitFactor int    = 3
	limitBurst  int    = 3
)

// Client specifies an http client.
type Client interface {
	SendObject(ctx context.Context, method, path string, snd, rcv any) (err error)
}

// Config is Consul configuration.
type Config struct {
	PollInterval time.Duration `json:"poll_interval" desc:"long polling duration" default:"1m"`
	Key          string        `json:"key" desc:"key to be watched" required:"true"`
}

// Consul is a Consul client.
type Consul struct {
	Client       Client
	Limiter      *rate.Limiter
	LimitDelay   time.Duration
	PollInterval time.Duration
	Key          string
	Idx          uint64
}

// New creates a Consul from Config.
func (cfg *Config) New(client Client) *Consul {

	rateLimit := rate.Every(cfg.PollInterval / time.Duration(limitFactor))

	return &Consul{
		Client:       client,
		Limiter:      rate.NewLimiter(rateLimit, limitBurst),
		PollInterval: cfg.PollInterval,
		Key:          cfg.Key,
	}
}

// GetKv gets a key/value with a long poll if idx is not zero.
func (csl *Consul) GetKv(ctx context.Context, key string, idx uint64) (value []byte, latest uint64, err error) {

	path := fmt.Sprintf(kvPath, key)
	if idx != 0 {
		path = fmt.Sprintf(blockPath, path, idx, int(csl.PollInterval.Seconds()))
	}

	results := []KvResult{}
	err = csl.Client.SendObject(ctx, "GET", path, nil, &results)
	if err != nil {
		return
	}
	//fmt.Printf(">>> results: %#v\n", results)

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

// Poll long-polls a key.
//
// On the first poll (when Idx is 0) it returns right away.
// On subsequent polls (when Idx is not 0) it returns:
//   - on a change of the key's value in Consul
//   - or at the end of PollInterval, whichever comes first
//
// It will not return more frequently than:
//   - PollInvterval divided by limitFactor
//   - with allowed burst of limitBurst
func (csl *Consul) Poll(ctx context.Context) (data []byte, err error) {

	// https://developer.hashicorp.com/consul/api-docs/features/blocking

	delay := csl.Limiter.Reserve().Delay()
	csl.LimitDelay += delay
	time.Sleep(delay)

	var newIdx uint64
	data, newIdx, err = csl.GetKv(ctx, csl.Key, csl.Idx)
	if err != nil {
		return
	}

	if newIdx < csl.Idx {
		newIdx = 0
	}
	csl.Idx = newIdx

	return
}

// KvResult is exported for test, bah.
type KvResult struct {
	CreateIndex uint64
	ModifyIndex uint64
	LockIndex   uint64
	Session     string
	Key         string
	Value       string
	Flags       uint64
}
