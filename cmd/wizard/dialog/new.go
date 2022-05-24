package dialog

import (
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
		{Text: chainlink.EnvTypeEVM1, Description: "Create 1 CL Node minimal local env (EVM)"},
		{Text: chainlink.EnvTypeEVM5, Description: "Create 5 CL Nodes OCR environment (EVM)"},
		{Text: chainlink.EnvTypeEVM5BS, Description: "Create 5 CL Nodes OCR environment (EVM) with a Blockscout"},
		{Text: chainlink.EnvTypeEVM5Soak, Description: "Create 5 CL Nodes OCR environment for a long running soak test (EVM)"},
	})
}

func NewEnvDialogue() {
	var choice string
	var ok bool
	if !Ctx.Connect {
		color.Green("Choose an environment type")
		choice = Input(envTypeSuggester)
	} else {
		completer, nsTypesMap := getNamespacesData()
		if nsTypesMap == nil {
			return
		}
		selectedNs := Input(completer)
		if selectedNs == "" {
			color.Red("No environment selected")
			return
		}
		if choice, ok = nsTypesMap[selectedNs]; !ok {
			color.Red("No type found for selected namespace, are labels present on the namespace?")
			return
		}
		// nolint
		os.Setenv("ENV_NAMESPACE", selectedNs)
	}
	// nolint
	defer os.Unsetenv("ENV_NAMESPACE")
	switch choice {
	case chainlink.EnvTypeEVM1:
		if err := presets.EnvEVMOneNode(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMMinimalLocal(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case chainlink.EnvTypeEVM5BS:
		if err := presets.EnvEVMMinimalLocalBS(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case chainlink.EnvTypeEVM5Soak:
		if err := presets.EnvEVMSoak(&environment.Config{
			DryRun: Ctx.DryRun,
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	default:
		color.Red("No environment preset found")
		return
	}
	color.Yellow("Environment is up and connected")
}
