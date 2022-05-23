package dialog

import (
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
)

func envTypeSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: chainlink.EnvTypeEVM1, Description: "Create 1 CL Node minimal local env (EVM)"},
		{Text: chainlink.EnvTypeEVM5, Description: "Create 5 CL Nodes OCR environment (EVM)"},
		{Text: chainlink.EnvTypeEVM5Soak, Description: "Create 5 CL Nodes OCR environment for a long running soak test (EVM)"},
	})
}

func NewEnvDialogue() {
	color.Green("Choose environment type")
	choice := Input(envTypeSuggester)
	switch choice {
	case chainlink.EnvTypeEVM1:
		if err := presets.EnvEVMOneNode(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
		return
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMMinimalLocal(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
		return
	case chainlink.EnvTypeEVM5Soak:
		if err := presets.EnvEVMSoak(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
		return
	case chainlink.EnvTypeSolana5:
	default:
		color.Red("No environment preset found")
		return
	}
	color.Yellow("Environment is up and connected")
}
