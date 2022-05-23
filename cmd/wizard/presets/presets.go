package presets

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/chains/ethereum"
)

// EnvEVMOneNode local development Chainlink deployment
func EnvEVMOneNode(config *environment.Config) error {
	return environment.New(config).
		DeployOrConnect(
			chainlink.NewChart(
				&chainlink.Props{
					Namespace: "chainlink-env",
					Labels:    []string{fmt.Sprintf("envType=%s", chainlink.EnvTypeEVM1)},
					ChainProps: []interface{}{
						&ethereum.Props{},
					},
					ResourcesMode: pkg.MinimalLocalResourcesMode,
					AppVersions: []chainlink.VersionProps{
						{
							Image:     "public.ecr.aws/chainlink/chainlink",
							Tag:       "1.4.1-root",
							Instances: 1,
						},
					},
				}))
}

// EnvEVMMinimalLocal local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR)
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

// EnvEVMSoak deployment for a long running soak tests
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
