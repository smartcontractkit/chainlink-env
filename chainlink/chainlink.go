package chainlink

import (
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/rs/zerolog/log"
	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/chains/ethereum"
	"github.com/smartcontractkit/chainlink-env/chains/solana"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"time"
)

const (
	AppName           = "chainlink-node"
	NodeContainerName = "node"
	GethURLsKey       = "geth"
	NodesLocalURLsKey = "chainlink_local"
	NodesInternalKey  = "chainlink_internal"
	DBsInternalKey    = "chainlink_db"
)

var (
	scrapeAnnotation = &map[string]*string{"prometheus.io/scrape": a.Jss("true")}
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
	Image       string `envconfig:"IMAGE"`
	Tag         string `envconfig:"TAG"`
	Instances   int
	Env         *NodeEnvVars
	Persistence PersistenceProps
}

// Props root Chainlink props
type Props struct {
	Namespace       string
	ChainProps      []interface{}
	AppVersions     []VersionProps
	TestRunnerProps interface{}
	vars            *internalChartVars
}

func pgIsReadyCheck() *[]*string {
	return &[]*string{
		a.Jss("pg_isready"),
		a.Jss("-U"),
		a.Jss("postgres"),
	}
}

func chains(chart constructs.Construct, chains []interface{}) {
	for _, c := range chains {
		switch c.(type) {
		case *ethereum.Props:
			ethereum.NewEthereumChart(chart, c.(*ethereum.Props))
		case *solana.Props:
			solana.NewSolanaChart(chart, c.(*solana.Props))
		}
	}
}

// versionedDeployments creates Deployment or StatefulSet for each selected CL version
func versionedDeployments(chart cdk8s.Chart, p *Props) {
	p.vars = &internalChartVars{InstanceCounter: 0}
	for _, verProps := range p.AppVersions {
		a.MustOverrideStruct("", &verProps)
		for sameVersionInstance := 0; sameVersionInstance < verProps.Instances; sameVersionInstance++ {
			p.vars.DeploymentName = fmt.Sprintf("chainlink-%d", p.vars.InstanceCounter)
			p.vars.ConfigMapName = fmt.Sprintf("chainlink-cm-%d", p.vars.InstanceCounter)
			p.vars.ServiceName = fmt.Sprintf("chainlink-service-%d", p.vars.InstanceCounter)
			p.vars.NodeLabels = map[string]*string{
				"app":      a.Jss(AppName),
				"instance": a.Jss(fmt.Sprintf("%d", p.vars.InstanceCounter)),
			}
			configMap(chart, p)
			service(chart, p)
			if verProps.Persistence.Capacity != "" {
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

func DefaultNodeEnvVars() *NodeEnvVars {
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
func chainlinkContainer(verProps VersionProps) *k8s.Container {
	c := &k8s.Container{
		Name:            a.Jss(NodeContainerName),
		Image:           a.Jss(fmt.Sprintf("%s:%s", verProps.Image, verProps.Tag)),
		ImagePullPolicy: a.Jss("Always"),
		Args: &[]*string{
			a.Jss("node"),
			a.Jss("start"),
			a.Jss("-d"),
			a.Jss("-p"),
			a.Jss("/etc/node-secrets-volume/node-password"),
			a.Jss("-a"),
			a.Jss("/etc/node-secrets-volume/apicredentials"),
			a.Jss("--vrfpassword=/etc/node-secrets-volume/apicredentials"),
		},
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Jss("chainlink-config-map"),
				MountPath: a.Jss("/etc/node-secrets-volume"),
			},
		},
		LivenessProbe: &k8s.Probe{
			HttpGet: &k8s.HttpGetAction{
				Port: k8s.IntOrString_FromNumber(a.Jsn(6688)),
				Path: a.Jss("/"),
			},
			InitialDelaySeconds: a.Jsn(20),
			PeriodSeconds:       a.Jsn(5),
		},
		ReadinessProbe: &k8s.Probe{
			HttpGet: &k8s.HttpGetAction{
				Port: k8s.IntOrString_FromNumber(a.Jsn(6688)),
				Path: a.Jss("/"),
			},
			InitialDelaySeconds: a.Jsn(10),
			PeriodSeconds:       a.Jsn(5),
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Jss("access"),
				ContainerPort: a.Jsn(6688),
			},
			{
				Name:          a.Jss("p2p"),
				ContainerPort: a.Jsn(8899),
			},
		},
		Env:       a.MustEnvVarsFromEnvconfigPrefix("", DefaultNodeEnvVars(), verProps.Env),
		Resources: a.ContainerResources("200m", "1024Mi", "200m", "1024Mi"),
	}
	return c
}

// postgresContainer postgres container spec
func postgresContainer(verProps VersionProps) *k8s.Container {
	c := &k8s.Container{
		Name:  a.Jss("chainlink-db"),
		Image: a.Jss("postgres:11.6"),
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Jss("postgres"),
				ContainerPort: a.Jsn(5432),
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
			InitialDelaySeconds: a.Jsn(60),
			PeriodSeconds:       a.Jsn(60),
		},
		ReadinessProbe: &k8s.Probe{
			Exec: &k8s.ExecAction{
				Command: pgIsReadyCheck()},
			InitialDelaySeconds: a.Jsn(2),
			PeriodSeconds:       a.Jsn(2),
		},
		Resources: a.ContainerResources("450m", "1024Mi", "450m", "1024Mi"),
	}
	if verProps.Persistence.Capacity != "" {
		c.VolumeMounts = &[]*k8s.VolumeMount{
			{
				Name:      a.Jss("postgres"),
				SubPath:   a.Jss("postgres-db"),
				MountPath: a.Jss("/var/lib/postgresql/data"),
			},
		}
	}
	return c
}

func deploymentConstruct(chart cdk8s.Chart, props *Props, verProps VersionProps) {
	k8s.NewKubeDeployment(
		chart,
		a.Jss(props.vars.DeploymentName),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Jss(props.vars.DeploymentName),
			},
			Spec: &k8s.DeploymentSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: &props.vars.NodeLabels,
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Annotations: scrapeAnnotation,
						Labels:      &props.vars.NodeLabels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Jss("chainlink-config-map"),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Jss(props.vars.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Jss("default"),
						Containers: &[]*k8s.Container{
							postgresContainer(verProps),
							chainlinkContainer(verProps),
						},
					},
				},
			},
		})
}

func statefulConstruct(chart cdk8s.Chart, props *Props, verProps VersionProps) {
	k8s.NewKubeStatefulSet(
		chart,
		a.Jss(props.vars.DeploymentName),
		&k8s.KubeStatefulSetProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Jss(props.vars.DeploymentName),
			},
			Spec: &k8s.StatefulSetSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: &props.vars.NodeLabels,
				},
				ServiceName:         a.Jss(props.vars.ServiceName),
				PodManagementPolicy: a.Jss("Parallel"),
				VolumeClaimTemplates: &[]*k8s.KubePersistentVolumeClaimProps{
					{
						Metadata: &k8s.ObjectMeta{
							Name: a.Jss("postgres"),
						},
						Spec: &k8s.PersistentVolumeClaimSpec{
							AccessModes: &[]*string{a.Jss("ReadWriteOnce")},
							Resources: &k8s.ResourceRequirements{
								Requests: &map[string]k8s.Quantity{
									"storage": k8s.Quantity_FromString(a.Jss(verProps.Persistence.Capacity)),
								},
							},
						},
					},
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Annotations: scrapeAnnotation,
						Labels:      &props.vars.NodeLabels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Jss("chainlink-config-map"),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Jss(props.vars.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Jss("default"),
						Containers: &[]*k8s.Container{
							postgresContainer(verProps),
							chainlinkContainer(verProps),
						},
					},
				},
			},
		})
}

// service k8s service spec
func service(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeService(chart, a.Jss(props.vars.ServiceName), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss(props.vars.ServiceName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Jss("access"),
					Port:       a.Jsn(6688),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(6688)),
				},
				{
					Name:       a.Jss("p2p"),
					Port:       a.Jsn(8899),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(8899)),
				}},
			Selector: &props.vars.NodeLabels,
		},
	})
}

// configMap k8s configMap spec
func configMap(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeConfigMap(chart, a.Jss(props.vars.ConfigMapName), &k8s.KubeConfigMapProps{
		Data: &map[string]*string{
			"apicredentials": a.Jss("notreal@fakeemail.ch\ntwochains"),
			"node-password":  a.Jss("T.tLHkcmwePT/p,]sYuntjwHKAsrhm#4eRs4LuKHwvHejWYAC2JP4M8HimwgmbaZ"),
		},
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss(props.vars.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Jss(props.vars.ConfigMapName),
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
	pods, err := fwd.Client.ListPods(m.Namespace, fmt.Sprintf("app=%s", AppName))
	if err != nil {
		return nil, err
	}
	urlsByApp[GethURLsKey] = append(urlsByApp[GethURLsKey], geth)
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

// Manifest root manifest creation function
func Manifest(props interface{}) (cdk8s.App, client.ManifestOutput) {
	p := props.(*Props)
	app := cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	chart := cdk8s.NewChart(app, a.Jss("chainlink"), &cdk8s.ChartProps{
		Labels:    nil,
		Namespace: a.Jss(p.Namespace),
	})
	k8s.NewKubeNamespace(chart.(constructs.Construct), a.Jss("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name:   a.Jss(p.Namespace),
			Labels: &map[string]*string{"generatedBy": a.Jss("cdk8s")},
		},
	})
	versionedDeployments(chart, p)
	chains(chart.(constructs.Construct), p.ChainProps)
	checkData := &ManifestOutputData{
		Namespace: p.Namespace,
		ReadyCheckData: client.ReadyCheckData{
			Timeout:   3 * time.Minute,
			Selector:  "app=chainlink-node",
			Container: "node",
			LogSubStr: "Subscribed to heads on chain",
		},
	}
	return app, checkData
}
