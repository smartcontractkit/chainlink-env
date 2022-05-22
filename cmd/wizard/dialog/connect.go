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
	color.Yellow("Searching for environments..")
	c := client.NewK8sClient()
	nss, err := c.ListNamespaces(fmt.Sprintf("%s=%s", chainlink.ControlLabelKey, chainlink.ControlLabelValue))
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	if len(nss.Items) == 0 {
		color.Red("No suitable environments found")
		return nil, nil
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
