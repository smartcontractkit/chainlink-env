package environment

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

type Config struct {
	KeepConnection    bool
	RemoveOnInterrupt bool
}

type Environment struct {
	Cfg    *Config
	Client *client.K8sClient
	Fwd    *client.Forwarder
	Out    client.ManifestOutput
	URLs   map[string][]string
}

func New(cfg *Config) *Environment {
	c := client.NewK8sClient()
	return &Environment{
		Client: c,
		Cfg:    cfg,
		Fwd:    client.NewForwarder(c, cfg.KeepConnection),
	}
}

func (m *Environment) DeployOrConnect(app cdk8s.App, out client.ManifestOutput) error {
	ns := os.Getenv("ENV_NAMESPACE")
	if !m.Client.NamespaceExists(ns) {
		if err := m.Deploy(app, out); err != nil {
			return err
		}
	} else {
		log.Info().Str("Namespace", ns).Msg("Namespace found")
		out.SetNamespace(ns)
	}
	if err := m.Fwd.Connect(out.GetNamespace(), ""); err != nil {
		return err
	}
	log.Debug().Interface("Mapping", m.Fwd.Info).Msg("Ports mapping")
	var err error
	m.URLs, err = out.ProcessConnections(m.Fwd)
	if err != nil {
		return err
	}
	m.Out = out
	log.Warn().Interface("Connections", m.URLs).Send()
	if m.Cfg.KeepConnection {
		log.Info().Msg("Keeping forwarder connections, press Ctrl+C to interrupt")
		if m.Cfg.RemoveOnInterrupt {
			log.Warn().Msg("Environment will be removed on interrupt")
		}
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Warn().Msg("Interrupted")
		if m.Cfg.RemoveOnInterrupt {
			return m.Client.RemoveNamespace(out.GetNamespace())
		}
	}
	return nil
}

func (m *Environment) Deploy(app cdk8s.App, c client.ManifestOutput) error {
	manifest := app.SynthYaml().(string)
	if err := m.Client.Create(manifest); err != nil {
		return err
	}
	return m.Client.CheckReady(c)
}
