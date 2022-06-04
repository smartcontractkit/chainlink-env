package main

import (
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/dialog"
	"github.com/smartcontractkit/chainlink-env/logging"
)

func main() {
	logging.Init()
	dialog.NewInitDialogue()
}
