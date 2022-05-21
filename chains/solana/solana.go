package solana

import (
	"fmt"
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"

	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
)

type Props struct{}

// internalChartVars some shared labels/selectors and names that must match in resources
type internalChartVars struct {
	SolLabels      *map[string]*string
	DeploymentName string
	ConfigMapName  string
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
			"id.json": a.Str(`[205,246,252,222,193,57,3,13,164,146,52,162,143,135,8,254,37,4,250,48,137,61,49,57,187,210,209,118,108,125,81,235,136,69,202,17,24,209,91,226,206,92,80,45,83,14,222,113,229,190,94,142,188,124,102,122,15,246,40,190,24,247,69,133]`),
			"config.yml": a.Str(
				`json_rpc_url: http://0.0.0.0:8899
websocket_url: ws://0.0.0.0:8900
keypair_path: /root/.config/solana/cli/id.json
commitment: finalized`),
		},
	})
}

func service(chart cdk8s.Chart, shared internalChartVars) {
	k8s.NewKubeService(chart, a.Str("sol"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str("sol"),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("ws-rpc"),
					Port:       a.Num(8900),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(8900)),
				},
				{
					Name:       a.Str("http-rpc"),
					Port:       a.Num(8899),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(8899)),
				}},
			Selector: shared.SolLabels,
		},
	})
}

func deployment(chart cdk8s.Chart, shared internalChartVars) {
	k8s.NewKubeDeployment(
		chart,
		a.Str("sol-deployment"),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(shared.DeploymentName),
			},
			Spec: &k8s.DeploymentSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: shared.SolLabels,
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: shared.SolLabels,
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
		Name:            a.Str("sol-val"),
		Image:           a.Str(fmt.Sprintf("%s:%s", "f4hrenh9it/sol-val", "v1")),
		ImagePullPolicy: a.Str("Always"),
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Str(shared.ConfigMapName),
				MountPath: a.Str("/root/.config/solana/cli"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("http-rpc"),
				ContainerPort: a.Num(8899),
			},
			{
				Name:          a.Str("ws-rpc"),
				ContainerPort: a.Num(8900),
			},
		},
		Env:       &[]*k8s.EnvVar{},
		Resources: a.ContainerResources("800m", "2000Mi", "1600m", "4000Mi"),
	}
}

func NewSolana(chart cdk8s.Chart, props *Props) cdk8s.Chart {
	s := internalChartVars{
		SolLabels: &map[string]*string{
			"app": a.Str("sol"),
		},
		ConfigMapName:  "sol-cm",
		DeploymentName: "sol",
	}
	service(chart, s)
	configMap(chart, s)
	deployment(chart, s)
	return chart
}
