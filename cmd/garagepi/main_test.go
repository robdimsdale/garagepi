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
	Eventually(session).Should(gbytes.Say("Listening on port"))
	return session
}

var _ = Describe("GaragepiExecutable", func() {
	BeforeEach(func() {
		startMainWithArgs()
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("Accepts GET requests to /", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(body)).Should(BeNumerically(">", 0))
	})

	It("Returns 404 to GET requests to /toggle", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/toggle", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("Accepts POST requests to /toggle", func() {
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/toggle", port), "", strings.NewReader(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(body)).Should(BeNumerically(">", 0))
	})

	It("Accepts GET requests to /webcam", func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/webcam", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		// body will be 0 bytes if upstream webcam server doesn't exist
		Expect(len(body)).Should(BeNumerically(">=", 0))
	})
})
