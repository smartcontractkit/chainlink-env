package env_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-env/e2e/common"
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