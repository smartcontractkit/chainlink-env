package mockserver

import (
	"fmt"
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

type Props struct{}

// internalChartVars some shared labels/selectors and names that must match in resources
type internalChartVars struct {
	Labels        *map[string]*string
	BaseName      string
	ConfigMapName string
	Props         *Props
}

func service(chart cdk8s.Chart, shared internalChartVars) {
	k8s.NewKubeService(chart, a.Str(fmt.Sprintf("%s-service", shared.BaseName)), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(shared.BaseName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("serviceport"),
					Port:       a.Num(1080),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(1080)),
				},
			},
			Selector: shared.Labels,
		},
	})
}

func deployment(chart cdk8s.Chart, shared internalChartVars) {
	k8s.NewKubeDeployment(
		chart,
		a.Str(fmt.Sprintf("%s-deployment", shared.BaseName)),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(shared.BaseName),
			},
			Spec: &k8s.DeploymentSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: shared.Labels,
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: shared.Labels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Str(shared.ConfigMapName),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Str(shared.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							container(shared),
						},
					},
				},
			},
		})
}

func container(shared internalChartVars) *k8s.Container {
	return &k8s.Container{
		Name:            a.Str(shared.BaseName),
		Image:           a.Str(fmt.Sprintf("%s:%s", "mockserver/mockserver", "mockserver-5.13.2")),
		ImagePullPolicy: a.Str("Always"),
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Str(shared.ConfigMapName),
				MountPath: a.Str("/config"),
			},
			{
				Name:      a.Str(shared.ConfigMapName),
				MountPath: a.Str("/libs"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("serviceport"),
				ContainerPort: a.Num(1080),
			},
		},
		Env: &[]*k8s.EnvVar{
			a.EnvVarStr("MOCKSERVER_LOG_LEVEL", "DEBUG"),
			a.EnvVarStr("SERVER_PORT", "1080"),
		},
		Resources: a.ContainerResources("200m", "528Mi", "200m", "528Mi"),
	}
}

func configMap(chart cdk8s.Chart, shared internalChartVars) {
	k8s.NewKubeConfigMap(chart, a.Str(shared.ConfigMapName), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(shared.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Str(shared.ConfigMapName),
			},
		},
		Data: &map[string]*string{
			"mockserver.properties": a.Str(`###############################
# MockServer & Proxy Settings #
###############################

# Socket & Port Settings

# socket timeout in milliseconds (default 120000)
mockserver.maxSocketTimeout=120000

# Json Initialization

mockserver.initializationJsonPath=/config/initializerJson.json
mockserver.watchInitializationJson=true

mockserver.livenessHttpGetPath=/liveness/probe`),
		},
	})
}

func NewChart(chart cdk8s.Chart, props *Props) cdk8s.Chart {
	s := internalChartVars{
		Labels: &map[string]*string{
			"app": a.Str("mockserver"),
		},
		ConfigMapName: "mockserver-cm",
		BaseName:      "mockserver",
		Props:         props,
	}
	configMap(chart, s)
	service(chart, s)
	deployment(chart, s)
	return chart
}
