package homepage_test

import (
	"html/template"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	light_fakes "github.com/robdimsdale/garagepi/api/light/fakes"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	"github.com/robdimsdale/garagepi/web/homepage"
	login_fakes "github.com/robdimsdale/garagepi/web/login/fakes"
)

const (
	headTemplate = `
{{define "head"}}
some head text
{{end}}`
	homepageTemplate = `
{{define "homepage"}}
{{template "head"}}
some text here
{{end}}`
)

var (
	fakeLogger         lager.Logger
	fakeLightHandler   *light_fakes.FakeHandler
	fakeLoginHandler   *login_fakes.FakeHandler
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	hh           homepage.Handler

	templates *template.Template
)

var _ = Describe("Homepage", func() {
	BeforeEach(func() {
		fakeLogger = lagertest.NewTestLogger("homepage handle test")
		fakeLightHandler = new(light_fakes.FakeHandler)
		fakeLoginHandler = new(login_fakes.FakeHandler)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		var err error

		templates, err = template.New("head").Parse(headTemplate)
		Expect(err).NotTo(HaveOccurred())
		templates, err = templates.New("homepage").Parse(homepageTemplate)
		Expect(err).NotTo(HaveOccurred())

		hh = homepage.NewHandler(
			fakeLogger,
			templates,
			fakeLightHandler,
			fakeLoginHandler,
		)

		dummyRequest = new(http.Request)
	})

	Describe("Homepage Handling", func() {
		It("Should write the contents of the homepage template to the response writer", func() {
			hh.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteCallCount()).To(BeNumerically(">=", 1))
		})
	})
})
