package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	// example of Chainlink cluster connected to external EVM node
	// you can also override all required vars from ENV like:
	// ETHEREUM_NETWORK_TYPE="external" CL_VALUES="chainlink_overrides.yaml" go run examples/simple/env.go
	err := environment.New(&environment.Config{
		Labels:            []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5External)},
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(&ethereum.Props{
			NetworkType: ethereum.ExternalEthereum,
		})).
		AddHelm(chainlink.New(map[string]interface{}{
			"env": map[string]interface{}{
				"eth_http_url": ethereum.KovanHTTPSURL,
				"eth_url":      ethereum.KovanWSURL,
				"eth_chain_id": "1",
			},
		})).
		Run()
	if err != nil {
		panic(err)
	}
}
