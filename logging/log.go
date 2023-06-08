package logging

import (
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/config"
)

// People often call this multiple times
var loggingMu sync.Mutex

func Init(t *testing.T) zerolog.Logger {
	loggingMu.Lock()
	defer loggingMu.Unlock()
	lvlStr := os.Getenv(config.EnvVarLogLevel)
	if lvlStr == "" {
		lvlStr = "info"
	}
	lvl, err := zerolog.ParseLevel(lvlStr)
	if err != nil {
		panic(err)
	}

	// use the test logger if t is set
	if t != nil {
		log.Logger = zerolog.New(zerolog.NewTestWriter(t)).Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(lvl)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(lvl)
	}
	return log.Logger
}
