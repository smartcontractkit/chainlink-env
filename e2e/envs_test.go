package e2e_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/presets"
	"github.com/stretchr/testify/require"
)

const (
	TestEnvType = "chainlink-env-test"
)

var (
	testSelector = fmt.Sprintf("envType=%s", TestEnvType)
)

func getTestEnvConfig(t *testing.T) *environment.Config {
	return &environment.Config{
		NamespacePrefix: TestEnvType,
		Labels:          []string{testSelector},
		Test:            t,
	}
}

func TestConnectWithoutManifest(t *testing.T) {
	nsPrefix := fmt.Sprintf("test-no-manifest-connection-%s", uuid.NewString()[0:5])
	e := environment.New(&environment.Config{
		NamespacePrefix: nsPrefix,
		Test:            t,
	}).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		}))
	err := e.Run()
	require.NoError(t, err)
	// nolint
	defer e.Shutdown()
	_ = os.Setenv("ENV_NAMESPACE", e.Cfg.Namespace)
	_ = os.Setenv("NO_MANIFEST_UPDATE", "true")
	err = environment.New(&environment.Config{
		NamespacePrefix: nsPrefix,
		Test:            t,
	}).
		Run()
	require.NoError(t, err)
	url, err := e.Fwd.FindPort("chainlink-0:0", "node", "access").As(client.LocalConnection, client.HTTP)
	require.NoError(t, err)
	urlGeth, err := e.Fwd.FindPort("geth:0", "geth-network", "http-rpc").As(client.LocalConnection, client.HTTP)
	require.NoError(t, err)
	r := resty.New()
	res, err := r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	res, err = r.R().Get(urlGeth)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
}

func Test5NodesSoakEnvironmentWithPVCs(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMSoak(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
}

func TestWithSingleNodeEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMOneNode(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
}

func TestMinResources5NodesEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
}

func TestMinResources5NodesEnvWithBlockscout(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocalBS(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
}

// TODO: fixme, use proper TOML config
// func Test5NodesPlus2MiningGethsReorgEnv(t *testing.T) {
// 	t.Parallel()
// 	testEnvConfig := getTestEnvConfig(t)
// 	e := presets.EVMReorg(testEnvConfig)
// 	err := e.Run()
// 	// nolint
// 	defer e.Shutdown()
// 	require.NoError(t, err)
// }

func TestMultipleInstancesOfTheSameType(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		AddHelm(chainlink.New(1, nil))
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
}

// Note: this test only works when run with a remote runner
func TestFundReturnShutdownLogic(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	if e.WillUseRemoteRunner() {
		require.Error(t, err, "Should return an error")
		return
	}
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
	fmt.Println(environment.FAILED_FUND_RETURN)
}

func TestRemoteRunnerMultipleRunCommands(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
	e.AddHelm(chainlink.New(1, nil))
	err = e.Run()
	require.NoError(t, err)
}

func TestRemoteRunnerOneSetupWithMultipeTests(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}

	log.Info().Str("Test", "Before").Msg("Before Tests")
	t.Run("do one", func(t *testing.T) {
		log.Info().Str("Test", "One").Msg("Inside test")
		time.Sleep(1 * time.Second)
	})
	t.Run("do two", func(t *testing.T) {
		log.Info().Str("Test", "Two").Msg("Inside test")
		time.Sleep(1 * time.Second)
	})
	log.Info().Str("Test", "After").Msg("After Tests")
}
