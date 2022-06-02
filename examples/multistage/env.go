package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/geth"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	// example of quick usage to debug env, removed on SIGINT
	e := environment.New(&environment.Config{
		Labels: []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5)},
	})
	err := e.
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		Run()
	if err != nil {
		panic(err)
	}
	e.Cfg.KeepConnection = true
	e.Cfg.RemoveOnInterrupt = true
	err = e.
		AddChart(blockscout.New(&blockscout.Props{})).
		AddHelm(geth.New(nil)).
		AddHelm(chainlink.New(nil)).
		Run()
	if err != nil {
		panic(err)
	}
}
