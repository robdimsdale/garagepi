package smoke_test

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func startMainWithArgs(args ...string) *gexec.Session {
	command := exec.Command(garagepiBinPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
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
	var (
		session *gexec.Session
		args    []string
	)

	BeforeEach(func() {
		args = []string{}
	})

	AfterEach(func() {
		session.Terminate()
	})

	Describe("request handling", func() {
		BeforeEach(func() {
			args = append(args, "-dev")
		})

		Context("when enableHTTPS is true", func() {
			BeforeEach(func() {
				args = append(args, "-enableHTTPS=true")
				args = append(args, fmt.Sprintf("-httpsPort=%d", httpsPort))
				args = append(args, fmt.Sprintf("-redirectPort=%d", httpsPort))
			})

			Context("when both -certFile and -keyFile are provided", func() {
				var (
					keyFile  string
					certFile string

					client *http.Client
				)

				BeforeEach(func() {
					testDir := getDirOfCurrentFile()
					fixturesDir := filepath.Join(testDir, "..", "fixtures")
					keyFile = filepath.Join(fixturesDir, "key.pem")
					certFile = filepath.Join(fixturesDir, "cert.pem")

					args = append(args, "-keyFile="+keyFile)
					args = append(args, "-certFile="+certFile)

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
					client = &http.Client{Transport: transport}
				})

				It("accepts HTTPS connections", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

					resp, err := client.Get(fmt.Sprintf("https://localhost:%d/", httpsPort))
					Expect(err).NotTo(HaveOccurred())
					validateSuccessNonZeroLengthBody(resp)
				})

				Context("when enableHTTP is true", func() {
					BeforeEach(func() {
						args = append(args, "-enableHTTP=true")
						args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
					})

					Context("when forceHTTPS is true", func() {
						BeforeEach(func() {
							args = append(args, "-forceHTTPS=true")
						})

						It("redirects HTTP to HTTPS", func() {
							session = startMainWithArgs(args...)
							Eventually(session).Should(gbytes.Say("garagepi started"))

							req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", httpPort), nil)
							Expect(err).NotTo(HaveOccurred())

							transport := http.Transport{}
							resp, err := transport.RoundTrip(req)

							Expect(resp.StatusCode).To(Equal(http.StatusFound))

							expectedLocation := fmt.Sprintf("localhost:%d", httpsPort)

							location, err := resp.Location()
							Expect(err).NotTo(HaveOccurred())
							Expect(location.Scheme).To(Equal("https"))
							Expect(location.Host).To(Equal(expectedLocation))
						})
					})
				})
			})
		})
	})
})

func getDirOfCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
