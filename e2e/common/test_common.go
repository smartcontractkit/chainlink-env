package common

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/chaos"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/presets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestEnvType = "chainlink-env-test"
)

var (
	testSelector = fmt.Sprintf("envType=%s", TestEnvType)
)

func GetTestEnvConfig(t *testing.T) *environment.Config {
	return &environment.Config{
		NamespacePrefix: TestEnvType,
		Labels:          []string{testSelector},
		Test:            t,
	}
}

func TestMultiStageMultiManifestConnection(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)

	ethChart := ethereum.New(nil)
	ethNetworkName := ethChart.GetProps().(*ethereum.Props).NetworkName

	// we adding the same chart with different index and executing multi-stage deployment
	// connections should be renewed
	e := environment.New(testEnvConfig)
	chainlinkChart, err := chainlink.New(0, nil)
	require.NoError(t, err)
	err = e.AddHelm(ethChart).
		AddHelm(chainlinkChart).
		Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
	require.Len(t, e.URLs[chainlink.NodesLocalURLsKey], 1)
	require.Len(t, e.URLs[chainlink.NodesInternalURLsKey], 1)
	require.Len(t, e.URLs[chainlink.DBsLocalURLsKey], 1)
	require.Len(t, e.URLs, 7)

	chainlinkChart2, err := chainlink.New(0, nil)
	require.NoError(t, err)
	err = e.AddHelm(chainlinkChart2).
		Run()
	require.NoError(t, err)
	require.Len(t, e.URLs[chainlink.NodesLocalURLsKey], 2)
	require.Len(t, e.URLs[chainlink.NodesInternalURLsKey], 2)
	require.Len(t, e.URLs[chainlink.DBsLocalURLsKey], 2)
	require.Len(t, e.URLs, 7)

	urls := make([]string, 0)
	if e.Cfg.InsideK8s {
		urls = append(urls, e.URLs[chainlink.NodesInternalURLsKey]...)
		urls = append(urls, e.URLs[ethNetworkName+"_internal_http"]...)
	} else {
		urls = append(urls, e.URLs[chainlink.NodesLocalURLsKey]...)
		urls = append(urls, e.URLs[ethNetworkName+"_http"]...)
	}

	r := resty.New()
	for _, u := range urls {
		log.Info().Str("URL", u).Send()
		res, err := r.R().Get(u)
		require.NoError(t, err)
		require.Equal(t, "200 OK", res.Status())
	}
}

func TestConnectWithoutManifest(t *testing.T) {
	existingEnvConfig := GetTestEnvConfig(t)
	testEnvConfig := GetTestEnvConfig(t)
	existingEnvAlreadySetupVar := "ENV_ALREADY_EXISTS"
	var existingEnv *environment.Environment

	// only run this section if we don't already have an existing environment
	// needed for remote runner based tests to prevent duplicate envs from being created
	if os.Getenv(existingEnvAlreadySetupVar) == "" {
		existingEnv = environment.New(existingEnvConfig)
		t.Log("Existing Env Namespace", existingEnv.Cfg.Namespace)
		// deploy environment to use as an existing one for the test
		existingEnv.Cfg.JobImage = ""
		chainlinkChart, err := chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		})
		require.NoError(t, err)
		existingEnv.AddHelm(ethereum.New(nil)).
			AddHelm(chainlinkChart)
		err = existingEnv.Run()
		require.NoError(t, err)
		// propagate the existing environment to the remote runner
		t.Setenv(fmt.Sprintf("TEST_%s", existingEnvAlreadySetupVar), "abc")
		// set the namespace to the existing one for local runs
		testEnvConfig.Namespace = existingEnv.Cfg.Namespace
	} else {
		t.Log("Environment already exists, verfying it is correct")
		require.NotEmpty(t, os.Getenv(config.EnvVarNamespace))
		noManifestUpdate, err := strconv.ParseBool(os.Getenv(config.EnvVarNoManifestUpdate))
		require.NoError(t, err, "Failed to parse the no manifest update env var")
		require.True(t, noManifestUpdate)
	}

	// Now run an environment without a manifest like a normal test
	testEnvConfig.NoManifestUpdate = true
	testEnv := environment.New(testEnvConfig)
	t.Log("Testing Env Namespace", testEnv.Cfg.Namespace)
	chainlinkChart, err := chainlink.New(0, map[string]interface{}{
		"replicas": 1,
	})
	require.NoError(t, err)
	err = testEnv.AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart).
		Run()
	require.NoError(t, err)
	if testEnv.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, testEnv.Shutdown())
	})

	connection := client.LocalConnection
	if testEnv.Cfg.InsideK8s {
		connection = client.RemoteConnection
	}
	url, err := testEnv.Fwd.FindPort("chainlink-0:0", "node", "access").As(connection, client.HTTP)
	require.NoError(t, err)
	urlGeth, err := testEnv.Fwd.FindPort("geth:0", "geth-network", "http-rpc").As(connection, client.HTTP)
	require.NoError(t, err)
	r := resty.New()
	t.Log("getting", url)
	res, err := r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	t.Log("getting", url)
	res, err = r.R().Get(urlGeth)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	t.Log("done", url)
}

func Test5NodesSoakEnvironmentWithPVCs(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e, err := presets.EVMSoak(testEnvConfig)
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func TestWithSingleNodeEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e, err := presets.EVMOneNode(testEnvConfig)
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func TestMinResources5NodesEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e, err := presets.EVMMinimalLocal(testEnvConfig)
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func TestMinResources5NodesEnvWithBlockscout(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e, err := presets.EVMMinimalLocalBS(testEnvConfig)
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func Test5NodesPlus2MiningGethsReorgEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e, err := presets.EVMReorg(testEnvConfig)
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func TestMultipleInstancesOfTheSameType(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	chainlinkChart1, err := chainlink.New(0, nil)
	require.NoError(t, err)
	chainlinkChart2, err := chainlink.New(1, nil)
	require.NoError(t, err)
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart1).
		AddHelm(chainlinkChart2)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

// TestWithChaos runs a test with chaos injected into the environment.
func TestWithChaos(t *testing.T) {
	t.Parallel()
	appLabel := "chainlink-0"
	testCase := struct {
		chaosFunc  chaos.ManifestFunc
		chaosProps *chaos.Props
	}{
		chaos.NewFailPods,
		&chaos.Props{
			LabelsSelector: &map[string]*string{"app": a.Str(appLabel)},
			DurationStr:    "30s",
		},
	}
	testEnvConfig := GetTestEnvConfig(t)
	chainlinkChart, err := chainlink.New(0, map[string]interface{}{
		"replicas": 1,
	})
	require.NoError(t, err)
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})

	connection := client.LocalConnection
	if e.Cfg.InsideK8s {
		connection = client.RemoteConnection
	}
	url, err := e.Fwd.FindPort("chainlink-0:0", "node", "access").As(connection, client.HTTP)
	require.NoError(t, err)
	r := resty.New()
	res, err := r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())

	// start chaos
	_, err = e.Chaos.Run(testCase.chaosFunc(e.Cfg.Namespace, testCase.chaosProps))
	require.NoError(t, err)
	gom := gomega.NewGomegaWithT(t)
	gom.Eventually(func(g gomega.Gomega) {
		res, err = r.R().Get(url)
		g.Expect(err).Should(gomega.HaveOccurred())
		t.Log("Expected error was found")
	}, "1m", "3s").Should(gomega.Succeed())

	t.Log("Waiting for Pod to start back up")
	err = e.Run()
	require.NoError(t, err)

	// verify that the node can recieve requests again
	url, err = e.Fwd.FindPort("chainlink-0:0", "node", "access").As(connection, client.HTTP)
	require.NoError(t, err)
	res, err = r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
}

func TestEmptyEnvironmentStartup(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := environment.New(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}
