package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"time"
)

func main() {
	e := environment.New(&environment.Config{
		NamespacePrefix: "modified-env",
		Labels:          []string{fmt.Sprintf("envType=Modified")},
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		}))
	err := e.Run()
	if err != nil {
		panic(err)
	}
	err = e.
		ModifyHelm("chainlink-0", chainlink.New(0, map[string]interface{}{
			"replicas": 2,
		})).Run()
	// nolint
	defer e.Shutdown()
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Minute)
}
