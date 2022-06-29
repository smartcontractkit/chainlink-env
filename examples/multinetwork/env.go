package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
)

func main() {
	// use in code or override like
	// EXTERNAL_NETWORK_HTTP_URL="..." EXTERNAL_NETWORK_WS_URL="..." EXTERNAL_NETWORK_CHAIN_ID="..." go run examples/multinetwork/env.go
	err := presets.MultiNetwork(&environment.Config{
		Labels:            []string{fmt.Sprintf("envType=%s", pkg.EnvTypeMultinetwork)},
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}, &presets.MultiNetworkOpts{
		Networks: presets.Networks{
			{
				Name:      "Testnet",
				Simulated: true,
				HttpURLs:  []string{"http://geth:8544"},
				WsURLs:    []string{"ws://geth:8546"},
				ChainID:   "1337",
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
