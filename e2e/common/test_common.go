package common

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-env/chaos"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/logging"
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
	l := logging.GetTestLogger(t)
	testEnvConfig := GetTestEnvConfig(t)

	ethChart := ethereum.New(nil)
	ethNetworkName := ethChart.GetProps().(*ethereum.Props).NetworkName

	// we adding the same chart with different index and executing multi-stage deployment
	// connections should be renewed
	e := environment.New(testEnvConfig)
	err := e.AddHelm(ethChart).
		AddHelm(chainlink.New(0, nil)).
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

	err = e.AddHelm(chainlink.New(1, nil)).
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
		l.Info().Str("URL", u).Send()
		res, err := r.R().Get(u)
		require.NoError(t, err)
		require.Equal(t, "200 OK", res.Status())
	}
}

func TestConnectWithoutManifest(t *testing.T) {
	l := logging.GetTestLogger(t)
	existingEnvConfig := GetTestEnvConfig(t)
	testEnvConfig := GetTestEnvConfig(t)
	existingEnvAlreadySetupVar := "ENV_ALREADY_EXISTS"
	var existingEnv *environment.Environment

	// only run this section if we don't already have an existing environment
	// needed for remote runner based tests to prevent duplicate envs from being created
	if os.Getenv(existingEnvAlreadySetupVar) == "" {
		existingEnv = environment.New(existingEnvConfig)
		l.Info().Str("Namespace", existingEnvConfig.Namespace).Msg("Existing Env Namespace")
		// deploy environment to use as an existing one for the test
		existingEnv.Cfg.JobImage = ""
		existingEnv.AddHelm(ethereum.New(nil)).
			AddHelm(chainlink.New(0, map[string]interface{}{
				"replicas": 1,
			}))
		err := existingEnv.Run()
		require.NoError(t, err)
		// propagate the existing environment to the remote runner
		t.Setenv(fmt.Sprintf("TEST_%s", existingEnvAlreadySetupVar), "abc")
		// set the namespace to the existing one for local runs
		testEnvConfig.Namespace = existingEnv.Cfg.Namespace
	} else {
		l.Info().Msg("Environment already exists, verfying it is correct")
		require.NotEmpty(t, os.Getenv(config.EnvVarNamespace))
		noManifestUpdate, err := strconv.ParseBool(os.Getenv(config.EnvVarNoManifestUpdate))
		require.NoError(t, err, "Failed to parse the no manifest update env var")
		require.True(t, noManifestUpdate)
	}

	// Now run an environment without a manifest like a normal test
	testEnvConfig.NoManifestUpdate = true
	testEnv := environment.New(testEnvConfig)
	l.Info().Msgf("Testing Env Namespace %s", testEnv.Cfg.Namespace)
	err := testEnv.AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		})).
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
	l.Info().Msgf("getting url: %s", url)
	res, err := r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	l.Info().Msgf("getting url: %s", url)
	res, err = r.R().Get(urlGeth)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	l.Info().Msgf("done getting url: %s", url)
}

func Test5NodesSoakEnvironmentWithPVCs(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMSoak(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})
}

func TestWithSingleNodeEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMOneNode(testEnvConfig)
	err := e.Run()
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
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
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
	e := presets.EVMMinimalLocalBS(testEnvConfig)
	err := e.Run()
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
	e := presets.EVMReorg(testEnvConfig)
	err := e.Run()
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
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		AddHelm(chainlink.New(1, nil))
	err := e.Run()
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
	l := logging.GetTestLogger(t)
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
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		}))
	err := e.Run()
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
		l.Info().Msg("Expected error was found")
	}, "1m", "3s").Should(gomega.Succeed())

	l.Info().Msg("Waiting for Pod to start back up")
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

func TestRolloutRestart(t *testing.T, statefulSet bool) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	cd, err := chainlink.NewDeployment(5, map[string]interface{}{
		"db": map[string]interface{}{
			"stateful": true,
			"capacity": "1Gi",
		},
	})
	require.NoError(t, err, "failed to create chainlink deployment")
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelmCharts(cd)
	err = e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, e.Shutdown())
	})

	if statefulSet {
		err = e.RolloutStatefulSets()
		require.NoError(t, err, "failed to rollout statefulsets")
	} else {
		err = e.Client.RolloutRestartBySelector(e.Cfg.Namespace, "deployment", "envType=chainlink-env-test")
		require.NoError(t, err, "failed to rollout restart deployment")
	}

	err = e.Run()
	require.NoError(t, err, "failed to run environment")
}
