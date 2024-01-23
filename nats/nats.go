package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const (
	limitInterval time.Duration = 15 * time.Second
	limitBurst    int           = 3
)

// Config is Nats configuration.
type Config struct {
	Url    string `json:"url" desc:"nats server url" required:"true"`
	Bucket string `json:"bucket" desc:"bucket be watched" required:"true"`
	Key    string `json:"key" desc:"key to be watched" required:"true"`
}

// Nats is a watch oriented representation of a nats server.
type Nats struct {
	Limiter    *rate.Limiter
	LimitDelay time.Duration
	updates    <-chan nats.KeyValueEntry
}

// New creates a Nats from Config.
func (cfg *Config) New() (nt *Nats, err error) {

	updates, err := updateChannel(cfg.Url, cfg.Bucket, cfg.Key)
	if err != nil {
		return
	}

	nt = &Nats{
		Limiter: rate.NewLimiter(rate.Every(limitInterval), limitBurst),
		updates: updates,
	}

	return
}

// Poll sings and dances its way to satisfying the Poller interface.
func (nt *Nats) Poll(ctx context.Context) ([]byte, error) {

	// rate limit just in case
	// even though we expect update channel to mostly block

	delay := nt.Limiter.Reserve().Delay()
	nt.LimitDelay += delay
	time.Sleep(delay)

	// return data found or signal shutdown

	select {
	case kve := <-nt.updates:
		if kve == nil {
			// seeing nil after first value from channel then it settles down
			return nil, errors.Errorf("got nil from kv watcher channel")
		}
		return kve.Value(), nil
	case <-ctx.Done():
		// convert to Canceled as that's how Poller rolls
		return nil, context.Canceled
	}
}

// updateChannel connects, ... , and eventually finds us an update channel.
//
// Of course, this code is not all that testable wo a nats server running
// b-but can unit Poll by sneaking in our own "updates" channel. Todo!
//
// In real life, some of these steps may have already been taken, with battle-hardened opts, etc.
func updateChannel(url, bucket, key string) (updates <-chan nats.KeyValueEntry, err error) {

	nc, err := nats.Connect(url)
	if err != nil {
		err = errors.Wrap(err, "failed to connect to nats")
		return
	}

	js, err := nc.JetStream()
	if err != nil {
		err = errors.Wrap(err, "failed to get jetstream context")
		return
	}

	kv, err := js.KeyValue(bucket)
	if err != nil {
		err = errors.Wrap(err, "failed to get kv store")
		return
	}

	watcher, err := kv.Watch(key)
	if err != nil {
		err = errors.Wrap(err, "failed to get kv watcher")
		return
	}

	updates = watcher.Updates()
	return
}
