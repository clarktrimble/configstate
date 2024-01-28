package nats

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nats Suite")
}

var _ = Describe("Nats", func() {

	Describe("polling nats for updates", func() {
		var (
		/*
			ctx  context.Context
			nt   *Nats
			data []byte
			err  error
		*/
		)

		When("all is well", func() {
			BeforeEach(func() {
				//data, err = nt.Poll(ctx)
			})

			It("does stuff", func() {
			})
		})
	})
})
