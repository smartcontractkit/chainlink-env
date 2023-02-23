package e2e_remote_runner_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/e2e/common"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/presets"
	"github.com/stretchr/testify/require"
)

func TestMultiStageMultiManifestConnection(t *testing.T) {
	common.TestMultiStageMultiManifestConnection(t)
}

func TestConnectWithoutManifest(t *testing.T) {
	common.TestConnectWithoutManifest(t)
}

func Test5NodesSoakEnvironmentWithPVCs(t *testing.T) {
	common.Test5NodesSoakEnvironmentWithPVCs(t)
}

func TestWithSingleNodeEnv(t *testing.T) {
	common.TestWithSingleNodeEnv(t)
}

func TestMinResources5NodesEnv(t *testing.T) {
	common.TestMinResources5NodesEnv(t)
}

func TestMinResources5NodesEnvWithBlockscout(t *testing.T) {
	common.TestMinResources5NodesEnvWithBlockscout(t)
}

func TestMultipleInstancesOfTheSameType(t *testing.T) {
	common.TestMultipleInstancesOfTheSameType(t)
}

func Test5NodesPlus2MiningGethsReorgEnv(t *testing.T) {
	common.Test5NodesPlus2MiningGethsReorgEnv(t)
}

func TestWithChaos(t *testing.T) {
	common.TestWithChaos(t)
}

func TestFundReturnShutdownLogic(t *testing.T) {
	t.Parallel()
	testEnvConfig := common.GetTestEnvConfig(t)
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

func TestRemoteRunnerOneSetupWithMultipeTests(t *testing.T) {
	t.Parallel()
	testEnvConfig := common.GetTestEnvConfig(t)
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
