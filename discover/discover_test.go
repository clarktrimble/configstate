package discover_test

import (
	"context"
	"fmt"
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
				err := ctx.Err()
				if err != nil {
					return nil, err
				}
				time.Sleep(time.Microsecond)
				return []byte(data), nil
			},
		}

		mockedLogger := &mock.LoggerMock{
			ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {
				fmt.Printf(">>> error : %s\n", msg)
			},
			InfoFunc: func(ctx context.Context, msg string, kv ...any) {
				fmt.Printf(">>> info  : %s\n", msg)
			},
			WithFieldsFunc: func(ctx context.Context, kv ...interface{}) context.Context {
				fmt.Printf(">>> fields: %v\n", kv)
				return ctx
			},
		}

		dsc = &Discover{
			Logger: mockedLogger,
			Poller: mockedPoller,
		}
	})

	Describe("testing asynchronously", func() {
		var (
			//services []entity.Service
			cancel context.CancelFunc
		)

		JustBeforeEach(func() {
			//services = dsc.Services()
		})

		When("all is well", func() {
			BeforeEach(func() {
				// start discovery and give it a chance to call poller
				ctx, cancel = context.WithCancel(ctx)
				dsc.Start(ctx, &wg)
				//time.Sleep(time.Millisecond)
				// Todo: test worker properly pretty pls
				// Todo: unstarted!
			})

			FIt("responds with decoded data and index", func() {
				//Expect(services).To(Equal(expected))
				Eventually(dsc.Services).Should(Equal(expected))
				Eventually(dsc.Services).Should(Equal(expected))
				cancel()

				time.Sleep(time.Millisecond)
				// Todo: wg instead!!

			})
		})
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
				Expect(services).To(Equal(expected))
				/*
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
				*/
			})
		})

	})

})

var expected = []entity.Service{
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
}

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
