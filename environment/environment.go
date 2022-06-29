package environment

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"github.com/smartcontractkit/chainlink-env/logging"
	"github.com/smartcontractkit/chainlink-env/pkg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

var (
	defaultAnnotations = map[string]*string{"prometheus.io/scrape": a.Str("true")}
)

// ConnectedChart interface to interact both with cdk8s apps and helm charts
type ConnectedChart interface {
	IsDeploymentNeeded() bool
	GetName() string
	GetPath() string
	GetProps() interface{}
	GetValues() *map[string]interface{}
	ExportData(e *Environment) error
}

// Config environment configuration
type Config struct {
	TTL               time.Duration
	NamespacePrefix   string
	Namespace         string
	Labels            []string
	NSLabels          *map[string]*string
	ReadyCheckData    *client.ReadyCheckData
	DryRun            bool
	InsideK8s         bool
	KeepConnection    bool
	RemoveOnInterrupt bool
}

func defaultEnvConfig() *Config {
	return &Config{
		TTL:             20 * time.Minute,
		NamespacePrefix: "chainlink-test-env",
		ReadyCheckData: &client.ReadyCheckData{
			ReadinessProbeCheckSelector: "",
			Timeout:                     4 * time.Minute,
		},
	}
}

// Environment describes a launched test environment
type Environment struct {
	App       cdk8s.App
	root      cdk8s.Chart
	Charts    []ConnectedChart  // All connected charts in the
	Cfg       *Config           // The environment specific config
	Client    *client.K8sClient // Client connecting to the K8s cluster
	Fwd       *client.Forwarder // Used to forward ports from local machine to the K8s cluster
	Artifacts *Artifacts
	Chaos     *client.Chaos
	URLs      map[string][]string // General URLs of launched resources. Uses '_local' to delineate forwarded ports
}

// New creates new environment
func New(cfg *Config) *Environment {
	logging.Init()
	if cfg == nil {
		cfg = &Config{}
	}
	targetCfg := defaultEnvConfig()
	config.MustEnvCodeOverrideStruct("ENV_CONFIG", targetCfg, cfg)
	c := client.NewK8sClient()
	e := &Environment{
		URLs:   make(map[string][]string),
		Charts: make([]ConnectedChart, 0),
		Client: c,
		Cfg:    targetCfg,
		Fwd:    client.NewForwarder(c, targetCfg.KeepConnection),
	}
	e.initApp(fmt.Sprintf("%s-%s", e.Cfg.NamespacePrefix, uuid.NewString()[0:5]))
	k8s.NewKubeNamespace(e.root, a.Str("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        a.Str(e.Cfg.Namespace),
			Labels:      e.Cfg.NSLabels,
			Annotations: &defaultAnnotations,
		},
	})
	e.Chaos = client.NewChaos(c, e.Cfg.Namespace)
	return e
}

func (m *Environment) initApp(namespace string) {
	var err error
	m.App = cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	m.Cfg.Namespace = namespace
	m.Cfg.Labels = append(m.Cfg.Labels, "generatedBy=cdk8s")
	m.Cfg.Labels = append(m.Cfg.Labels, fmt.Sprintf("owner=%s", os.Getenv(config.EnvVarUser)))
	if os.Getenv(config.EnvVarCLCommitSha) != "" {
		m.Cfg.Labels = append(m.Cfg.Labels, os.Getenv(config.EnvVarCLCommitSha))
	}

	m.Cfg.NSLabels, err = a.ConvertLabels(m.Cfg.Labels)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defaultAnnotations[pkg.TTLLabelKey] = a.ShortDur(m.Cfg.TTL)
	m.root = cdk8s.NewChart(m.App, a.Str("root-chart"), &cdk8s.ChartProps{
		Labels:    m.Cfg.NSLabels,
		Namespace: a.Str(m.Cfg.Namespace),
	})
}

func (m *Environment) AddChart(f func(root cdk8s.Chart) ConnectedChart) *Environment {
	m.Charts = append(m.Charts, f(m.root))
	return m
}

func (m *Environment) AddHelm(chart ConnectedChart) *Environment {
	if chart.IsDeploymentNeeded() {
		log.Trace().
			Str("Chart", chart.GetName()).
			Str("Path", chart.GetPath()).
			Interface("Props", chart.GetProps()).
			Interface("Values", chart.GetValues()).
			Msg("Chart deployment values")
		cdk8s.NewHelm(m.root, a.Str(chart.GetName()), &cdk8s.HelmProps{
			Chart: a.Str(chart.GetPath()),
			HelmFlags: &[]*string{
				a.Str("--namespace"),
				a.Str(m.Cfg.Namespace),
			},
			ReleaseName: a.Str(chart.GetName()),
			Values:      chart.GetValues(),
		})
	}
	m.Charts = append(m.Charts, chart)
	return m
}

func (m *Environment) PrintURLs() error {
	for _, c := range m.Charts {
		err := c.ExportData(m)
		if err != nil {
			return err
		}
	}
	log.Debug().Interface("URLs", m.URLs).Msg("Connection URLs")
	return nil
}

func (m *Environment) Clear() {
	m.initApp(m.Cfg.Namespace)
}

// Run deploys or connects to already created environment
func (m *Environment) Run() error {
	ns := os.Getenv("ENV_NAMESPACE")
	if !m.Client.NamespaceExists(ns) {
		manifest := m.App.SynthYaml().(string)
		if err := m.Deploy(manifest); err != nil {
			log.Error().Err(err).Msg("Error deploying environment")
			return m.Shutdown()
		}
	} else {
		log.Info().Str("Namespace", ns).Msg("Namespace found")
		m.Cfg.Namespace = ns
	}
	if m.Cfg.DryRun {
		log.Info().Msg("Dry-run mode, manifest synthesized and saved as tmp-manifest.yaml")
		return nil
	}
	if err := m.Fwd.Connect(m.Cfg.Namespace, "", m.Cfg.InsideK8s); err != nil {
		return err
	}
	log.Debug().Interface("Ports", m.Fwd.Info).Msg("Forwarded ports")
	if err := m.PrintURLs(); err != nil {
		return err
	}
	arts, err := NewArtifacts(m.Client, m.Cfg.Namespace)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create artifacts client")
	}
	m.Artifacts = arts
	m.Clear()
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
			return m.Client.RemoveNamespace(m.Cfg.Namespace)
		}
	}
	return nil
}

func (m *Environment) enumerateApps() error {
	apps, err := m.Client.UniqueLabels(m.Cfg.Namespace, "app")
	if err != nil {
		return err
	}
	for _, app := range apps {
		if err := m.Client.EnumerateInstances(m.Cfg.Namespace, fmt.Sprintf("app=%s", app)); err != nil {
			return err
		}
	}
	return nil
}

// Deploy deploy synthesized manifest and check logs for readiness
func (m *Environment) Deploy(manifest string) error {
	log.Info().Str("Namespace", m.Cfg.Namespace).Msg("Deploying namespace")
	if m.Cfg.DryRun {
		if err := m.Client.DryRun(manifest); err != nil {
			return err
		}
		return nil
	}
	if err := m.Client.Create(manifest); err != nil {
		return err
	}
	if err := m.enumerateApps(); err != nil {
		return err
	}
	return m.Client.CheckReady(m.Cfg.Namespace, m.Cfg.ReadyCheckData)
}

// Shutdown environment, remove namespace
func (m *Environment) Shutdown() error {
	return m.Client.RemoveNamespace(m.Cfg.Namespace)
}
