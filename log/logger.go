package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func Default() zerolog.Logger {
	return log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
