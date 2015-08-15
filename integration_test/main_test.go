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

	Describe("routing", func() {
		BeforeEach(func() {
			args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
			args = append(args, "-dev")
			args = append(args, "-enableHTTPS=false")
			session = startMainWithArgs(args...)
			Eventually(session).Should(gbytes.Say("garagepi started"))
		})

		It("Should accept GET requests to /", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d", httpPort))
			Expect(err).NotTo(HaveOccurred())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should reject GET requests to /toggle with 404", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/toggle", httpPort))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("Should accept POST requests to /toggle", func() {
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/toggle", httpPort), "", strings.NewReader(""))
			Expect(err).NotTo(HaveOccurred())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should accept GET requests to /light", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/light", httpPort))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("Should accept POST requests to /light", func() {
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/light", httpPort), "", strings.NewReader(""))
			Expect(err).NotTo(HaveOccurred())
			validateSuccessNonZeroLengthBody(resp)
		})

		It("Should accept GET requests to /webcam", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/webcam", httpPort))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("Should serve static files", func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/static/css/application.css", httpPort))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("request handling", func() {
		BeforeEach(func() {
			args = append(args, "-dev")
		})

		Context("when enableHTTP is true", func() {
			BeforeEach(func() {
				args = append(args, "-enableHTTP=true")
				args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
				args = append(args, "-enableHTTPS=false")
			})

			It("accepts HTTP connections", func() {
				session = startMainWithArgs(args...)
				Eventually(session).Should(gbytes.Say("garagepi started"))

				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", httpPort))
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when enableHTTPS is true", func() {
			BeforeEach(func() {
				args = append(args, "-enableHTTPS=true")
				args = append(args, fmt.Sprintf("-httpsPort=%d", httpsPort))
			})

			It("exits with error when -keyFile is not provided", func() {
				args = append(args, "-certFile=someCert")
				args = append(args, "-keyFile=")

				session = startMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(2))
			})

			It("exits with error when -certFile is not provided", func() {
				args = append(args, "-keyFile=someKey")
				args = append(args, "-certFile=")

				session = startMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(2))
			})

			Context("when both -certFile and -keyFile are provided", func() {
				var (
					keyFile  string
					certFile string
				)

				BeforeEach(func() {
					testDir := getDirOfCurrentFile()
					fixturesDir := filepath.Join(testDir, "..", "fixtures")
					keyFile = filepath.Join(fixturesDir, "key.pem")
					certFile = filepath.Join(fixturesDir, "cert.pem")

					args = append(args, "-keyFile="+keyFile)
					args = append(args, "-certFile="+certFile)
				})

				It("accepts HTTPS connections", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

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

					resp, err := client.Get(fmt.Sprintf("https://localhost:%d/", httpsPort))
					Expect(err).NotTo(HaveOccurred())
					validateSuccessNonZeroLengthBody(resp)
				})
			})
		})
	})

	Describe("authentication", func() {
		Context("when dev is enabled", func() {
			BeforeEach(func() {
				args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
				args = append(args, "-dev")
				args = append(args, "-enableHTTPS=false")
			})

			It("accepts unauthenticated requests", func() {
				session = startMainWithArgs(args...)
				Eventually(session).Should(gbytes.Say("garagepi started"))

				resp, err := http.Get(fmt.Sprintf("http://localhost:%d", httpPort))
				Expect(err).NotTo(HaveOccurred())
				validateSuccessNonZeroLengthBody(resp)
			})
		})

		Context("when dev is disabled", func() {
			BeforeEach(func() {
				args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
				args = append(args, "-dev=false")
				args = append(args, "-enableHTTPS=false")
			})

			It("exits with error when -username is not provided", func() {
				args = append(args, "-username=")
				args = append(args, "-password=password")

				session = startMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(2))
			})

			It("exits with error when -password is not provided", func() {
				args = append(args, "-username=username")
				args = append(args, "-password=")

				session = startMainWithArgs(args...)
				Eventually(session).Should(gexec.Exit(2))
			})

			Context("when username and password are provided", func() {
				BeforeEach(func() {
					args = append(args, "-username=username")
					args = append(args, "-password=password")
				})

				It("rejects unauthenticated requests", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

					resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", httpPort))
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("rejects unauthorized requests", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

					req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", httpPort), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth("username", "badpassword")

					client := &http.Client{}
					resp, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
				})

				It("accepts authorized requests", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

					req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", httpPort), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth("username", "password")

					client := &http.Client{}
					resp, err := client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				})
			})
		})
	})
})

func getDirOfCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
