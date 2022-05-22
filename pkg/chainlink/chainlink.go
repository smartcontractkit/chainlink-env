package chainlink

import (
	"fmt"
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"github.com/smartcontractkit/chainlink-env/pkg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
	"github.com/smartcontractkit/chainlink-env/pkg/chains/ethereum"
	ms "github.com/smartcontractkit/chainlink-env/pkg/mockserver"
	"os"
	"time"
)

// Control labels used to list envs created by the wizard
const (
	ControlLabelKey        = "generatedBy"
	ControlLabelValue      = "cdk8s"
	ControlLabelEnvTypeKey = "envType"
)

const (
	EnvTypeEVM5     = "evm-5-minimal-local"
	EnvTypeEVM5Soak = "evm-5-soak"
	EnvTypeSolana5  = "solana-5-default"
)

const (
	AppName           = "chainlink-node"
	NodeContainerName = "node"
	GethURLsKey       = "geth"
	NodesLocalURLsKey = "chainlink_local"
	NodesInternalKey  = "chainlink_internal"
	DBsInternalKey    = "chainlink_db"
	MockServerURLsKey = "mockserver"
)

var (
	defaultAnnotations = map[string]*string{"prometheus.io/scrape": a.Str("true")}
)

// internalChartVars some shared labels/selectors and names that must match in resources
type internalChartVars struct {
	InstanceCounter int
	NodeLabels      map[string]*string
	DeploymentName  string
	ConfigMapName   string
	ServiceName     string
}

// PersistenceProps database persistence props
type PersistenceProps struct {
	Capacity string
}

// VersionProps CL application props for a particular version
type VersionProps struct {
	Image         string `envconfig:"IMAGE"`
	Tag           string `envconfig:"TAG"`
	Instances     int
	Env           *NodeEnvVars
	ResourcesMode pkg.ResourcesMode
}

// Props root Chainlink props
type Props struct {
	Namespace       string
	TTL             time.Duration
	Labels          []string
	ChainProps      []interface{}
	AppVersions     []VersionProps
	TestRunnerProps interface{}
	Persistence     PersistenceProps
	ResourcesMode   pkg.ResourcesMode
	vars            *internalChartVars
}

func pgIsReadyCheck() *[]*string {
	return &[]*string{
		a.Str("pg_isready"),
		a.Str("-U"),
		a.Str("postgres"),
	}
}

func chains(chart cdk8s.Chart, p *Props) {
	for _, chainProps := range p.ChainProps {
		switch c := chainProps.(type) {
		case *ethereum.Props:
			ethereum.NewEthereum(chart, c, p.ResourcesMode)
		default:
			log.Fatal().Msg("no chain props found, provide one of a supported chain props")
		}
	}
}

// versionedDeployments creates Deployment or StatefulSet for each selected CL version
func versionedDeployments(chart cdk8s.Chart, p *Props) {
	for _, verProps := range p.AppVersions {
		config.MustEnvOverrideStruct("", &verProps)
		for sameVersionInstance := 0; sameVersionInstance < verProps.Instances; sameVersionInstance++ {
			p.vars.DeploymentName = fmt.Sprintf("chainlink-%d", p.vars.InstanceCounter)
			p.vars.ConfigMapName = fmt.Sprintf("chainlink-cm-%d", p.vars.InstanceCounter)
			p.vars.ServiceName = fmt.Sprintf("chainlink-service-%d", p.vars.InstanceCounter)
			p.vars.NodeLabels = map[string]*string{
				"app":      a.Str(AppName),
				"instance": a.Str(fmt.Sprintf("%d", p.vars.InstanceCounter)),
			}
			configMap(chart, p)
			service(chart, p)
			if p.Persistence.Capacity != "" {
				statefulConstruct(chart, p, verProps)
			} else {
				deploymentConstruct(chart, p, verProps)
			}
			p.vars.InstanceCounter++
		}
	}
}

// NodeEnvVars node environment variables, get them from CL repo later
type NodeEnvVars struct {
	DatabaseURL                               string `envconfig:"DATABASE_URL" `
	DatabaseName                              string `envconfig:"DATABASE_NAME" `
	EthURL                                    string `envconfig:"ETH_URL" `
	EthChainID                                string `envconfig:"ETH_CHAIN_ID" `
	AllowOrigins                              string `envconfig:"ALLOW_ORIGINS" `
	ChainlinkDev                              string `envconfig:"CHAINLINK_DEV" `
	ETHDisabled                               string `envconfig:"ETH_DISABLED" `
	FeatureExternalInitiators                 string `envconfig:"FEATURE_EXTERNAL_INITIATORS" `
	ChainlinkPassword                         string `envconfig:"CHAINLINK_PGPASSWORD" `
	ChainlinkPort                             string `envconfig:"CHAINLINK_PORT" `
	ChainlinkTLSPort                          string `envconfig:"CHAINLINK_TLS_PORT" `
	DefaultHTTPAllowUnrestrictedNetworkAccess string `envconfig:"DEFAULT_HTTP_ALLOW_UNRESTRICTED_NETWORK_ACCESS" `
	EnableBulletproofTXManager                string `envconfig:"ENABLE_BULLETPROOF_TX_MANAGER" `
	FeatureOffchainReporting                  string `envconfig:"FEATURE_OFFCHAIN_REPORTING" `
	JsonConsole                               string `envconfig:"JSON_CONSOLE" `
	LogLevel                                  string `envconfig:"LOG_LEVEL" `
	MaxExportHTMLThreads                      string `envconfig:"MAX_EXPORT_HTML_THREADS" `
	MinimumContractPaymentLinkJuels           string `envconfig:"MINIMUM_CONTRACT_PAYMENT_LINK_JUELS" `
	OCRTraceLogging                           string `envconfig:"OCR_TRACE_LOGGING" `
	P2PListenIP                               string `envconfig:"P2P_LISTEN_IP" `
	P2PListenPort                             string `envconfig:"P2P_LISTEN_PORT" `
	Root                                      string `envconfig:"ROOT" `
	SecureCookies                             string `envconfig:"SECURE_COOKIES" `
	ETHMaxInFlightTransactions                string `envconfig:"ETH_MAX_IN_FLIGHT_TRANSACTIONS" `
	// Explorer
	ExplorerURL       string `envconfig:"EXPLORER_URL" `
	ExplorerAccessKey string `envconfig:"EXPLORER_ACCESS_KEY" `
	ExplorerSecret    string `envconfig:"EXPLORER_SECRET" `
	// Keeper
	KeeperDefaultTransactionQueueDepth string `envconfig:"KEEPER_DEFAULT_TRANSACTION_QUEUE_DEPTH" `
	KeeperRegistrySyncInterval         string `envconfig:"KEEPER_REGISTRY_SYNC_INTERVAL" `
	KeeperMinimumRequiredConfirmations string `envconfig:"KEEPER_MINIMUM_REQUIRED_CONFIRMATIONS" `
	KeeperRegistryPerformGasOverhead   string `envconfig:"KEEPER_REGISTRY_PERFORM_GAS_OVERHEAD" `
}

func defaultNodeEnvVars() *NodeEnvVars {
	return &NodeEnvVars{
		DatabaseURL:               "postgresql://postgres:node@0.0.0.0/chainlink?sslmode=disable",
		DatabaseName:              "chainlink",
		EthURL:                    "ws://geth:8546",
		EthChainID:                "1337",
		AllowOrigins:              "*",
		ChainlinkDev:              "true",
		ETHDisabled:               "false",
		FeatureExternalInitiators: "false",
		ChainlinkPassword:         "node",
		ChainlinkPort:             "6688",
		ChainlinkTLSPort:          "0",
		DefaultHTTPAllowUnrestrictedNetworkAccess: "true",
		EnableBulletproofTXManager:                "true",
		FeatureOffchainReporting:                  "true",
		JsonConsole:                               "false",
		LogLevel:                                  "debug",
		MaxExportHTMLThreads:                      "2",
		MinimumContractPaymentLinkJuels:           "0",
		OCRTraceLogging:                           "true",
		P2PListenIP:                               "0.0.0.0",
		P2PListenPort:                             "6690",
		Root:                                      "./clroot",
		SecureCookies:                             "false",
		ETHMaxInFlightTransactions:                "5000",
		ExplorerURL:                               "",
		ExplorerAccessKey:                         "",
		ExplorerSecret:                            "",
		KeeperDefaultTransactionQueueDepth:        "1",
		KeeperRegistrySyncInterval:                "5s",
		KeeperMinimumRequiredConfirmations:        "1",
		KeeperRegistryPerformGasOverhead:          "1",
	}
}

// chainlinkContainer CL container spec
func chainlinkContainer(p *Props, verProps VersionProps) *k8s.Container {
	c := &k8s.Container{
		Name:            a.Str(NodeContainerName),
		Image:           a.Str(fmt.Sprintf("%s:%s", verProps.Image, verProps.Tag)),
		ImagePullPolicy: a.Str("Always"),
		Args: &[]*string{
			a.Str("node"),
			a.Str("start"),
			a.Str("-d"),
			a.Str("-p"),
			a.Str("/etc/node-secrets-volume/node-password"),
			a.Str("-a"),
			a.Str("/etc/node-secrets-volume/apicredentials"),
			a.Str("--vrfpassword=/etc/node-secrets-volume/apicredentials"),
		},
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Str("chainlink-config-map"),
				MountPath: a.Str("/etc/node-secrets-volume"),
			},
		},
		LivenessProbe: &k8s.Probe{
			HttpGet: &k8s.HttpGetAction{
				Port: k8s.IntOrString_FromNumber(a.Num(6688)),
				Path: a.Str("/"),
			},
			InitialDelaySeconds: a.Num(20),
			PeriodSeconds:       a.Num(5),
		},
		ReadinessProbe: &k8s.Probe{
			HttpGet: &k8s.HttpGetAction{
				Port: k8s.IntOrString_FromNumber(a.Num(6688)),
				Path: a.Str("/"),
			},
			InitialDelaySeconds: a.Num(10),
			PeriodSeconds:       a.Num(5),
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("access"),
				ContainerPort: a.Num(6688),
			},
			{
				Name:          a.Str("p2p"),
				ContainerPort: a.Num(8899),
			},
		},
		Env: a.MustChartEnvVarsFromStruct("", defaultNodeEnvVars(), verProps.Env),
	}
	switch p.ResourcesMode {
	case pkg.MinimalLocalResourcesMode:
		c.Resources = a.ContainerResources("200m", "1024Mi", "200m", "1024Mi")
	case pkg.SoakResourcesMode:
		c.Resources = a.ContainerResources("1000m", "2048Mi", "1000m", "2048Mi")
	default:
		log.Fatal().Msg("unrecognized resources mode")
	}
	return c
}

// postgresContainer postgres container spec
func postgresContainer(p *Props, verProps VersionProps) *k8s.Container {
	c := &k8s.Container{
		Name:  a.Str("chainlink-db"),
		Image: a.Str("postgres:11.6"),
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("postgres"),
				ContainerPort: a.Num(5432),
			},
		},
		Env: &[]*k8s.EnvVar{
			a.EnvVarStr("POSTGRES_DB", "chainlink"),
			a.EnvVarStr("POSTGRES_PASSWORD", "node"),
			a.EnvVarStr("PGPASSWORD", "node"),
			a.EnvVarStr("PGUSER", "postgres"),
		},
		LivenessProbe: &k8s.Probe{
			Exec: &k8s.ExecAction{
				Command: pgIsReadyCheck()},
			InitialDelaySeconds: a.Num(60),
			PeriodSeconds:       a.Num(60),
		},
		ReadinessProbe: &k8s.Probe{
			Exec: &k8s.ExecAction{
				Command: pgIsReadyCheck()},
			InitialDelaySeconds: a.Num(2),
			PeriodSeconds:       a.Num(2),
		},
		Resources: a.ContainerResources("450m", "1024Mi", "450m", "1024Mi"),
	}
	switch p.ResourcesMode {
	case pkg.MinimalLocalResourcesMode:
		c.Resources = a.ContainerResources("300m", "1024Mi", "300m", "1024Mi")
	case pkg.SoakResourcesMode:
		c.Resources = a.ContainerResources("2000m", "4086Mi", "2000m", "4086Mi")
	default:
		log.Fatal().Msg("unrecognized resources mode")
	}
	if p.Persistence.Capacity != "" {
		c.VolumeMounts = &[]*k8s.VolumeMount{
			{
				Name:      a.Str("postgres"),
				SubPath:   a.Str("postgres-db"),
				MountPath: a.Str("/var/lib/postgresql/data"),
			},
		}
	}
	return c
}

func deploymentConstruct(chart cdk8s.Chart, props *Props, verProps VersionProps) {
	k8s.NewKubeDeployment(
		chart,
		a.Str(props.vars.DeploymentName),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(props.vars.DeploymentName),
			},
			Spec: &k8s.DeploymentSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: &props.vars.NodeLabels,
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: &props.vars.NodeLabels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Str("chainlink-config-map"),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Str(props.vars.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							postgresContainer(props, verProps),
							chainlinkContainer(props, verProps),
						},
					},
				},
			},
		})
}

func statefulConstruct(chart cdk8s.Chart, p *Props, verProps VersionProps) {
	k8s.NewKubeStatefulSet(
		chart,
		a.Str(p.vars.DeploymentName),
		&k8s.KubeStatefulSetProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(p.vars.DeploymentName),
			},
			Spec: &k8s.StatefulSetSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: &p.vars.NodeLabels,
				},
				ServiceName:         a.Str(p.vars.ServiceName),
				PodManagementPolicy: a.Str("Parallel"),
				VolumeClaimTemplates: &[]*k8s.KubePersistentVolumeClaimProps{
					{
						Metadata: &k8s.ObjectMeta{
							Name: a.Str("postgres"),
						},
						Spec: &k8s.PersistentVolumeClaimSpec{
							AccessModes: &[]*string{a.Str("ReadWriteOnce")},
							Resources: &k8s.ResourceRequirements{
								Requests: &map[string]k8s.Quantity{
									"storage": k8s.Quantity_FromString(a.Str(p.Persistence.Capacity)),
								},
							},
						},
					},
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: &p.vars.NodeLabels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Str("chainlink-config-map"),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Str(p.vars.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							postgresContainer(p, verProps),
							chainlinkContainer(p, verProps),
						},
					},
				},
			},
		})
}

// service k8s service spec
func service(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeService(chart, a.Str(props.vars.ServiceName), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(props.vars.ServiceName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("access"),
					Port:       a.Num(6688),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(6688)),
				},
				{
					Name:       a.Str("p2p"),
					Port:       a.Num(8899),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(8899)),
				}},
			Selector: &props.vars.NodeLabels,
		},
	})
}

// configMap k8s configMap spec
func configMap(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeConfigMap(chart, a.Str(props.vars.ConfigMapName), &k8s.KubeConfigMapProps{
		Data: &map[string]*string{
			"apicredentials": a.Str("notreal@fakeemail.ch\ntwochains"),
			"node-password":  a.Str("T.tLHkcmwePT/p,]sYuntjwHKAsrhm#4eRs4LuKHwvHejWYAC2JP4M8HimwgmbaZ"),
		},
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(props.vars.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Str(props.vars.ConfigMapName),
			},
		},
	})
}

// ManifestOutputData checks if all selected pods are ready by tailing the logs and checking substrings
type ManifestOutputData struct {
	Namespace       string
	ReadyCheckData  client.ReadyCheckData
	CommonChartVars internalChartVars
}

func (m *ManifestOutputData) SetNamespace(ns string) {
	m.Namespace = ns
}

func (m *ManifestOutputData) GetNamespace() string {
	return m.Namespace
}

func (m *ManifestOutputData) GetReadyCheckData() client.ReadyCheckData {
	return m.ReadyCheckData
}

func (m *ManifestOutputData) ProcessConnections(fwd *client.Forwarder) (map[string][]string, error) {
	urlsByApp := make(map[string][]string)
	geth, err := fwd.FindPort("geth:", "geth", "ws-rpc").As(client.LocalConnection, client.WS)
	if err != nil {
		return nil, err
	}
	urlsByApp[GethURLsKey] = append(urlsByApp[GethURLsKey], geth)

	mock, err := fwd.FindPort("mockserver:", "mockserver", "serviceport").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return nil, err
	}
	urlsByApp[MockServerURLsKey] = append(urlsByApp[MockServerURLsKey], mock)
	pods, err := fwd.Client.ListPods(m.Namespace, fmt.Sprintf("app=%s", AppName))
	if err != nil {
		return nil, err
	}
	log.Info().Str("URL", geth).Msg("Geth network")
	for i := 0; i < len(pods.Items); i++ {
		n, err := fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "node", "access").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return nil, err
		}
		urlsByApp[NodesLocalURLsKey] = append(urlsByApp[NodesLocalURLsKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("Local connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "node", "access").
			As(client.RemoteConnection, client.HTTP)
		if err != nil {
			return nil, err
		}
		urlsByApp[NodesInternalKey] = append(urlsByApp[NodesInternalKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("Remote (in cluster) connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "chainlink-db", "postgres").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return nil, err
		}
		urlsByApp[DBsInternalKey] = append(urlsByApp[DBsInternalKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("DB local Connection")
	}
	return urlsByApp, nil
}

func mockserver(chart cdk8s.Chart, _ *Props) {
	ms.NewChart(chart, &ms.Props{})
}

// NewChart root manifest creation function
func NewChart(props interface{}) (cdk8s.App, client.ManifestOutput) {
	p := props.(*Props)
	app := cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	p.Namespace = fmt.Sprintf("%s-%s", p.Namespace, uuid.NewString()[0:5])
	p.Labels = append(p.Labels, "generatedBy=cdk8s")
	p.Labels = append(p.Labels, fmt.Sprintf("owner=%s", os.Getenv("CHAINLINK_ENV_USER")))
	labels, err := a.ConvertLabels(p.Labels)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	chart := cdk8s.NewChart(app, a.Str("chainlink"), &cdk8s.ChartProps{
		Labels:    nil,
		Namespace: a.Str(p.Namespace),
	})
	defaultAnnotations["janitor/ttl"] = a.ShortDur(p.TTL)
	k8s.NewKubeNamespace(chart, a.Str("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        a.Str(p.Namespace),
			Labels:      labels,
			Annotations: &defaultAnnotations,
		},
	})
	p.vars = &internalChartVars{InstanceCounter: 0}

	versionedDeployments(chart, p)
	chains(chart, p)
	mockserver(chart, p)
	return app, &ManifestOutputData{
		Namespace: p.Namespace,
		ReadyCheckData: client.ReadyCheckData{
			Timeout:   3 * time.Minute,
			Selector:  "app=chainlink-node",
			Container: "node",
			LogSubStr: "Subscribed to heads on chain",
		},
	}
}
