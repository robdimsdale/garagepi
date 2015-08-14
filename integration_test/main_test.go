package main_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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
	Eventually(session).Should(gbytes.Say("garagepi starting"))
	return session
}

func validateSuccessAnyLengthBody(resp *http.Response, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	validateBody(resp, true)
}

func validateSuccessNonZeroLengthBody(resp *http.Response) {
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

	Describe("routing", func() {
		BeforeEach(func() {
			startMainWithArgs()
			Eventually(session).Should(gbytes.Say("garagepi started"))
		})

		AfterEach(func() {
			session.Terminate()
		})

		It("Should accept GET requests to /", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
			Expect(err).To(BeNil())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should reject GET requests to /toggle with 404", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/toggle", port))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("Should accept POST requests to /toggle", func() {
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/toggle", port), "", strings.NewReader(""))
			Expect(err).To(BeNil())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should accept GET requests to /light", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/light", port))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("Should accept POST requests to /light", func() {
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/light", port), "", strings.NewReader(""))
			Expect(err).To(BeNil())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should accept GET requests to /webcam", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/webcam", port))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("Should serve static files", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/static/css/application.css", port))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when enableHTTPS is true", func() {
		var args []string

		BeforeEach(func() {
			args = append(args, "-enableHTTPS=true")
		})

		It("exits with error when -keyFile is not provided", func() {
			args = append(args, "-certFile=someCert")
			args = append(args, "-keyFile=")
			startMainWithArgs(args...)
			Eventually(session).Should(gexec.Exit(2))
		})

		It("exits with error when -certFile is not provided", func() {
			args = append(args, "-keyFile=someKey")
			args = append(args, "-certFile=")
			startMainWithArgs(args...)
			Eventually(session).Should(gexec.Exit(2))
		})

		Context("when both -certFile and -keyFile are provided", func() {
			var keyFile string
			var certFile string

			BeforeEach(func() {
				testDir := getDirOfCurrentFile()
				fixturesDir := filepath.Join(testDir, "..", "fixtures")
				keyFile = filepath.Join(fixturesDir, "key.pem")
				certFile = filepath.Join(fixturesDir, "cert.pem")

				args = append(args, "-keyFile="+keyFile)
				args = append(args, "-certFile="+certFile)
				startMainWithArgs(args...)
				Eventually(session).Should(gbytes.Say("garagepi started"))
			})

			AfterEach(func() {
				session.Terminate()
			})

			It("Should accept requests via https", func() {
				// Load client cert
				cert, err := tls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					log.Fatal(err)
				}

				// Load CA cert
				caCert, err := ioutil.ReadFile(certFile)
				if err != nil {
					log.Fatal(err)
				}
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)

				// Setup HTTPS client
				tlsConfig := &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      caCertPool,
				}
				tlsConfig.BuildNameToCertificate()
				transport := &http.Transport{TLSClientConfig: tlsConfig}
				client := &http.Client{Transport: transport}

				resp, err := client.Get(fmt.Sprintf("https://localhost:%d", port))
				Expect(err).To(BeNil())
				validateSuccessNonZeroLengthBody(resp)
			})
		})
	})

	Context("when enableHTTPS is false", func() {
		BeforeEach(func() {
			startMainWithArgs()
			Eventually(session).Should(gbytes.Say("garagepi started"))
		})

		AfterEach(func() {
			session.Terminate()
		})

		It("Should accept requests via HTTP", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
			Expect(err).To(BeNil())
			validateSuccessNonZeroLengthBody(resp)
		})
	})
})

func getDirOfCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
