package acceptance_test

/*

This test is about:

ASGs should not interact with container network policy

ASGs do not affect the default-deny behavior of the overlay network

An ASG with an IP range that covers a container IP does not allow traffic to that container




Install an ASG that allows access to 0.0.0.0/0.

Push 1 instance of a test app.

Push 1 proxy app instance.

Ping test app.  It tests connectivity from proxy to itself.
That attempt should fail for the following reason:
- test should be able to reach proxy via HTTP Router
- proxy should not be able to connect to test app via overlay network.

*/

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const Timeout_Short = 10 * time.Second

var _ = Describe("ASGs and Overlay Policy interaction", func() {
	var (
		appProxy     string
		appSmoke     string
		appInstances int
		prefix       string
		spaceName    string
		orgName      string
	)

	BeforeEach(func() {
		prefix = testConfig.Prefix

		orgName = prefix + "interaction-org"
		Expect(cf.Cf("target", "-o", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))

		spaceName := prefix + "interaction-space"
		Expect(cf.Cf("create-space", spaceName).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(Timeout_Push)).To(gexec.Exit(0))

		appInstances = testConfig.AppInstances

		appProxy = prefix + "proxy"
		appSmoke = prefix + "smoke"

	})

	AfterEach(func() {
		Expect(cf.Cf("delete-space", spaceName, "-f").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	It("allows the user to configure policies", func(done Done) {
		By("creating and binding a wide open security group")

		By("pushing the proxy and smoke test apps")
		pushApp(appProxy, "proxy")
		pushApp(appSmoke, "smoke", "--no-start")
		setEnv(appSmoke, "PROXY_APP_URL", fmt.Sprintf("http://%s.%s", appProxy, config.AppsDomain))
		start(appSmoke)

		scaleApp(appSmoke, appInstances)

		ports := []int{8080}
		appsSmoke := []string{appSmoke}

		By(fmt.Sprintf("checking that %s can NOT reach %s", appProxy, appsSmoke))
		assertSelfProxyConnectionFails(appSmoke, appInstances)

		close(done)
	}, 30*60 /* <-- overall spec timeout in seconds */)
})

func assertSelfProxyConnectionFails(sourceApp string, appInstances int) {
	for i := 0; i < appInstances; i++ {
		assertSelfProxyResponseContains(sourceApp, "FAILED")
	}
}

func assertSelfProxyResponseContains(sourceAppName, desiredResponse string) {
	proxyTest := func() (string, error) {
		resp, err := httpGetBytes(fmt.Sprintf("http://%s.%s/selfproxy", sourceAppName, config.AppsDomain))
		if err != nil {
			return "", err
		}
		return string(resp.Body), nil
	}
	Eventually(proxyTest, 10*time.Second, 500*time.Millisecond).Should(ContainSubstring(desiredResponse))
}

func pushApp(appName, kind string, extraArgs ...string) {
	args := append([]string{
		"push", appName,
		"-p", appDir(kind),
		"-f", defaultManifest(kind),
	}, extraArgs...)
	Expect(cf.Cf(args...).Wait(Timeout_Push)).To(gexec.Exit(0))
}

func setEnv(appName, envVar, value string) {
	Expect(cf.Cf(
		"set-env", appName,
		envVar, value,
	).Wait(Timeout_Short)).To(gexec.Exit(0))
}

func start(appName string) {
	Expect(cf.Cf(
		"start", appName,
	).Wait(Timeout_Push)).To(gexec.Exit(0))
}
