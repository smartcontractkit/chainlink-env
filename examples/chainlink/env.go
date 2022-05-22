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
				TTL:       2 * time.Minute,
				Labels:    []string{fmt.Sprintf("envType=%s", chainlink.EnvTypeEVM5)},
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
	if err != nil {
		panic(err)
	}
}
