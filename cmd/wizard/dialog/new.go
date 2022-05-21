package dialog

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/chainlink"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
)

func envTypeSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: chainlink.EnvTypeEVM5, Description: "Create 5 CL Nodes OCR environment (EVM)"},
		{Text: chainlink.EnvTypeSolana5, Description: "Create 5 CL Nodes OCR environment (Solana)"},
	})
}

func NewEnvDialogue() {
	color.Green("Choose environment type")
	choice := Input(envTypeSuggester)
	switch choice {
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMDefault(nil); err != nil {
			log.Fatal().Err(err).Send()
		}
		color.Yellow("Environment is up and connected")
	case chainlink.EnvTypeSolana5:
	default:
		fmt.Println("no environment preset found")
	}
	NewInitDialogue()
}
