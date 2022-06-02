package dialog

import (
	"encoding/json"
	"fmt"
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/pkg"
)

const (
	PromptHeader = ">> "
)

func Input(suggester prompt.Completer) string {
	return prompt.Input(
		PromptHeader,
		suggester,
		prompt.OptionInputTextColor(prompt.DarkGreen),
		prompt.OptionSelectedSuggestionBGColor(prompt.Black),
		prompt.OptionSelectedSuggestionTextColor(prompt.DarkGreen),
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.DarkGreen),
	)
}

func defaultSuggester(d prompt.Document, s []prompt.Suggest) []prompt.Suggest {
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func defaultCompleter(s []prompt.Suggest) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		return defaultSuggester(d, s)
	}
}

func getNamespacesData() (prompt.Completer, map[string]string) {
	color.Yellow("Searching for environments..")
	c := client.NewK8sClient()
	nss, err := c.ListNamespaces(fmt.Sprintf("%s=%s", pkg.ControlLabelKey, pkg.ControlLabelValue))
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
		envNameToType[ns.Name] = ns.Labels[pkg.ControlLabelEnvTypeKey]
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
