package main

import (
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

func main() {
	chainlinkChart, err := chainlink.New(0, nil)
	if err != nil {
		panic(err)
	}
	e := environment.New(nil).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart)
	if err := e.Run(); err != nil {
		panic(err)
	}
	if err := e.DumpLogs("logs/mytest"); err != nil {
		panic(err)
	}
}
