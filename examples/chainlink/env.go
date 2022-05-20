package main

import (
	"github.com/smartcontractkit/chainlink-env/chainlink"
	"github.com/smartcontractkit/chainlink-env/chains/ethereum"
	"github.com/smartcontractkit/chainlink-env/environment"
)

func main() {
	err := environment.New(&environment.Config{
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).DeployOrConnect(
		chainlink.NewChart(
			&chainlink.Props{
				Namespace: "zclcdk-deployment",
				ChainProps: []interface{}{
					&ethereum.Props{
						DevPeriod: "2",
					},
				},
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
