package env_test

import (
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-env/e2e/common"
	"github.com/smartcontractkit/chainlink-env/environment"
	mercury_server "github.com/smartcontractkit/chainlink-env/pkg/helm/mercury-server"
	"github.com/stretchr/testify/require"
)

func TestOCIChart(t *testing.T) {
	t.Parallel()
	url := "oci://my-erc/my-repo"
	ver := "v1.0.8"
	t.Run("invalid URL", func(t *testing.T) {
		t.Parallel()
		err := environment.New(nil).
			AddHelm(mercury_server.New(url, ver, nil)).
			Run()
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf(environment.ErrInvalidOCI, url), "The error is not ErrInvalidOCI")
	})
	t.Run("failed to pull a valid URL", func(t *testing.T) {
		t.Parallel()
		url := "oci://my-erc/my-repo/my-chart"
		err := environment.New(nil).
			AddHelm(mercury_server.New(url, ver, nil)).
			Run()
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf(environment.ErrOCIPull, url), "The error is not ErrInvalidOCI")
	})
}

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

func TestEmptyEnvironmentStartup(t *testing.T) {
	common.TestEmptyEnvironmentStartup(t)
}
