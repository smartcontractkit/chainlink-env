package dialog

import (
	prompt "github.com/c-bata/go-prompt"
	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
	"syscall"
)

const (
	PromptHeader = "> "
)

// this kind of magic is needed to handle SIGINT inside original chainlink-env process first
var FD int
var OriginalTermios *unix.Termios

// SaveInitialTTY might be useful later, remove after all features are done
func SaveInitialTTY() {
	var err error
	FD, err = syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	// get the original settings
	OriginalTermios, err = termios.Tcgetattr(uintptr(FD))
	if err != nil {
		panic(err)
	}
}

// signalAwareEnv might be useful later, remove after all features are done
//func signalAwareEnv(taskFunc func(config *environment.Config) error, config *environment.Config) error {
//	// restore the original settings to allow ctrl-c to generate signal
//	if err := termios.Tcsetattr(uintptr(FD), termios.TCSANOW, OriginalTermios); err != nil {
//		panic(err)
//	}
//	ctx, cancel := context.WithCancel(context.Background())
//	c := make(chan os.Signal, 1)
//	signal.Notify(c,
//		syscall.SIGHUP,
//		syscall.SIGINT,
//		syscall.SIGTERM,
//		syscall.SIGQUIT)
//	errCh := make(chan error)
//
//	go func() {
//		select {
//		case <-c:
//			cancel()
//		}
//	}()
//	go func() {
//		defer cancel()
//		if err := taskFunc(config); err != nil {
//			errCh <- err
//		}
//	}()
//	select {
//	case <-ctx.Done():
//		return nil
//	case err := <-errCh:
//		return err
//	}
//}

func Input(suggester prompt.Completer) string {
	return prompt.Input(PromptHeader, suggester, prompt.OptionInputTextColor(prompt.Green))
}

func defaultSuggester(d prompt.Document, s []prompt.Suggest) []prompt.Suggest {
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func defaultCompleter(s []prompt.Suggest) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		return defaultSuggester(d, s)
	}
}
