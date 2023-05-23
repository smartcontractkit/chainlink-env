package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
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
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart)
	err = e.Run()
	if err != nil {
		panic(err)
	}
	e.Cfg.KeepConnection = true
	e.Cfg.RemoveOnInterrupt = true
	chainlinkChart2, err := chainlink.New(0, map[string]interface{}{
		"replicas": 2,
	})
	if err != nil {
		panic(err)
	}
	err = e.
		ModifyHelm("chainlink-0", chainlinkChart2).Run()
	if err != nil {
		panic(err)
	}
}
