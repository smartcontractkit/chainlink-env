package dialog

import (
	"github.com/fatih/color"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/config"
	"os"
)

func NewMultiNetworkOptsDialogue() *presets.MultiNetworkOpts {
	color.Yellow("Please provide a path to your networks configuration file:")
	path := Input(defaultCompleter(nil))
	// nolint
	os.Setenv(config.EnvVarNetworksConfigFile, path)
	var opts presets.MultiNetworkOpts
	config.MustEnvOverrideStruct("", &opts)
	return &opts
}
