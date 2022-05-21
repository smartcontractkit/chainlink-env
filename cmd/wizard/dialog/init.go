package dialog

import (
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

// Context wizard global context settings
type Context struct {
	DryRun bool
}

var Ctx = &Context{}

const (
	OptionNew     = "new"
	OptionDryRun  = "dry-run"
	OptionConnect = "connect"
	OptionDump    = "dump"
	OptionQuit    = "quit"
)

func rootSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: OptionNew, Description: "Create new environment"},
		{Text: OptionDryRun, Description: "Generate environment manifest and save in tmp-manifest.yaml"},
		{Text: OptionConnect, Description: "Connect to already created environment"},
		{Text: OptionDump, Description: "Dump environment logs to a dir"},
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
	case OptionDryRun:
		Ctx.DryRun = true
		NewEnvDialogue()
	case OptionConnect:
		NewConnectDialogue()
	case OptionDump:
		NewDumpDialogue()
	case OptionQuit:
		color.Green("terminating process, bye!")
	default:
		color.Red("no such option, please choose what is available")
	}
}
