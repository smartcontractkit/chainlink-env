package e2e_test

import (
	"fmt"
	"os"
	"path"
	"testing"

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

// TODO: GHA have some problems with filepaths, using afero will add another param for the filesystem
func TestLogs(t *testing.T) {
	t.Skip("problems with GHA and afero")
	t.Parallel()
	testEnvConfig := getTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	// nolint
	defer e.Shutdown()
	require.NoError(t, err)
	logDir := "./logs/mytest"
	err = e.DumpLogs(logDir)
	require.NoError(t, err)
	clDir := "chainlink-0_0"
	_, err = os.Stat(path.Join(logDir, clDir))
	require.NoError(t, err)
	fi, err := os.Stat(path.Join(logDir, clDir, "node.log"))
	require.NoError(t, err)
	require.Greaterf(t, fi.Size(), int64(0), "file is empty")
	_, err = os.Stat(path.Join(logDir, "geth_0"))
	require.NoError(t, err)
	_, err = os.Stat(path.Join(logDir, "mockserver_0"))
	require.NoError(t, err)
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
	defer e.Shutdown()
	require.NoError(t, err)
	err = e.Run()
	require.NoError(t, err)
}
