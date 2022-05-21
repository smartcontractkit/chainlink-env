package dialog

import (
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

const (
	OptionNew     = "new"
	OptionConnect = "connect"
	OptionQuit    = "quit"
)

func rootSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: OptionNew, Description: "Create new environment"},
		{Text: OptionConnect, Description: "Connect to already created environment"},
		{Text: OptionQuit, Description: "Exit application"},
	})
}

func NewInitDialogue() {
	var choice string
	color.Green("Chainlink interactive environments control")
	choice = Input(rootSuggester)
	switch choice {
	case OptionNew:
		NewEnvDialogue()
	case OptionConnect:
		NewConnect()
	case OptionQuit:
		color.Green("terminating process, bye!")
	default:
		color.Red("no such option, please choose what is available")
	}
}
