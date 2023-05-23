package main

import (
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/examples/deployment_part_cdk8s"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

func main() {
	chainlinkChart, err := chainlink.New(0, map[string]interface{}{
		"replicas": 2,
	})
	if err != nil {
		panic(err)
	}
	e := environment.New(nil).
		AddChart(deployment_part_cdk8s.New(&deployment_part_cdk8s.Props{})).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart)
	if err := e.Run(); err != nil {
		panic(err)
	}
	e.Shutdown()
}
