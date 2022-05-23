package dialog

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"os"
)

func NewConnectDialogue() {
	completer, nsTypesMap := getNamespacesData()
	if nsTypesMap == nil {
		return
	}
	selectedNs := Input(completer)
	if selectedNs == "" {
		color.Red("No environment selected")
		return
	}
	// nolint
	os.Setenv("ENV_NAMESPACE", selectedNs)
	selectedType := nsTypesMap[selectedNs]
	switch selectedType {
	case chainlink.EnvTypeEVM1:
		if err := presets.EnvEVMOneNode(nil); err != nil {
			log.Fatal().Err(err).Send()
		}
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMMinimalLocal(nil); err != nil {
			log.Fatal().Err(err).Send()
		}
		color.Yellow("Environment is up and connected")
	case chainlink.EnvTypeEVM5Soak:
		if err := presets.EnvEVMSoak(nil); err != nil {
			log.Fatal().Err(err).Send()
		}
		color.Yellow("Environment is up and connected")
	default:
		fmt.Printf("not a valid type, please select from suggested")
	}
	// nolint
	os.Unsetenv("ENV_NAMESPACE")
}
