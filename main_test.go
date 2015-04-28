package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	session *gexec.Session
)

func startMainWithArgs(args ...string) *gexec.Session {
	args = append(args, fmt.Sprintf("-port=%d", port))
	command := exec.Command(garagepiBinPath, args...)
	var err error
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gbytes.Say(".*garagepi started"))
	return session
}

func validateSuccessAnyLengthBody(resp *http.Response, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	validateBody(resp, true)
}

func validateSuccessNonZeroLengthBody(resp *http.Response, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	validateBody(resp, false)
}

func validateBody(resp *http.Response, anySize bool) {
	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	if anySize {
		Expect(len(body)).Should(BeNumerically(">=", 0))
	} else {

		Expect(len(body)).Should(BeNumerically(">", 0))
	}
}

var _ = Describe("GaragepiExecutable", func() {
	BeforeEach(func() {
		startMainWithArgs()
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("Should accept GET requests to /", func() {
		validateSuccessNonZeroLengthBody(http.Get(fmt.Sprintf("http://127.0.0.1:%d", port)))
	})

	It("Should reject GET requests to /toggle with 404", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/toggle", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("Should accept POST requests to /toggle", func() {
		validateSuccessNonZeroLengthBody(http.Post(fmt.Sprintf("http://127.0.0.1:%d/toggle", port), "", strings.NewReader("")))
	})

	It("Should accept GET requests to /light", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/light", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
	})

	It("Should accept POST requests to /light", func() {
		validateSuccessNonZeroLengthBody(http.Post(fmt.Sprintf("http://127.0.0.1:%d/light", port), "", strings.NewReader("")))
	})

	It("Should accept GET requests to /webcam", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/webcam", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
	})
})
