package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"os"
)

func main() {
	// when connecting to an env we need to construct manifest the same way
	os.Setenv("NAMESPACE_NAME", "chainlink-test-env-8a4a4")
	e := environment.New(&environment.Config{
		NamespacePrefix: "modified-env",
		Labels:          []string{fmt.Sprintf("envType=Modified")},
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 2,
		}))
	err := e.Run()
	if err != nil {
		panic(err)
	}
}
