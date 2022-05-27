package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/chains/ethereum"
	"time"
)

func main() {
	// example of quick usage to debug env, removed on SIGINT
	err := environment.New(&environment.Config{
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).DeployOrConnect(
		chainlink.NewChart(
			&chainlink.Props{
				Namespace: "chainlink-env",
				// you can set TTL if you are using https://codeberg.org/hjacobs/kube-janitor
				TTL: 12 * time.Hour,
				// envType field is required to properly connect the environment
				Labels: []string{fmt.Sprintf("envType=%s", chainlink.EnvTypeEVM5)},
				// additional chains can be deployed and connected, props can be overridden using default method
				ChainProps: []interface{}{
					&ethereum.ReorgProps{},
				},
				// almost all vars can be overridden in order ENV_VARS -> Code -> Code defaults
				// see config package for more examples
				AppVersions: []chainlink.VersionProps{
					{
						Image:     "public.ecr.aws/chainlink/chainlink",
						Tag:       "1.4.1-root",
						Instances: 1,
					},
				},
			}))
	if err != nil {
		panic(err)
	}
}
