package dialog

import (
	"encoding/json"
	"fmt"
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/pkg/chainlink"
	"os"
)

func getNamespacesData() (prompt.Completer, map[string]string) {
	c := client.NewK8sClient()
	nss, err := c.ListNamespaces(fmt.Sprintf("%s=%s", chainlink.ControlLabelKey, chainlink.ControlLabelValue))
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	sug := make([]prompt.Suggest, 0)
	envNameToType := make(map[string]string)
	for _, ns := range nss.Items {
		labels, _ := json.Marshal(ns.Labels)
		envNameToType[ns.Name] = ns.Labels[chainlink.ControlLabelEnvTypeKey]
		sug = append(sug, prompt.Suggest{
			Text:        ns.Name,
			Description: string(labels),
		})
	}
	if len(envNameToType) == 0 {
		color.Red("No chainlink-env environments found")
		NewInitDialogue()
	}
	color.Green("Found environments, use autocomplete to select")
	return defaultCompleter(sug), envNameToType
}

func NewConnectDialogue() {
	color.Yellow("Searching for environments..")
	completer, nsTypesMap := getNamespacesData()
	selectedNs := Input(completer)
	// nolint
	os.Setenv("ENV_NAMESPACE", selectedNs)
	selectedType := nsTypesMap[selectedNs]
	switch selectedType {
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
	NewInitDialogue()
}
