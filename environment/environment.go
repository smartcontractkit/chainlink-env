package environment

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"github.com/smartcontractkit/chainlink-env/logging"
	"github.com/smartcontractkit/chainlink-env/pkg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

const (
	COVERAGE_DIR = "cover"
)

var (
	JSIIGlobalMu       = &sync.Mutex{}
	defaultAnnotations = map[string]*string{"prometheus.io/scrape": a.Str("true")}
)

// ConnectedChart interface to interact both with cdk8s apps and helm charts
type ConnectedChart interface {
	// IsDeploymentNeeded
	// true - we deploy/connect and expose environment data
	// false - we are using external environment, but still exposing data
	IsDeploymentNeeded() bool
	// GetName name of the deployed part
	GetName() string
	// GetPath get Helm chart path, repo or local path
	GetPath() string
	// GetVersion gets the chart's version, empty string if none is specified
	GetVersion() string
	// GetProps get code props if it's typed environment
	GetProps() interface{}
	// GetValues get values.yml props as map, if it's Helm
	GetValues() *map[string]interface{}
	// ExportData export deployment part data in the env
	ExportData(e *Environment) error
}

// Config is an environment common configuration, labels, annotations, connection types, readiness check, etc.
type Config struct {
	// TTL is time to live for the environment, used with kube-janitor
	TTL time.Duration
	// NamespacePrefix is a static namespace prefix
	NamespacePrefix string
	// Namespace is full namespace name
	Namespace string
	// Labels is a set of labels applied to the namespace in a format of "key=value"
	Labels   []string
	nsLabels *map[string]*string
	// Allow deployment to nodes with these olerances
	Tolerations []map[string]string
	// Restrict deployment to only nodes matching a particular node role
	NodeSelector map[string]string
	// ReadyCheckData is settings for readiness probes checks for all deployment components
	// checking that all pods are ready by default with 8 minutes timeout
	//	&client.ReadyCheckData{
	//		ReadinessProbeCheckSelector: "",
	//		Timeout:                     8 * time.Minute,
	//	}
	ReadyCheckData *client.ReadyCheckData
	// DryRun if true, app will just generate a manifest in local dir
	DryRun bool
	// InsideK8s used for long-running soak tests where you connect to env from the inside
	InsideK8s bool
	// KeepConnection keeps connection until interrupted with a signal, useful when prototyping and debugging a new env
	KeepConnection bool
	// RemoveOnInterrupt automatically removes an environment on interrupt
	RemoveOnInterrupt bool
	// UpdateWaitInterval an interval to wait for deployment update started
	UpdateWaitInterval time.Duration
}

func defaultEnvConfig() *Config {
	return &Config{
		TTL:             20 * time.Minute,
		NamespacePrefix: "chainlink-test-env",
		ReadyCheckData: &client.ReadyCheckData{
			ReadinessProbeCheckSelector: "",
			Timeout:                     8 * time.Minute,
		},
	}
}

// Environment describes a launched test environment
type Environment struct {
	App             cdk8s.App
	CurrentManifest string
	root            cdk8s.Chart
	Charts          []ConnectedChart  // All connected charts in the
	Cfg             *Config           // The environment specific config
	Client          *client.K8sClient // Client connecting to the K8s cluster
	Fwd             *client.Forwarder // Used to forward ports from local machine to the K8s cluster
	Artifacts       *Artifacts
	Chaos           *client.Chaos
	httpClient      *resty.Client
	URLs            map[string][]string // General URLs of launched resources. Uses '_local' to delineate forwarded ports
}

// New creates new environment
func New(cfg *Config) *Environment {
	logging.Init()
	if cfg == nil {
		cfg = &Config{}
	}
	targetCfg := defaultEnvConfig()
	config.MustMerge(targetCfg, cfg)
	c := client.NewK8sClient()
	e := &Environment{
		URLs:   make(map[string][]string),
		Charts: make([]ConnectedChart, 0),
		Client: c,
		Cfg:    targetCfg,
		Fwd:    client.NewForwarder(c, targetCfg.KeepConnection),
	}
	ns := os.Getenv(config.EnvVarNamespace)
	if e.Client.NamespaceExists(ns) {
		log.Info().Str("Namespace", ns).Msg("Namespace found")
		e.Cfg.Namespace = ns
	} else {
		e.Cfg.Namespace = fmt.Sprintf("%s-%s", e.Cfg.NamespacePrefix, uuid.NewString()[0:5])
	}
	arts, err := NewArtifacts(e.Client, e.Cfg.Namespace)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create artifacts client")
	}
	e.Artifacts = arts

	JSIIGlobalMu.Lock()
	defer JSIIGlobalMu.Unlock()
	e.initApp()
	k8s.NewKubeNamespace(e.root, a.Str("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        a.Str(e.Cfg.Namespace),
			Labels:      e.Cfg.nsLabels,
			Annotations: &defaultAnnotations,
		},
	})
	e.Chaos = client.NewChaos(c, e.Cfg.Namespace)
	return e
}

func (m *Environment) initApp() {
	var err error
	m.App = cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	m.Cfg.Labels = append(m.Cfg.Labels, "generatedBy=cdk8s")
	m.Cfg.Labels = append(m.Cfg.Labels, fmt.Sprintf("owner=%s", os.Getenv(config.EnvVarUser)))

	if os.Getenv(config.EnvVarCLCommitSha) != "" {
		m.Cfg.Labels = append(m.Cfg.Labels, fmt.Sprintf("commit=%s", os.Getenv(config.EnvVarCLCommitSha)))
	}
	testTrigger := os.Getenv(config.EnvVarTestTrigger)
	if testTrigger == "" {
		testTrigger = "manual"
	}
	m.Cfg.Labels = append(m.Cfg.Labels, fmt.Sprintf("triggered-by=%s", testTrigger))

	if tolerationRole := os.Getenv(config.EnvVarToleration); tolerationRole != "" {
		m.Cfg.Tolerations = []map[string]string{{
			"key":      "node-role",
			"operator": "Equal",
			"value":    tolerationRole,
			"effect":   "NoSchedule",
		}}
	}

	if selectorRole := os.Getenv(config.EnvVarNodeSelector); selectorRole != "" {
		m.Cfg.NodeSelector = map[string]string{
			"node-role": selectorRole,
		}
	}

	m.Cfg.nsLabels, err = a.ConvertLabels(m.Cfg.Labels)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defaultAnnotations[pkg.TTLLabelKey] = a.ShortDur(m.Cfg.TTL)
	m.root = cdk8s.NewChart(m.App, a.Str(fmt.Sprintf("root-chart-%s", m.Cfg.Namespace)), &cdk8s.ChartProps{
		Labels:    m.Cfg.nsLabels,
		Namespace: a.Str(m.Cfg.Namespace),
	})
}

// AddChart adds a chart to the deployment
func (m *Environment) AddChart(f func(root cdk8s.Chart) ConnectedChart) *Environment {
	JSIIGlobalMu.Lock()
	defer JSIIGlobalMu.Unlock()
	m.Charts = append(m.Charts, f(m.root))
	return m
}

func (m *Environment) removeChart(name string) {
	for i, c := range m.Charts {
		if c.GetName() == name {
			m.Charts = append(m.Charts[:i], m.Charts[i+1:]...)
		}
	}
	m.root.Node().TryRemoveChild(a.Str(name))
}

// ModifyHelm modifies helm chart in deployment
func (m *Environment) ModifyHelm(name string, chart ConnectedChart) *Environment {
	JSIIGlobalMu.Lock()
	defer JSIIGlobalMu.Unlock()
	m.removeChart(name)
	if chart.IsDeploymentNeeded() {
		log.Trace().
			Str("Chart", chart.GetName()).
			Str("Path", chart.GetPath()).
			Interface("Props", chart.GetProps()).
			Interface("Values", chart.GetValues()).
			Msg("Chart deployment values")
		cdk8s.NewHelm(m.root, a.Str(name), &cdk8s.HelmProps{
			Chart: a.Str(chart.GetPath()),
			HelmFlags: &[]*string{
				a.Str("--namespace"),
				a.Str(m.Cfg.Namespace),
			},
			ReleaseName: a.Str(name),
			Values:      chart.GetValues(),
		})
	}
	m.Charts = append(m.Charts, chart)
	return m
}

func (m *Environment) AddHelm(chart ConnectedChart) *Environment {
	JSIIGlobalMu.Lock()
	defer JSIIGlobalMu.Unlock()
	if chart.IsDeploymentNeeded() {
		values := &map[string]interface{}{
			"tolerations":  m.Cfg.Tolerations,
			"nodeSelector": m.Cfg.NodeSelector,
		}
		config.MustMerge(values, chart.GetValues())
		log.Trace().
			Str("Chart", chart.GetName()).
			Str("Path", chart.GetPath()).
			Interface("Props", chart.GetProps()).
			Interface("Values", values).
			Msg("Chart deployment values")
		helmFlags := []*string{
			a.Str("--namespace"),
			a.Str(m.Cfg.Namespace),
		}
		if chart.GetVersion() != "" {
			helmFlags = append(helmFlags, a.Str("--version"), a.Str(chart.GetVersion()))
		}
		cdk8s.NewHelm(m.root, a.Str(chart.GetName()), &cdk8s.HelmProps{
			Chart:       a.Str(chart.GetPath()),
			HelmFlags:   &helmFlags,
			ReleaseName: a.Str(chart.GetName()),
			Values:      values,
		})
	}
	m.Charts = append(m.Charts, chart)
	return m
}

// PrintExportData prints export data
func (m *Environment) PrintExportData() error {
	for _, c := range m.Charts {
		err := c.ExportData(m)
		if err != nil {
			return err
		}
	}
	log.Debug().Interface("URLs", m.URLs).Msg("Connection URLs")
	return nil
}

// DumpLogs dumps all logs into a file
func (m *Environment) DumpLogs(path string) error {
	arts, err := NewArtifacts(m.Client, m.Cfg.Namespace)
	if err != nil {
		return err
	}
	if path == "" {
		path = fmt.Sprintf("logs/%s-%d", m.Cfg.Namespace, time.Now().Unix())
	}
	return arts.DumpTestResult(path, "chainlink")
}

// ResourcesSummary returns resources summary for selected pods as a map, used in reports
func (m *Environment) ResourcesSummary(selector string) (map[string]map[string]string, error) {
	pl, err := m.Client.ListPods(m.Cfg.Namespace, selector)
	if err != nil {
		return nil, err
	}
	if len(pl.Items) == 0 {
		return nil, errors.Errorf("no pods found for selector: %s", selector)
	}
	resources := make(map[string]map[string]string)
	for _, p := range pl.Items {
		for _, c := range p.Spec.Containers {
			if resources[c.Name] == nil {
				resources[c.Name] = make(map[string]string)
			}
			cpuRes := c.Resources.Requests["cpu"]
			resources[c.Name]["cpu"] = cpuRes.String()
			memRes := c.Resources.Requests["memory"]
			resources[c.Name]["memory"] = memRes.String()
		}
	}
	return resources, nil
}

// ClearCharts recreates cdk8s app
func (m *Environment) ClearCharts() {
	m.Charts = make([]ConnectedChart, 0)
	m.initApp()
}

func (m *Environment) Manifest() string {
	return m.CurrentManifest
}

// Run deploys or connects to already created environment
func (m *Environment) Run() error {
	JSIIGlobalMu.Lock()
	m.CurrentManifest = m.App.SynthYaml().(string)
	JSIIGlobalMu.Unlock()
	if !m.Cfg.InsideK8s {
		if err := m.Deploy(m.CurrentManifest); err != nil {
			log.Error().Err(err).Msg("Error deploying environment")
			_ = m.Shutdown()
			return err
		}
	}
	if m.Cfg.DryRun {
		log.Info().Msg("Dry-run mode, manifest synthesized and saved as tmp-manifest.yaml")
		return nil
	}
	if err := m.Fwd.Connect(m.Cfg.Namespace, "", m.Cfg.InsideK8s); err != nil {
		return err
	}
	log.Debug().Interface("Ports", m.Fwd.Info).Msg("Forwarded ports")
	if err := m.PrintExportData(); err != nil {
		return err
	}
	if len(m.URLs["goc"]) != 0 {
		m.httpClient = resty.New().SetBaseURL(m.URLs["goc"][0])
	}
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
	if err := m.Client.Apply(manifest); err != nil {
		return err
	}
	if int64(m.Cfg.UpdateWaitInterval) != 0 {
		time.Sleep(m.Cfg.UpdateWaitInterval)
	}
	if err := m.Client.CheckReady(m.Cfg.Namespace, m.Cfg.ReadyCheckData); err != nil {
		return err
	}
	return m.enumerateApps()
}

type CoverageProfileParams struct {
	Force             bool     `form:"force" json:"force"`
	Service           []string `form:"service" json:"service"`
	Address           []string `form:"address" json:"address"`
	CoverFilePatterns []string `form:"coverfile" json:"coverfile"`
	SkipFilePatterns  []string `form:"skipfile" json:"skipfile"`
}

func (m *Environment) getCoverageList() (map[string]interface{}, error) {
	var servicesMap map[string]interface{}
	resp, err := m.httpClient.R().
		SetResult(&servicesMap).
		Get("v1/cover/list")
	if err != nil {
		return nil, err
	}
	if resp.Status() != "200 OK" {
		return nil, errors.New("coverage service list request is not 200")
	}
	return servicesMap, nil
}

func (m *Environment) ClearCoverage() error {
	servicesMap, err := m.getCoverageList()
	if err != nil {
		return err
	}
	for serviceName := range servicesMap {
		r, err := m.httpClient.R().
			SetBody(CoverageProfileParams{Service: []string{serviceName}}).
			Post("v1/cover/clear")
		if err != nil {
			return err
		}
		if r.Status() != "200 OK" {
			return errors.New("coverage service list request is not 200")
		}
		log.Debug().Str("Service", serviceName).Msg("Coverage cleared")
	}
	return nil
}

func (m *Environment) SaveCoverage() error {
	if err := MkdirIfNotExists(COVERAGE_DIR); err != nil {
		return err
	}
	servicesMap, err := m.getCoverageList()
	if err != nil {
		return err
	}
	log.Debug().Interface("Services", servicesMap).Msg("Services eligible for coverage")
	for serviceName := range servicesMap {
		r, err := m.httpClient.R().
			SetBody(CoverageProfileParams{Service: []string{serviceName}}).
			Post("v1/cover/profile")
		if err != nil {
			return err
		}
		if r.Status() != "200 OK" {
			return errors.New("coverage service list request is not 200")
		}
		log.Debug().Str("Service", serviceName).Msg("Coverage received")
		if err := os.WriteFile(fmt.Sprintf("%s/%s.cov", COVERAGE_DIR, serviceName), r.Body(), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown environment, remove namespace
func (m *Environment) Shutdown() error {
	return m.Client.RemoveNamespace(m.Cfg.Namespace)
}
