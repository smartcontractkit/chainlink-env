package dialog

import (
	"fmt"
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"os"
)

func envTypeSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: chainlink.EnvTypeEVM5, Description: "Create 5 CL Nodes OCR environment (EVM)"},
		{Text: chainlink.EnvTypeEVM5Soak, Description: "Create 5 CL Nodes OCR environment for a long running soak test (EVM)"},
	})
}

func NewEnvDialogue() {
	// nolint
	os.Unsetenv("ENV_NAMESPACE")
	color.Green("Choose environment type")
	choice := Input(envTypeSuggester)
	switch choice {
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMMinimalLocal(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
		color.Yellow("Environment is up and connected")
	case chainlink.EnvTypeEVM5Soak:
		if err := presets.EnvEVMSoak(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case chainlink.EnvTypeSolana5:
	default:
		fmt.Println("no environment preset found")
	}
	NewInitDialogue()
}
