package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

func main() {
	chainlinkChart, err := chainlink.New(0, map[string]interface{}{
		"replicas": 1,
	})
	if err != nil {
		panic(err)
	}
	e := environment.New(&environment.Config{
		NamespacePrefix: "modified-env",
		Labels:          []string{fmt.Sprintf("envType=Modified")},
	}).
		AddChart(blockscout.New(&blockscout.Props{
			WsURL:   "ws://geth:8546",
			HttpURL: "http://geth:8544",
		})).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart)
	err = e.Run()
	if err != nil {
		panic(err)
	}
	e.ClearCharts()
	chainlinkChart2, err := chainlink.New(0, map[string]interface{}{
		"replicas": 1,
	})
	if err != nil {
		panic(err)
	}
	err = e.
		AddChart(blockscout.New(&blockscout.Props{
			HttpURL: "http://geth:9000",
		})).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart2).
		Run()
	if err != nil {
		panic(err)
	}
}
