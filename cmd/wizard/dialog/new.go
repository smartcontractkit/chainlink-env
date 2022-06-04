package dialog

import (
	"fmt"
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"os"
)

func envTypeSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: pkg.EnvTypeEVM1, Description: "Create 1 CL Node minimal local env (EVM)"},
		{Text: pkg.EnvTypeEVM5, Description: "Create 5 CL Nodes environment (EVM)"},
		{Text: pkg.EnvTypeEVM5External, Description: "Create 5 CL Nodes environment (EVM) with an external network"},
		{Text: pkg.EnvTypeETH5Reorg, Description: "Create 5 CL Nodes environment for Ethereum reorg"},
		{Text: pkg.EnvTypeEVM5BS, Description: "Create 5 CL Nodes environment (EVM) with a Blockscout"},
		{Text: pkg.EnvTypeEVM5Soak, Description: "Create 5 CL Nodes environment for a long running soak test (EVM)"},
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
	case pkg.EnvTypeEVM1:
		if err := presets.EVMOneNode(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case pkg.EnvTypeEVM5:
		if err := presets.EVMMinimalLocal(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case pkg.EnvTypeEVM5External:
		answers := NewExternalOptsDialogue()
		if err := presets.EVMExternal(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}, answers); err != nil {
			log.Fatal().Err(err).Send()
		}
	case pkg.EnvTypeETH5Reorg:
		if err := presets.EVMReorg(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case pkg.EnvTypeEVM5BS:
		if err := presets.EVMMinimalLocalBS(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	case pkg.EnvTypeEVM5Soak:
		if err := presets.EVMSoak(&environment.Config{
			DryRun: Ctx.DryRun,
			Labels: []string{fmt.Sprintf("envType=%s", choice)},
		}); err != nil {
			log.Fatal().Err(err).Send()
		}
	default:
		color.Red("No environment preset found, env must have envType=... label of a known environment")
		return
	}
	color.Yellow("Environment is up and connected")
}
