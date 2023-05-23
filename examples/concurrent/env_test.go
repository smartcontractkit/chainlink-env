package concurrent_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/stretchr/testify/require"
)

func TestConcurrentEnvs(t *testing.T) {
	t.Run("test 1", func(t *testing.T) {
		t.Parallel()
		chainlinkChart, err := chainlink.New(0, nil)
		require.NoError(t, err)
		e := environment.New(nil).
			AddHelm(chainlinkChart)
		defer e.Shutdown()
		err = e.Run()
		require.NoError(t, err)
	})
	t.Run("test 2", func(t *testing.T) {
		t.Parallel()
		chainlinkChart, err := chainlink.New(0, nil)
		require.NoError(t, err)
		e := environment.New(nil).
			AddHelm(chainlinkChart)
		defer e.Shutdown()
		err = e.Run()
		require.NoError(t, err)
		chainlinkChart2, err := chainlink.New(0, map[string]interface{}{
			"replicas": 2,
		})
		require.NoError(t, err)
		err = e.
			ModifyHelm("chainlink-0", chainlinkChart2).Run()
		require.NoError(t, err)
	})
}
