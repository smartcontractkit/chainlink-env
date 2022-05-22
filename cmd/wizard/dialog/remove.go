package dialog

import (
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
)

func NewRemoveDialogue() {
	completer, envNameToType := getNamespacesData()
	if envNameToType == nil {
		return
	}
	selectedNs := Input(completer)
	if selectedNs == "" {
		color.Red("No environment selected")
		return
	}
	c := client.NewK8sClient()
	if err := c.RemoveNamespace(selectedNs); err != nil {
		log.Fatal().Err(err).Send()
	}
}
