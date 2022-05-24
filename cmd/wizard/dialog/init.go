package dialog

import (
	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

// Context wizard global context settings
type Context struct {
	DryRun  bool
	Connect bool
}

var Ctx = &Context{}

const (
	OptionNew     = "new"
	OptionDryRun  = "dry-run"
	OptionConnect = "connect"
	OptionDump    = "dump"
	OptionRemove  = "remove"
	OptionQuit    = "quit"
	OptionNone    = ""
)

func rootSuggester(d prompt.Document) []prompt.Suggest {
	return defaultSuggester(d, []prompt.Suggest{
		{Text: OptionNew, Description: "Create new environment"},
		{Text: OptionDryRun, Description: "Generate environment manifest and save in tmp-manifest.yaml"},
		{Text: OptionConnect, Description: "Connect to already created environment"},
		{Text: OptionDump, Description: "Dump environment logs to a dir"},
		{Text: OptionRemove, Description: "Remove an environment"},
		{Text: OptionQuit, Description: "Exit application"},
	})
}

func NewInitDialogue() {
	for {
		var choice string
		color.Green("Chainlink environments wizard")
		choice = Input(rootSuggester)
		switch choice {
		case OptionNew:
			NewEnvDialogue()
		case OptionDryRun:
			Ctx.DryRun = true
			NewEnvDialogue()
			Ctx.DryRun = false
		case OptionConnect:
			Ctx.Connect = true
			NewEnvDialogue()
			Ctx.Connect = false
		case OptionDump:
			NewDumpDialogue()
		case OptionRemove:
			NewRemoveDialogue()
		case OptionQuit:
			fallthrough
		case OptionNone:
			return
		default:
			color.Red("no such option, please choose what is available")
		}
	}
}
