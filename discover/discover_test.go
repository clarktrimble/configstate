package discover_test

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "configstate/discover"
	"configstate/discover/mock"
	"configstate/entity"
)

func TestDiscover(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Discover Suite")
}

var _ = Describe("Discover", func() {

	var (
		ctx context.Context
		wg  sync.WaitGroup
		dsc *Discover
	)

	BeforeEach(func() {

		ctx = context.Background()

		mockedPoller := &mock.PollerMock{
			PollFunc: func(ctx context.Context) ([]byte, error) {
				return []byte(data), nil
			},
		}

		mockedLogger := &mock.LoggerMock{
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
			InfoFunc:  func(ctx context.Context, msg string, kv ...any) {},
			WithFieldsFunc: func(ctx context.Context, kv ...interface{}) context.Context {
				return ctx
			},
		}

		dsc = &Discover{
			Logger: mockedLogger,
			Poller: mockedPoller,
		}
	})

	Describe("getting a key-value", func() {
		var (
			services []entity.Service
		)

		JustBeforeEach(func() {
			services = dsc.Services()
		})

		When("all is well", func() {
			BeforeEach(func() {
				// start discovery and give it a chance to call poller
				dsc.Start(ctx, &wg)
				time.Sleep(time.Millisecond)
				// Todo: test worker properly pretty pls
				// Todo: unstarted!
			})

			It("responds with decoded data and index", func() {
				Expect(services).To(Equal([]entity.Service{
					{
						Uri: "http://pool04.boxworld.org/api/v2",
						Caps: []entity.Capability{
							{Name: "resize", Capacity: 23},
						},
					},
					{
						Uri: "http://pool24.boxworld.org/api/v2",
						Caps: []entity.Capability{
							{Name: "resize", Capacity: 5},
						},
					},
				}))
			})
		})

	})

})

const (
	data string = `
[
  {
    "uri": "http://pool04.boxworld.org/api/v2",
    "capabilities": [
      {
        "name": "resize",
        "capacity": 23
      }
    ]
  },
  {
    "uri": "http://pool24.boxworld.org/api/v2",
    "capabilities": [
      {
        "name": "resize",
        "capacity": 5
      }
    ]
  }
]`
)
