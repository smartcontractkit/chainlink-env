package environment

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"github.com/smartcontractkit/chainlink-env/logging"
	"github.com/smartcontractkit/chainlink-env/pkg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

const (
	COVERAGE_DIR       string = "cover"
	FAILED_FUND_RETURN string = "FAILED_FUND_RETURN"
)

const (
	ErrInvalidOCI string = "OCI chart url should be in format oci://$ECR_URL/$ECR_REGISTRY_NAME/$CHART_NAME:[?$CHART_VERSION], was %s"
	ErrOCIPull    string = "failed to pull OCI repo: %s"
)

var (
	defaultAnnotations = map[string]*string{
		"prometheus.io/scrape":                             a.Str("true"),
		"backyards.banzaicloud.io/image-registry-access":   a.Str("true"),
		"backyards.banzaicloud.io/public-dockerhub-access": a.Str("true"),
	}
	defaultPodAnnotations = map[string]string{"cluster-autoscaler.kubernetes.io~1safe-to-evict": "false"}
	defaultPodLabels      = map[string]string{"clenv": "true"}
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
	GetProps() any
	// GetValues get values.yml props as map, if it's Helm
	GetValues() *map[string]any
	// ExportData export deployment part data in the env
	ExportData(e *Environment) error
}

// Config is an environment common configuration, labels, annotations, connection types, readiness check, etc.
type Config struct {
	// TTL is time to live for the environment, used with kube-janitor
	TTL time.Duration
	// JobImage an image to run environment as a job inside k8s
	JobImage string
	// jobDeployed used to limit us to 1 remote runner deploy
	jobDeployed bool
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
	//		Timeout:                     15 * time.Minute,
	//	}
	ReadyCheckData *client.ReadyCheckData
	// DryRun if true, app will just generate a manifest in local dir
	DryRun bool
	// InsideK8s used for long-running soak tests where you connect to env from the inside
	InsideK8s bool
	// NoManifestUpdate is a flag to skip manifest updating when connecting
	NoManifestUpdate bool
	// DetachRunner should we detach the remote runner after starting the test
	detachRunner bool
	// KeepConnection keeps connection until interrupted with a signal, useful when prototyping and debugging a new env
	KeepConnection bool
	// RemoveOnInterrupt automatically removes an environment on interrupt
	RemoveOnInterrupt bool
	// UpdateWaitInterval an interval to wait for deployment update started
	UpdateWaitInterval time.Duration
	// fundReturnFailed the status of a fund return
	fundReturnFailed bool
	// Test the testing library current Test struct
	Test *testing.T
}

func defaultEnvConfig() *Config {
	return &Config{
		TTL:                20 * time.Minute,
		NamespacePrefix:    "chainlink-test-env",
		UpdateWaitInterval: 1 * time.Second,
		ReadyCheckData: &client.ReadyCheckData{
			ReadinessProbeCheckSelector: "",
			Timeout:                     15 * time.Minute,
		},
	}
}

// Environment describes a launched test environment
type Environment struct {
	App                  cdk8s.App
	CurrentManifest      string
	root                 cdk8s.Chart
	Charts               []ConnectedChart  // All connected charts in the
	Cfg                  *Config           // The environment specific config
	Client               *client.K8sClient // Client connecting to the K8s cluster
	Fwd                  *client.Forwarder // Used to forward ports from local machine to the K8s cluster
	Artifacts            *Artifacts
	Chaos                *client.Chaos
	httpClient           *resty.Client
	URLs                 map[string][]string    // General URLs of launched resources. Uses '_local' to delineate forwarded ports
	ChainlinkNodeDetails []*ChainlinkNodeDetail // ChainlinkNodeDetails has convenient details for connecting to chainlink deployments
	err                  error
}

// ChainlinkNodeDetail contains details about a chainlink node deployment
type ChainlinkNodeDetail struct {
	// ChartName details the name of the Helm chart this node uses, handy for modifying deployment values
	// Note: if you are using replicas of the same chart, this will be the same for all nodes
	// Use NewDeployment function for Chainlink nodes to make use of this
	ChartName string
	// PodName is the name of the pod running the chainlink node
	PodName string
	// LocalIP is the URL to connect to the node from the local machine
	LocalIP string
	// InternalIP is the URL to connect to the node from inside the K8s cluster
	InternalIP string
	// DBLocalIP is the URL to connect to the node's database from the local machine
	DBLocalIP string
}

// New creates new environment
func New(cfg *Config) *Environment {
	logging.Init()
	if cfg == nil {
		cfg = &Config{}
	}
	targetCfg := defaultEnvConfig()
	config.MustMerge(targetCfg, cfg)
	ns := os.Getenv(config.EnvVarNamespace)
	if ns != "" {
		cfg.Namespace = ns
	}
	if cfg.Namespace != "" {
		log.Info().Str("Namespace", cfg.Namespace).Msg("Namespace selected")
		targetCfg.Namespace = cfg.Namespace
	} else {
		targetCfg.Namespace = fmt.Sprintf("%s-%s", targetCfg.NamespacePrefix, uuid.NewString()[0:5])
		log.Info().Str("Namespace", targetCfg.Namespace).Msg("Creating new namespace")
	}
	jobImage := os.Getenv(config.EnvVarJobImage)
	if jobImage != "" {
		targetCfg.JobImage = jobImage
		targetCfg.detachRunner, _ = strconv.ParseBool(os.Getenv(config.EnvVarDetachRunner))
	} else {
		targetCfg.InsideK8s, _ = strconv.ParseBool(os.Getenv(config.EnvVarInsideK8s))
	}

	c, err := client.NewK8sClient()
	if err != nil {
		return &Environment{err: err}
	}
	e := &Environment{
		URLs:   make(map[string][]string),
		Charts: make([]ConnectedChart, 0),
		Client: c,
		Cfg:    targetCfg,
		Fwd:    client.NewForwarder(c, targetCfg.KeepConnection),
	}
	arts, err := NewArtifacts(e.Client, e.Cfg.Namespace)
	if err != nil {
		log.Error().Err(err).Msg("failed to create artifacts client")
		return &Environment{err: err}
	}
	e.Artifacts = arts

	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()
	if err := e.initApp(); err != nil {
		log.Error().Err(err).Msg("failed to create ns and service account")
		return &Environment{err: err}
	}
	e.Chaos = client.NewChaos(c, e.Cfg.Namespace)

	// setup test cleanup if this is using a remote runner
	// and not in detached mode
	// and not using an existing environment
	if targetCfg.JobImage != "" && !targetCfg.detachRunner && !targetCfg.NoManifestUpdate {
		targetCfg.fundReturnFailed = false
		if targetCfg.Test != nil {
			targetCfg.Test.Cleanup(func() {
				err := e.Shutdown()
				require.NoError(targetCfg.Test, err)
			})
		}
	}
	return e
}

func (m *Environment) initApp() error {
	var err error
	m.App = cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	m.Cfg.Labels = append(m.Cfg.Labels, "app.kubernetes.io/managed-by=cdk8s")
	owner := os.Getenv(config.EnvVarUser)
	if owner == "" {
		return fmt.Errorf("missing owner environment variable, please set %s to your name or if you are seeing this in CI please set it to ${{ github.actor }}", config.EnvVarUser)
	}
	m.Cfg.Labels = append(m.Cfg.Labels, fmt.Sprintf("owner=%s", owner))

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
		return err
	}
	defaultAnnotations[pkg.TTLLabelKey] = a.ShortDur(m.Cfg.TTL)
	m.root = cdk8s.NewChart(m.App, a.Str(fmt.Sprintf("root-chart-%s", m.Cfg.Namespace)), &cdk8s.ChartProps{
		Labels:    m.Cfg.nsLabels,
		Namespace: a.Str(m.Cfg.Namespace),
	})
	k8s.NewKubeNamespace(m.root, a.Str("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        a.Str(m.Cfg.Namespace),
			Labels:      m.Cfg.nsLabels,
			Annotations: &defaultAnnotations,
		},
	})
	k8s.NewKubeServiceAccount(m.root, a.Str("docker-creds-svc-acc"), &k8s.KubeServiceAccountProps{
		AutomountServiceAccountToken: nil,
		ImagePullSecrets: &[]*k8s.LocalObjectReference{
			{
				Name: a.Str("docker-creds"),
			},
		},
		Metadata: &k8s.ObjectMeta{
			Name:      a.Str("default"),
			Namespace: a.Str(m.Cfg.Namespace),
		},
	})
	zero := float64(0)
	k8s.NewKubePodDisruptionBudget(m.root, a.Str("pdb"), &k8s.KubePodDisruptionBudgetProps{
		Metadata: &k8s.ObjectMeta{
			Name:      a.Str("clenv-pdb"),
			Namespace: a.Str(m.Cfg.Namespace),
		},
		Spec: &k8s.PodDisruptionBudgetSpec{
			MaxUnavailable: k8s.IntOrString_FromNumber(&zero),
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{
					pkg.NamespaceLabelKey: a.Str(m.Cfg.Namespace),
				},
			},
		},
	})
	m.CurrentManifest = *m.App.SynthYaml()
	return m.Client.Apply(m.CurrentManifest)
}

// AddChart adds a chart to the deployment
func (m *Environment) AddChart(f func(root cdk8s.Chart) ConnectedChart) *Environment {
	if m.err != nil {
		return m
	}
	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()
	m.Charts = append(m.Charts, f(m.root))
	return m
}

func (m *Environment) removeChart(name string) error {
	chartIndex, _, err := m.findChart(name)
	if err != nil {
		return err
	}
	m.Charts = append(m.Charts[:chartIndex], m.Charts[chartIndex+1:]...)
	m.root.Node().TryRemoveChild(a.Str(name))
	return nil
}

// findChart finds a chart by name, returning the index of it in the Charts slice, and the chart itself
func (m *Environment) findChart(name string) (index int, chart ConnectedChart, err error) {
	for i, c := range m.Charts {
		if c.GetName() == name {
			return i, c, nil
		}
	}
	return -1, nil, fmt.Errorf("chart %s not found", name)
}

// ReplaceHelm entirely replaces an existing helm chart with a new one
// Note: you need to call Run() after this to apply the changes. If you're modifying ConfigMap values, you'll probably
// need to use RollOutStatefulSets to apply the changes to the pods. https://stackoverflow.com/questions/57356521/rollingupdate-for-stateful-set-doesnt-restart-pods-and-changes-from-updated-con
func (m *Environment) ReplaceHelm(name string, chart ConnectedChart) (*Environment, error) {
	if m.err != nil {
		return nil, m.err
	}
	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()
	if err := m.removeChart(name); err != nil {
		return nil, err
	}
	if m.Cfg.JobImage != "" || !chart.IsDeploymentNeeded() {
		return m, fmt.Errorf("cannot modify helm chart '%s' that does not need deployment, it may be in a remote runner or detached mode", name)
	}
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
	m.Charts = append(m.Charts, chart)
	return m, nil
}

// ModifyHelm entirely replaces a helm chart with a new one
//
// Deprecated: use ReplaceHelm instead to avoid silent errors
func (m *Environment) ModifyHelm(name string, chart ConnectedChart) *Environment {
	if m.err != nil {
		return m
	}
	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()
	if err := m.removeChart(name); err != nil {
		return nil
	}
	if m.Cfg.JobImage != "" || !chart.IsDeploymentNeeded() {
		return nil
	}
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
	m.Charts = append(m.Charts, chart)
	return m
}

// UpdateHelm update a helm chart with new values. The pod will launch with an `updated=true` label if it's a Chainlink node.
// Note: If you're modifying ConfigMap values, you'll probably need to use RollOutStatefulSets to apply the changes to the pods.
// https://stackoverflow.com/questions/57356521/rollingupdate-for-stateful-set-doesnt-restart-pods-and-changes-from-updated-con
func (m *Environment) UpdateHelm(name string, values map[string]any) (*Environment, error) {
	if m.err != nil {
		return nil, m.err
	}
	_, chart, err := m.findChart(name)
	if err != nil {
		return nil, err
	}
	if _, labelsExist := values["labels"]; !labelsExist {
		values["labels"] = make(map[string]*string)
	}
	values["labels"].(map[string]*string)["updated"] = a.Str("true")
	if err = mergo.Merge(chart.GetValues(), values, mergo.WithOverride); err != nil {
		return nil, err
	}
	return m.ReplaceHelm(name, chart)
}

// AddHelmCharts adds multiple helm charts to the testing environment
func (m *Environment) AddHelmCharts(charts []ConnectedChart) *Environment {
	if m.err != nil {
		return m
	}
	for _, c := range charts {
		m.AddHelm(c)
	}
	return m
}

// AddHelm adds a helm chart to the testing environment
func (m *Environment) AddHelm(chart ConnectedChart) *Environment {
	if m.err != nil {
		return m
	}
	if m.Cfg.JobImage != "" || !chart.IsDeploymentNeeded() {
		return m
	}
	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()

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
		a.Str("--skip-tests"),
	}
	if chart.GetVersion() != "" {
		helmFlags = append(helmFlags, a.Str("--version"), a.Str(chart.GetVersion()))
	}
	chartPath, err := m.PullOCIChart(chart)
	if err != nil {
		m.err = err
		return m
	}
	cdk8s.NewHelm(m.root, a.Str(chart.GetName()), &cdk8s.HelmProps{
		Chart:       a.Str(chartPath),
		HelmFlags:   &helmFlags,
		ReleaseName: a.Str(chart.GetName()),
		Values:      values,
	})
	m.Charts = append(m.Charts, chart)
	return m
}

// PullOCIChart handles working with OCI format repositories
// https://helm.sh/docs/topics/registries/
// API is not compatible between helm repos and OCI repos, so we download and untar the chart
func (m *Environment) PullOCIChart(chart ConnectedChart) (string, error) {
	if !strings.HasPrefix(chart.GetPath(), "oci") {
		return chart.GetPath(), nil
	}
	cp := strings.Split(chart.GetPath(), "/")
	if len(cp) != 5 {
		return "", fmt.Errorf(ErrInvalidOCI, chart.GetPath())
	}
	sp := strings.Split(chart.GetPath(), ":")

	var cmd string
	var chartName string
	chartName = cp[len(cp)-1]
	chartDir := uuid.NewString()
	switch len(sp) {
	case 2:
		cmd = fmt.Sprintf("helm pull %s --untar --untardir %s", chart.GetPath(), chartDir)
	case 3:
		chartName = strings.Split(chartName, ":")[0]
		cmd = fmt.Sprintf("helm pull %s --version %s --untar --untardir %s", fmt.Sprintf("%s:%s", sp[0], sp[1]), sp[2], chartDir)
	default:
		return "", fmt.Errorf(ErrInvalidOCI, chart.GetPath())
	}
	log.Info().Str("CMD", cmd).Msg("Running helm cmd")
	if err := client.ExecCmd(cmd); err != nil {
		return "", fmt.Errorf(ErrOCIPull, chart.GetPath())
	}
	localChartPath := fmt.Sprintf("%s/%s/", chartDir, chartName)
	log.Info().Str("Path", localChartPath).Msg("Local chart path")
	return localChartPath, nil
}

// PrintExportData prints export data
func (m *Environment) PrintExportData() error {
	m.URLs = make(map[string][]string)
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
func (m *Environment) ClearCharts() error {
	m.Charts = make([]ConnectedChart, 0)
	if err := m.initApp(); err != nil {
		log.Error().Err(err).Msg("failed to create ns and service account")
		return err
	}
	return nil
}

func (m *Environment) Manifest() string {
	return m.CurrentManifest
}

// Update current manifest based on the cdk8s app state
func (m *Environment) UpdateManifest() {
	config.JSIIGlobalMu.Lock()
	m.CurrentManifest = *m.App.SynthYaml()
	config.JSIIGlobalMu.Unlock()
}

// RunCustomReadyConditions Runs the environment with custom ready conditions for a supplied pod count
func (m *Environment) RunCustomReadyConditions(customCheck *client.ReadyCheckData, podCount int) error {
	if m.err != nil {
		return m.err
	}
	if m.Cfg.jobDeployed {
		return nil
	}
	if m.Cfg.JobImage != "" {
		rrSelector := map[string]*string{"remote-type": a.Str("test")}
		m.AddChart(NewRunner(&Props{
			BaseName:         REMOTE_RUNNER_NAME,
			TargetNamespace:  m.Cfg.Namespace,
			Labels:           &rrSelector,
			Image:            m.Cfg.JobImage,
			TestName:         m.Cfg.Test.Name(),
			NoManifestUpdate: m.Cfg.NoManifestUpdate,
		}))
	}
	m.UpdateManifest()
	m.ChainlinkNodeDetails = []*ChainlinkNodeDetail{} // Resets potentially old details if re-deploying
	if m.Cfg.DryRun {
		log.Info().Msg("Dry-run mode, manifest synthesized and saved as tmp-manifest.yaml")
		return nil
	}
	manifestUpdate := os.Getenv(config.EnvVarNoManifestUpdate)
	if manifestUpdate != "" {
		mu, err := strconv.ParseBool(manifestUpdate)
		if err != nil {
			return fmt.Errorf("manifest update should be bool: true, false")
		}
		m.Cfg.NoManifestUpdate = mu
	}
	log.Info().Bool("ManifestUpdate", !m.Cfg.NoManifestUpdate).Msg("Update mode")
	if !m.Cfg.NoManifestUpdate || m.Cfg.JobImage != "" {
		if err := m.DeployCustomReadyConditions(customCheck, podCount); err != nil {
			log.Error().Err(err).Msg("Error deploying environment")
			_ = m.Shutdown()
			return err
		}
	}
	if m.Cfg.JobImage != "" {
		log.Info().Msg("Waiting for remote runner to complete")
		// Do not wait for the job to complete if we are running something like a soak test in the remote runner
		if m.Cfg.detachRunner {
			return nil
		}
		if err := m.Client.WaitForJob(m.Cfg.Namespace, "remote-test-runner", func(message string) {
			m.Cfg.Test.Log(message)
			found := strings.Contains(message, FAILED_FUND_RETURN)
			if found {
				m.Cfg.fundReturnFailed = true
			}
		}); err != nil {
			return err
		}
		if m.Cfg.fundReturnFailed {
			return errors.New("failed to return funds in remote runner.")
		}
		m.Cfg.jobDeployed = true
	} else {
		if err := m.Fwd.Connect(m.Cfg.Namespace, "", m.Cfg.InsideK8s); err != nil {
			return err
		}
		log.Info().Interface("Ports", m.Fwd.Info).Msg("Forwarded ports")
		m.Fwd.PrintLocalPorts()
		if err := m.PrintExportData(); err != nil {
			return err
		}
		arts, err := NewArtifacts(m.Client, m.Cfg.Namespace)
		if err != nil {
			log.Error().Err(err).Msg("failed to create artifacts client")
			return err
		}
		m.Artifacts = arts
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
	}
	return nil
}

// RunUpdated runs the environment and checks for pods with `updated=true` label
func (m *Environment) RunUpdated(podCount int) error {
	if m.err != nil {
		return m.err
	}
	conds := &client.ReadyCheckData{
		ReadinessProbeCheckSelector: "updated=true",
		Timeout:                     10 * time.Minute,
	}
	return m.RunCustomReadyConditions(conds, podCount)
}

// Run deploys or connects to already created environment
func (m *Environment) Run() error {
	if m.err != nil {
		return m.err
	}
	return m.RunCustomReadyConditions(nil, 0)
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

// DeployCustomReadyConditions deploy current manifest with added custom readiness checks
func (m *Environment) DeployCustomReadyConditions(customCheck *client.ReadyCheckData, customPodCount int) error {
	if m.err != nil {
		return m.err
	}
	log.Info().Str("Namespace", m.Cfg.Namespace).Msg("Deploying namespace")

	if m.Cfg.DryRun {
		if err := m.Client.DryRun(m.CurrentManifest); err != nil {
			return err
		}
		return nil
	}
	if err := m.Client.Apply(m.CurrentManifest); err != nil {
		return err
	}
	if int64(m.Cfg.UpdateWaitInterval) != 0 {
		time.Sleep(m.Cfg.UpdateWaitInterval)
	}

	expectedPodCount := m.findPodCountInDeploymentManifest()

	if err := m.Client.WaitPodsReady(m.Cfg.Namespace, m.Cfg.ReadyCheckData, expectedPodCount); err != nil {
		return err
	}
	if customCheck != nil {
		if err := m.Client.WaitPodsReady(m.Cfg.Namespace, customCheck, customPodCount); err != nil {
			return err
		}
	}
	if err := m.enumerateApps(); err != nil {
		return err
	}
	podList, err := m.Client.ListPods(m.Cfg.Namespace, "")
	if err != nil {
		return err
	}
	defaultPodLabels[pkg.NamespaceLabelKey] = m.Cfg.Namespace
	if err := m.Client.AddPodsLabels(m.Cfg.Namespace, podList, defaultPodLabels); err != nil {
		return err
	}

	return m.Client.AddPodsAnnotations(m.Cfg.Namespace, podList, defaultPodAnnotations)
}

// Deploy deploy current manifest and check logs for readiness
func (m *Environment) Deploy() error {
	return m.DeployCustomReadyConditions(nil, 0)
}

// RolloutStatefulSets applies "rollout statefulset" to all existing statefulsets in our namespace
func (m *Environment) RolloutStatefulSets() error {
	if m.err != nil {
		return m.err
	}
	return m.Client.RolloutStatefulSets(m.Cfg.Namespace)
}

// findPodsInDeploymentManifest finds all the pods we will be deploying
func (m *Environment) findPodCountInDeploymentManifest() int {
	podCount := 0
	config.JSIIGlobalMu.Lock()
	defer config.JSIIGlobalMu.Unlock()
	charts := m.App.Charts()
	for _, chart := range *charts {
		json := chart.ToJson()
		if json == nil {
			continue
		}
		for _, j := range *json {
			m := j.(map[string]interface{})
			// if the kind is a deployment then we want to see if it has replicas to count towards the app count
			kind := m["kind"].(string)
			if kind == "Deployment" || kind == "StatefulSet" {
				podCount += getReplicaCount(m["spec"].(map[string]interface{}))
			}
		}

	}
	return podCount
}

func getReplicaCount(spec map[string]interface{}) int {
	if spec == nil {
		return 0
	}
	s := spec["selector"].(map[string]interface{})
	if s == nil {
		return 0
	}
	m := s["matchLabels"].(map[string]interface{})
	if m == nil {
		return 0
	}
	l := m["app"]
	if l == nil {
		return 0
	}

	replicaCount := 0
	replicas := spec["replicas"]
	if replicas != nil {
		replicaCount += int(replicas.(float64))
	} else {
		replicaCount += 1
	}

	return replicaCount
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
	// don't shutdown if returning of funds failed
	if m.Cfg.fundReturnFailed {
		return nil
	}

	// don't shutdown if this is a test running remotely
	if m.Cfg.InsideK8s {
		return nil
	}

	keepEnvs := os.Getenv(config.EnvVarKeepEnvironments)
	if keepEnvs == "" {
		keepEnvs = "NEVER"
	}

	shouldShutdown := false
	switch strings.ToUpper(keepEnvs) {
	case "ALWAYS":
		return nil
	case "ONFAIL":
		if m.Cfg.Test != nil {
			if !m.Cfg.Test.Failed() {
				shouldShutdown = true
			}
		}
	case "NEVER":
		shouldShutdown = true
	default:
		log.Warn().Str("Invalid Keep Value", keepEnvs).
			Msg("Invalid 'keep_environments' value, see the KEEP_ENVIRONMENTS env var")
	}

	if shouldShutdown {
		return m.Client.RemoveNamespace(m.Cfg.Namespace)
	}
	return nil
}

// BeforeTest sets the test name variable and determines if we need to start the remote runner
func (m *Environment) WillUseRemoteRunner() bool {
	val, _ := os.LookupEnv(config.EnvVarJobImage)
	return val != "" && m.Cfg.Test.Name() != ""
}
