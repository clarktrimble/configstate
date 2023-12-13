package consul_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "configstate/consul"
	"configstate/consul/mock"
)

func TestConsul(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul Suite")
}

var _ = Describe("Consul", func() {

	var (
		cfg     *Config
		client  *mock.ClientMock
		csl     *Consul
		ctx     context.Context
		data    []byte
		err     error
		encoded string
		decoded string
	)

	BeforeEach(func() {
		cfg = &Config{
			PollInterval: time.Minute,
		}
		client = &mock.ClientMock{}
		client.SendObjectFunc = func(ctx context.Context, method, path string, snd any, rcv any) error {
			result := rcv.(*[]KvResult)
			*result = append(*result, KvResult{
				ModifyIndex: 48,
				Value:       encoded,
			})
			return nil
		}
		csl = cfg.New(client)
		ctx = context.Background()
		encoded = "WyAgeyAgICAidXJpIjogImh0dHA6Ly9wb29sMDQuYm94d29ybGQub3JnL2FwaS92MiIsICAgICJjYXBhYmlsaXRpZXMiOiBbICAgICAgeyAgICAgICAgIm5hbWUiOiAicmVzaXplIiwgICAgICAgICJjYXBhY2l0eSI6IDIzICAgICAgfSAgICBdICB9LCAgeyAgICAidXJpIjogImh0dHA6Ly9wb29sMjQuYm94d29ybGQub3JnL2FwaS92MiIsICAgICJjYXBhYmlsaXRpZXMiOiBbICAgICAgeyAgICAgICAgIm5hbWUiOiAicmVzaXplIiwgICAgICAgICJjYXBhY2l0eSI6IDUgICAgICB9ICAgIF0gIH1d"
		decoded = `[  {    "uri": "http://pool04.boxworld.org/api/v2",    "capabilities": [      {        "name": "resize",        "capacity": 23      }    ]  },  {    "uri": "http://pool24.boxworld.org/api/v2",    "capabilities": [      {        "name": "resize",        "capacity": 5      }    ]  }]`
	})

	Describe("getting a key-value", func() {
		var (
			key    string
			idx    uint64
			newIdx uint64
		)

		JustBeforeEach(func() {
			data, newIdx, err = csl.GetKv(ctx, key, idx)
		})

		When("all is well", func() {
			BeforeEach(func() {
				key = "sample_key"
				idx = 5
			})

			It("responds with decoded data and index", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(decoded))
				Expect(newIdx).To(BeEquivalentTo(48))

				Expect(client.SendObjectCalls()).To(HaveLen(1))
				call := client.SendObjectCalls()[0]
				Expect(call.Ctx).To(Equal(ctx))
				Expect(call.Method).To(Equal("GET"))
				Expect(call.Path).To(Equal("/v1/kv/sample_key?index=5&wait=60s"))
				Expect(call.Snd).To(BeNil())
			})
		})

		When("index is 0", func() {
			BeforeEach(func() {
				key = "sample_key"
				idx = 0
			})

			It("does not long poll", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(decoded))
				Expect(newIdx).To(BeEquivalentTo(48))

				Expect(client.SendObjectCalls()).To(HaveLen(1))
				call := client.SendObjectCalls()[0]
				Expect(call.Ctx).To(Equal(ctx))
				Expect(call.Method).To(Equal("GET"))
				Expect(call.Path).To(Equal("/v1/kv/sample_key"))
				Expect(call.Snd).To(BeNil())
			})
		})
	})

	Describe("polling a key-value", func() {

		// Todo: check rate-limit

		JustBeforeEach(func() {
			data, err = csl.Poll(ctx)
		})

		When("all is well", func() {
			BeforeEach(func() {
				csl.Key = "sample_key"
				csl.Idx = 5
			})

			It("responds with decoded data and sets index", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(decoded))
				Expect(csl.Idx).To(BeEquivalentTo(48))

				Expect(client.SendObjectCalls()).To(HaveLen(1))
				call := client.SendObjectCalls()[0]
				Expect(call.Ctx).To(Equal(ctx))
				Expect(call.Method).To(Equal("GET"))
				Expect(call.Path).To(Equal("/v1/kv/sample_key?index=5&wait=60s"))
				Expect(call.Snd).To(BeNil())
			})
		})

		When("index backtracks", func() {
			BeforeEach(func() {
				csl.Key = "sample_key"
				csl.Idx = 55
			})

			It("responds with decoded data and resets index", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(decoded))
				Expect(csl.Idx).To(BeEquivalentTo(0))

				Expect(client.SendObjectCalls()).To(HaveLen(1))
				call := client.SendObjectCalls()[0]
				Expect(call.Ctx).To(Equal(ctx))
				Expect(call.Method).To(Equal("GET"))
				Expect(call.Path).To(Equal("/v1/kv/sample_key?index=55&wait=60s"))
				Expect(call.Snd).To(BeNil())
			})
		})
	})

})
