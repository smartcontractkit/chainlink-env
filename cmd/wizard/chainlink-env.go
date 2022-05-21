package main

import (
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/dialog"
)

func main() {
	dialog.SaveInitialTTY()
	dialog.NewInitDialogue()
}
