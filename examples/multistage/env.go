package main

import (
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	// example of quick usage to debug env, removed on SIGINT
	e := environment.New(&environment.Config{
		Labels: []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5)}, // set more additional labels
	})
	err := e.
		AddHelm(mockservercfg.New(nil)). // add more Helm charts, all charts got merged in a manifest and deployed with kubectl
		AddHelm(mockserver.New(nil)).
		Run()
	if err != nil {
		panic(err)
	}
	// do some other stuff with deployed charts
	time.Sleep(5 * time.Second)
	err = e.
		AddChart(blockscout.New(&blockscout.Props{})). // you can also add cdk8s charts if you like Go code
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		Run()
	if err != nil {
		panic(err)
	}
}
