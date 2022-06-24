package e2e_test

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	TestEnvType = "test"
)

var (
	testSelector = fmt.Sprintf("envType=%s", TestEnvType)
)

func cleanEnvs(t *testing.T) {
	c := client.NewK8sClient()
	nsList, err := c.ListNamespaces(testSelector)
	require.NoError(t, err)
	for _, ns := range nsList.Items {
		_ = c.RemoveNamespace(ns.Name)
	}
}

func TestSimpleEnv(t *testing.T) {
	t.Run("test 5 nodes soak environment with PVCs", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.EVMSoak(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		})
		require.NoError(t, err)
	})
	t.Run("smoke test with a single node env", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.EVMOneNode(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		})
		require.NoError(t, err)
	})
	t.Run("test min resources 5 nodes env", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.EVMMinimalLocal(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		})
		require.NoError(t, err)
	})
	t.Run("test min resources 5 nodes env with blockscout", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.EVMMinimalLocalBS(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		})
		require.NoError(t, err)
	})
	t.Run("test 5 nodes + 2 mining geths, reorg env", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.EVMReorg(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		})
		require.NoError(t, err)
	})
	t.Run("test 5 nodes env with an external network", func(t *testing.T) {
		defer cleanEnvs(t)
		err := presets.MultiNetwork(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		}, &presets.MultiNetworkOpts{})
		require.NoError(t, err)
	})
	t.Run("test multiple instances of the same type", func(t *testing.T) {
		defer cleanEnvs(t)
		err := environment.New(&environment.Config{
			Labels: []string{fmt.Sprintf("envType=%s", TestEnvType)},
		}).
			AddHelm(ethereum.New(nil)).
			AddHelm(chainlink.New(0, nil)).
			AddHelm(chainlink.New(1, nil)).
			Run()
		require.NoError(t, err)
	})
	// TODO: assert export data
	// TODO: App.SynthYaml() is not thread safe, global lock in Go doesn't help, Node JS issue?
}
