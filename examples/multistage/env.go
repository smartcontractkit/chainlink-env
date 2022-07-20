package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	e := environment.New(nil)
	err := e.
		AddChart(blockscout.New(&blockscout.Props{})). // you can also add cdk8s charts if you like Go code
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		Run()
	if err != nil {
		panic(err)
	}
	// do some other stuff with deployed charts
	pl, err := e.Client.ListPods(e.Cfg.Namespace, "app=chainlink-0")
	if err != nil {
		panic(err)
	}
	dstPath := fmt.Sprintf("%s/%s:/", e.Cfg.Namespace, pl.Items[0].Name)
	if _, _, _, err = e.Client.CopyToPod(e.Cfg.Namespace, "./examples/multistage/someData.txt", dstPath, "node"); err != nil {
		panic(err)
	}
	// deploy another part
	err = e.
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		Run()
	// nolint
	defer e.Shutdown()
	if err != nil {
		panic(err)
	}
}
