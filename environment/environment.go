package environment

import (
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.TraceLevel)
}

type Config struct {
	DryRun            bool
	KeepConnection    bool
	RemoveOnInterrupt bool
}

type Environment struct {
	Cfg       *Config
	Client    *client.K8sClient
	Artifacts *Artifacts
	Fwd       *client.Forwarder
	Out       client.ManifestOutput
	URLs      map[string][]string
}

// New creates new environment
func New(cfg *Config) *Environment {
	if cfg == nil {
		cfg = &Config{}
	}
	c := client.NewK8sClient()
	e := &Environment{
		Client: c,
		Cfg:    cfg,
		Fwd:    client.NewForwarder(c, cfg.KeepConnection),
	}
	return e
}

// DeployOrConnect deploys or connects to already created environment
func (m *Environment) DeployOrConnect(app cdk8s.App, out client.ManifestOutput) error {
	ns := os.Getenv("ENV_NAMESPACE")
	if !m.Client.NamespaceExists(ns) {
		if err := m.Deploy(app, out); err != nil {
			return m.Shutdown()
		}
	} else {
		log.Info().Str("Namespace", ns).Msg("Namespace found")
		out.SetNamespace(ns)
	}
	if m.Cfg.DryRun {
		log.Info().Msg("Dry-run mode, manifest synthesized and saved as tmp-manifest.yaml")
		return nil
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
	a, err := NewArtifacts(m.Client, m.Out.GetNamespace())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create artifacts client")
	}
	m.Artifacts = a
	if m.Cfg.KeepConnection {
		log.Info().Msg("Keeping forwarder connections, press Ctrl+C to interrupt")
		if m.Cfg.RemoveOnInterrupt {
			log.Warn().Msg("Environment will be removed on interrupt")
		}
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Warn().Msg("Interrupted")
		if m.Cfg.RemoveOnInterrupt {
			return m.Client.RemoveNamespace(out.GetNamespace())
		}
	}
	return nil
}

// Deploy deploy synthesized manifest and check logs for readiness
func (m *Environment) Deploy(app cdk8s.App, c client.ManifestOutput) error {
	manifest := app.SynthYaml().(string)
	log.Info().Str("Namespace", c.GetNamespace()).Msg("Deploying namespace")
	if m.Cfg.DryRun {
		if err := m.Client.DryRun(manifest); err != nil {
			return err
		}
		return nil
	}
	if err := m.Client.Create(manifest); err != nil {
		return err
	}
	return m.Client.CheckReady(c)
}

// Shutdown shutdown environment, remove namespace
func (m *Environment) Shutdown() error {
	return m.Client.RemoveNamespace(m.Out.GetNamespace())
}
