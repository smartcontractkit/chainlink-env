package dialog

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/smartcontractkit/chainlink-env/config"
)

func NewHelpDialogue() {
	fmt.Printf(``)
	color.Green("Chainlink environments wizard helps you configure chainlink environment for tests!")
	color.Green("You can use env variables before running:")
	color.Yellow(fmt.Sprintf("%s=\"%s\"\t\t%s", config.EnvVarUser, config.EnvVarUserExample, config.EnvVarUserDescription))
	color.Yellow(fmt.Sprintf("%s=\"%s\"\t\t%s", config.EnvVarCLImage, config.EnvVarCLImageExample, config.EnvVarCLImageDescription))
	color.Yellow(fmt.Sprintf("%s=\"%s\"\t\t%s", config.EnvVarCLTag, config.EnvVarCLTagExample, config.EnvVarCLTagDescription))
	color.Yellow(fmt.Sprintf("%s=\"%s\"\t\t%s", config.EnvVarLogLevel, config.EnvVarLogLevelExample, config.EnvVarLogLevelDescription))
	color.Yellow(fmt.Sprintf("%s=\"%s\"\t\t%s", config.EnvVarNetworksConfigFile, config.EnvVarNetworksConfigFileExample, config.EnvVarNetworksConfigFileDescription))
}
