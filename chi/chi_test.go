package chi_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "configstate/chi"
)

func TestChi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chi Suite")
}

var _ = Describe("Chi", func() {

	var (
		method  string
		path    string
		handler http.HandlerFunc
	)

	BeforeEach(func() {
		method = "GET"
		path = "/api/borgu/v1/templates"
		handler = func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(`{"status":"ok"}`))
		}
	})

	Describe("handling a get photos request", func() {
		var (
			mux *chi.Mux
			che *Chi
		)

		JustBeforeEach(func() {
			che.Set(method, path, handler)
		})

		When("all is well", func() {
			BeforeEach(func() {
				mux = chi.NewMux()
				che = &Chi{
					Mux: mux,
				}
			})

			It("should respond with photos", func() {
				Expect(mux.Routes()[0].Pattern).To(Equal("/api/borgu/v1/templates"))
			})
		})
	})
})
