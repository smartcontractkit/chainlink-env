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
	testSelector  = fmt.Sprintf("envType=%s", TestEnvType)
	testEnvConfig = &environment.Config{
		NamespacePrefix: TestEnvType,
		Labels:          []string{testSelector},
	}
)

// TODO: GHA have some problems with filepaths, using afero will add another param for the filesystem
func TestLogs(t *testing.T) {
	t.Skip("problems with GHA and afero")
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

func TestSimpleEnv(t *testing.T) {
	t.Run("test 5 nodes soak environment with PVCs", func(t *testing.T) {
		e := presets.EVMSoak(testEnvConfig)
		err := e.Run()
		// nolint
		defer e.Shutdown()
		require.NoError(t, err)
	})
	t.Run("smoke test with a single node env", func(t *testing.T) {
		e := presets.EVMOneNode(testEnvConfig)
		err := e.Run()
		// nolint
		defer e.Shutdown()
		require.NoError(t, err)
	})
	t.Run("test min resources 5 nodes env", func(t *testing.T) {
		e := presets.EVMMinimalLocal(testEnvConfig)
		err := e.Run()
		// nolint
		defer e.Shutdown()
		require.NoError(t, err)
	})
	t.Run("test min resources 5 nodes env with blockscout", func(t *testing.T) {
		e := presets.EVMMinimalLocalBS(testEnvConfig)
		err := e.Run()
		// nolint
		defer e.Shutdown()
		require.NoError(t, err)
	})
	// TODO: fixme, use proper TOML config
	//t.Run("test 5 nodes + 2 mining geths, reorg env", func(t *testing.T) {
	//	e := presets.EVMReorg(testEnvConfig)
	//	err := e.Run()
	//	// nolint
	//	defer e.Shutdown()
	//	require.NoError(t, err)
	//})
	t.Run("test multiple instances of the same type", func(t *testing.T) {
		e := environment.New(testEnvConfig).
			AddHelm(ethereum.New(nil)).
			AddHelm(chainlink.New(0, nil)).
			AddHelm(chainlink.New(1, nil))
		err := e.Run()
		// nolint
		defer e.Shutdown()
		require.NoError(t, err)
	})
}
