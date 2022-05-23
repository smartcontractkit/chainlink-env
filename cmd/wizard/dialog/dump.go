package dialog

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
	"time"
)

func NewDumpDialogue() {
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
	a, err := environment.NewArtifacts(c, selectedNs)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	logDirName := fmt.Sprintf("logs/%s-%d", selectedNs, time.Now().Unix())
	if err = a.DumpTestResult(logDirName, "chainlink"); err != nil {
		log.Fatal().Err(err).Send()
	}
}
