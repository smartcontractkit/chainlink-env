package pkg

import a "github.com/smartcontractkit/chainlink-env/pkg/alias"

// Control labels used to list envs created by the wizard
const (
	ControlLabelKey        = "generatedBy"
	ControlLabelValue      = "cdk8s"
	ControlLabelEnvTypeKey = "envType"
	TTLLabelKey            = "janitor/ttl"
)

// Environment types, envs got selected by having a label of that type
const (
	EnvTypeEVM1             = "evm-1-minimal"
	EnvTypeEVM5             = "evm-5-minimal"
	EnvTypeEVM5RemoteRunner = "evm-5-remote-runner"
	EnvTypeMultinetwork     = "evm-5-multinetwork"
	EnvTypeETH5Reorg        = "evm-5-reorg"
	EnvTypeEVM5BS           = "evm-5-minimal-blockscout"
	EnvTypeEVM5Soak         = "evm-5-soak"
	EnvTypeSolana5          = "solana-5-default"
)

type ResourcesMode int

const (
	MinimalLocalResourcesMode ResourcesMode = iota
	SoakResourcesMode
)

func PGIsReadyCheck() *[]*string {
	return &[]*string{
		a.Str("pg_isready"),
		a.Str("-U"),
		a.Str("postgres"),
	}
}
