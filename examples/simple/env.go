package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/geth"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	// example of quick usage to debug env, removed on SIGINT
	err := environment.New(&environment.Config{
		Labels:         []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5)},
		KeepConnection: true,
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(geth.New(nil)).
		AddHelm(chainlink.New(nil)).
		Run()
	if err != nil {
		panic(err)
	}
}
