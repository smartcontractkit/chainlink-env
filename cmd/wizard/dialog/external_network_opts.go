package dialog

import (
	"github.com/fatih/color"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
)

func NewExternalOptsDialogue() *presets.ExternalNetworkOpts {
	color.Yellow("Enter HTTP or HTTPS URL for EVM network:")
	httpURL := Input(defaultCompleter(nil))
	color.Yellow("Enter WS or WSS URL for EVM network:")
	wsURL := Input(defaultCompleter(nil))
	color.Yellow("Enter chain id for EVM network:")
	chainID := Input(defaultCompleter(nil))
	return &presets.ExternalNetworkOpts{
		HttpURL: httpURL,
		WsURL:   wsURL,
		ChainID: chainID,
	}
}
