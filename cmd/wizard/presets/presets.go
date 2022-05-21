package presets

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/chains/ethereum"
)

func EnvEVMMinimalLocal(config *environment.Config) error {
	return environment.New(config).
		DeployOrConnect(
			chainlink.NewChart(
				&chainlink.Props{
					Namespace: "chainlink-env",
					Labels:    []string{fmt.Sprintf("envType=%s", chainlink.EnvTypeEVM5)},
					ChainProps: []interface{}{
						&ethereum.Props{},
					},
					ResourcesMode: pkg.MinimalLocalResourcesMode,
					AppVersions: []chainlink.VersionProps{
						{
							Image:     "public.ecr.aws/chainlink/chainlink",
							Tag:       "1.4.1-root",
							Instances: 5,
						},
					},
				}))
}

func EnvEVMSoak(config *environment.Config) error {
	return environment.New(config).
		DeployOrConnect(
			chainlink.NewChart(
				&chainlink.Props{
					Namespace: "chainlink-env",
					Labels:    []string{fmt.Sprintf("envType=%s", chainlink.EnvTypeEVM5Soak)},
					ChainProps: []interface{}{
						&ethereum.Props{},
					},
					Persistence:   chainlink.PersistenceProps{Capacity: "20Gi"},
					ResourcesMode: pkg.SoakResourcesMode,
					AppVersions: []chainlink.VersionProps{
						{
							Image:     "public.ecr.aws/chainlink/chainlink",
							Tag:       "1.4.1-root",
							Instances: 5,
						},
					},
				}))
}
