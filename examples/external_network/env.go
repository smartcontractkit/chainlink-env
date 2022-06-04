package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
)

func main() {
	// use in code or override like
	// EXTERNAL_NETWORK_HTTP_URL="..." EXTERNAL_NETWORK_WS_URL="..." EXTERNAL_NETWORK_CHAIN_ID="..." go run examples/external_network/env.go
	err := presets.EVMExternal(&environment.Config{
		Labels:            []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5External)},
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}, &presets.ExternalNetworkOpts{
		HttpURL: "http or https url",
		WsURL:   "ws or wss url",
		ChainID: "1",
	})
	if err != nil {
		panic(err)
	}
}
