package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/remotetestrunner"
)

func main() {
	// example of quick usage to debug env, removed on SIGINT
	err := environment.New(&environment.Config{
		Labels:         []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5RemoteRunner)},
		KeepConnection: true,
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(remotetestrunner.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		Run()
	if err != nil {
		panic(err)
	}
}
