package e2e_remote_runner_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
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

func TestFundReturnShutdownLogic(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	if e.WillUseRemoteRunner() {
		require.Error(t, err, "Should return an error")
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
	require.NoError(t, err)
	fmt.Println(environment.FAILED_FUND_RETURN)
}

func TestRemoteRunnerMultipleRunCommands(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
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
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
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

func TestMultiStageMultiManifestConnection(t *testing.T) {
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)

	ethChart := ethereum.New(nil)
	ethNetworkName := ethChart.GetProps().(*ethereum.Props).NetworkName

	// we adding the same chart with different index and executing multi-stage deployment
	// connections should be renewed
	e := environment.New(testEnvConfig)
	err := e.AddHelm(ethChart).
		AddHelm(chainlink.New(0, nil)).
		Run()
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
	require.NoError(t, err)
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
	urls = append(urls, e.URLs[chainlink.NodesLocalURLsKey]...)
	urls = append(urls, e.URLs[ethNetworkName+"_http"]...)

	r := resty.New()
	for _, u := range urls {
		log.Info().Str("URL", u).Send()
		res, err := r.R().Get(u)
		require.NoError(t, err)
		require.Equal(t, "200 OK", res.Status())
	}
}
