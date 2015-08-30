package ui_test

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

func startMainWithArgs(args ...string) *gexec.Session {
	command := exec.Command(garagepiBinPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gbytes.Say("garagepi starting"))
	return session
}

var _ = Describe("GaragepiExecutable", func() {
	var (
		args []string
	)

	BeforeEach(func() {
		args = []string{}
	})

	Describe("long-running operation", func() {
		var (
			session *gexec.Session
		)

		AfterEach(func() {
			session.Terminate()
		})

		Describe("UI", func() {
			var page *agouti.Page

			BeforeEach(func() {
				var err error
				page, err = agoutiDriver.NewPage()
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(page.Destroy()).To(Succeed())
			})

			Context("when dev is enabled", func() {
				BeforeEach(func() {
					args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
					args = append(args, "-dev")
				})

				It("does not redirect to /login", func() {
					session = startMainWithArgs(args...)
					Eventually(session).Should(gbytes.Say("garagepi started"))

					url := fmt.Sprintf("http://localhost:%d/", httpPort)

					Expect(page.Navigate(url)).To(Succeed())
					Expect(page).Should(HaveURL(url))
				})
			})

			Context("when dev is disabled and username/password are provided", func() {
				Describe("logging in", func() {
					const (
						username = "some-user"
						password = "8eEd3g4vf0"
					)

					var (
						expectedLoginURL    string
						expectedHomepageURL string
					)

					BeforeEach(func() {
						args = append(args, fmt.Sprintf("-httpPort=%d", httpPort))
						args = append(args, fmt.Sprintf("-username=%s", username))
						args = append(args, fmt.Sprintf("-password=%s", password))

						expectedLoginURL = fmt.Sprintf("http://localhost:%d/login", httpPort)
						expectedHomepageURL = fmt.Sprintf("http://localhost:%d/", httpPort)
					})

					It("allows the user to login and logout", func() {
						session = startMainWithArgs(args...)
						Eventually(session).Should(gbytes.Say("garagepi started"))

						By("redirecting the user to the login form from the home page", func() {
							url := fmt.Sprintf("http://localhost:%d/", httpPort)

							Expect(page.Navigate(url)).To(Succeed())
							Expect(page).Should(HaveURL(expectedLoginURL))
						})

						By("allowing the user to fill out the login form and submit it", func() {
							Eventually(page.FindByLabel("Username")).Should(BeFound())
							Expect(page.FindByLabel("Username").Fill(username)).To(Succeed())
							Expect(page.FindByLabel("Password").Fill(password)).To(Succeed())
							Expect(page.Find("#login").Submit()).To(Succeed())
						})

						By("validating the user is redirected to the home page", func() {
							Eventually(page).Should(HaveURL(expectedHomepageURL))
							Eventually(page.Find("#webcam")).Should(BeFound())
						})

						By("allowing the user to log out", func() {
							Expect(page.Find("#logout").Submit()).To(Succeed())

							Eventually(page).Should(HaveURL(expectedLoginURL))
						})
					})
				})
			})
		})
	})
})
