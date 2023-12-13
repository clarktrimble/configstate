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

	Describe("getting discovered services", func() {

		var (
			ctx    context.Context
			cancel context.CancelFunc
			wg     sync.WaitGroup

			mockedLogger *mock.LoggerMock
			dsc          *Discover

			data     string
			mu       sync.RWMutex
			dataSpec string
			expected []entity.Service
		)

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())

			mockedLogger = &mock.LoggerMock{
				ErrorFunc: func(ctx context.Context, msg string, err error, kv ...any) {},
				InfoFunc:  func(ctx context.Context, msg string, kv ...any) {},
				WithFieldsFunc: func(ctx context.Context, kv ...interface{}) context.Context {
					return ctx
				},
			}

			dsc = &Discover{
				Logger: mockedLogger,
				Poller: &mock.PollerMock{
					PollFunc: func(ctx context.Context) ([]byte, error) {
						err := ctx.Err()
						if err != nil {
							return nil, err
						}
						time.Sleep(time.Microsecond)

						mu.RLock()
						defer mu.RUnlock()
						return []byte(data), nil
					},
				},
			}

			dataSpec = `[
				{"uri":"http://pool04.boxworld.org/api/v2","capabilities":[{"name":"resize","capacity":23}]},
				{"uri":"http://pool24.boxworld.org/api/v2","capabilities":[{"name":"resize","capacity":%d}]}
			]`
			expected = []entity.Service{
				{
					Uri:  "http://pool04.boxworld.org/api/v2",
					Caps: []entity.Capability{{Name: "resize", Capacity: 23}},
				},
				{
					Uri:  "http://pool24.boxworld.org/api/v2",
					Caps: []entity.Capability{{Name: "resize", Capacity: 5}},
				},
			}
			data = fmt.Sprintf(dataSpec, 5)
		})

		When("worker is running", func() {
			BeforeEach(func() {
				dsc.Start(ctx, &wg)
			})

			It("reports services, detects change, and stops when ctx is cancelled; logging all the way", func(ctx SpecContext) {

				Eventually(dsc.Services).Should(Equal(expected))

				mu.Lock()
				data = fmt.Sprintf(dataSpec, 55)
				mu.Unlock()
				expected[1].Caps[0].Capacity = 55

				Eventually(dsc.Services).Should(Equal(expected))

				cancel()
				wg.Wait()

				Expect(mockedLogger.WithFieldsCalls()).To(HaveLen(1))
				Expect(mockedLogger.WithFieldsCalls()[0].Kv[0]).To(Equal("worker_id"))

				infoCalls := mockedLogger.InfoCalls()
				Expect(infoCalls).To(HaveLen(5))
				Expect(infoCalls[0].Msg).To(Equal("worker starting"))
				Expect(infoCalls[1].Msg).To(Equal("updating services"))
				Expect(infoCalls[2].Msg).To(Equal("updating services"))
				Expect(infoCalls[3].Msg).To(Equal("worker shutting down"))
				Expect(infoCalls[4].Msg).To(Equal("worker stopped"))

				Expect(mockedLogger.ErrorCalls()).To(HaveLen(0))

			}, SpecTimeout(time.Second))
		})

		When("worker has not started", func() {
			It("returns empty services", func() {
				Expect(dsc.Services()).To(Equal([]entity.Service{}))
			})
		})

	})

})
