package dialog

import (
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/chainlink"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/environment"
	"os"
)

func NewConnect() {
	color.Yellow("Searching for environments..")
	c := client.NewK8sClient()
	nss, err := c.ListNamespaces(fmt.Sprintf("%s=%s", chainlink.ControlLabelKey, chainlink.ControlLabelValue))
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	namespaces := make([]string, 0)
	sug := make([]prompt.Suggest, 0)
	envNameToType := make(map[string]string)
	for _, ns := range nss.Items {
		namespaces = append(namespaces, ns.Name)
		labels, _ := json.Marshal(ns.Labels)
		envNameToType[ns.Name] = ns.Labels[chainlink.ControlLabelEnvTypeKey]
		sug = append(sug, prompt.Suggest{
			Text:        ns.Name,
			Description: string(labels),
		})
	}
	color.Green("Found environments, use autocomplete to select")
	selectedNs := Input(defaultCompleter(sug))
	os.Setenv("ENV_NAMESPACE", selectedNs)
	selectedType := envNameToType[selectedNs]
	switch selectedType {
	case chainlink.EnvTypeEVM5:
		if err := presets.EnvEVMDefault(&environment.Config{}); err != nil {
			log.Fatal().Err(err).Send()
		}
		color.Yellow("Environment is up and connected")
	default:
		fmt.Printf("not a valid type, please select from suggested")
	}
	NewInitDialogue()
}
