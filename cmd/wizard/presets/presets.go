package presets

import (
	"github.com/smartcontractkit/chainlink-env/chainlink"
	"github.com/smartcontractkit/chainlink-env/chains/ethereum"
	"github.com/smartcontractkit/chainlink-env/environment"
)

func EnvEVMDefault(config *environment.Config) error {
	return environment.New(config).
		DeployOrConnect(
			chainlink.NewChart(
				&chainlink.Props{
					Namespace: "chainlink-env",
					Labels:    []string{"envType=evm-5-default"},
					ChainProps: []interface{}{
						&ethereum.Props{},
					},
					AppVersions: []chainlink.VersionProps{
						{
							Image:     "public.ecr.aws/chainlink/chainlink",
							Tag:       "1.4.1-root",
							Instances: 5,
						},
					},
				}))
}
