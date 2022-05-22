package dialog

import (
	prompt "github.com/c-bata/go-prompt"
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
