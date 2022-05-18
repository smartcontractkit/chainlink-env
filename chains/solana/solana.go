package solana

import (
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"

	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
)

const (
	ChainType = "solana"
)

type Props struct{}

// SharedConstructVars some shared labels/selectors and names that must match in resources
type SharedConstructVars struct {
	SolLabels      *map[string]*string
	DeploymentName string
	ConfigMapName  string
}

func configMap(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeConfigMap(chart, a.Jss(shared.ConfigMapName), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss(shared.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Jss(shared.ConfigMapName),
			},
		},
		Data: &map[string]*string{
			"id.json": a.Jss(`[205,246,252,222,193,57,3,13,164,146,52,162,143,135,8,254,37,4,250,48,137,61,49,57,187,210,209,118,108,125,81,235,136,69,202,17,24,209,91,226,206,92,80,45,83,14,222,113,229,190,94,142,188,124,102,122,15,246,40,190,24,247,69,133]`),
			"config.yml": a.Jss(
				`json_rpc_url: http://0.0.0.0:8899
websocket_url: ws://0.0.0.0:8900
keypair_path: /root/.config/solana/cli/id.json
commitment: finalized`),
		},
	})
}

func service(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeService(chart, a.Jss("sol"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss("sol"),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Jss("ws-rpc"),
					Port:       a.Jsn(8900),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(8900)),
				},
				{
					Name:       a.Jss("http-rpc"),
					Port:       a.Jsn(8899),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(8899)),
				}},
			Selector: shared.SolLabels,
		},
	})
}

func deployment(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeDeployment(
		chart,
		a.Jss("sol-deployment"),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Jss(shared.DeploymentName),
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
								Name: a.Jss(shared.ConfigMapName),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Jss(shared.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Jss("default"),
						Containers: &[]*k8s.Container{
							container(shared),
						},
					},
				},
			},
		})
}

func container(shared SharedConstructVars) *k8s.Container {
	return &k8s.Container{
		Name:            a.Jss("sol-val"),
		Image:           a.Jss(fmt.Sprintf("%s:%s", "f4hrenh9it/sol-val", "v1")),
		ImagePullPolicy: a.Jss("Always"),
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/.config/solana/cli"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Jss("http-rpc"),
				ContainerPort: a.Jsn(8899),
			},
			{
				Name:          a.Jss("ws-rpc"),
				ContainerPort: a.Jsn(8900),
			},
		},
		Env:       &[]*k8s.EnvVar{},
		Resources: a.ContainerResources("800m", "2000Mi", "1600m", "4000Mi"),
	}
}

func NewSolanaChart(chart constructs.Construct, props *Props) constructs.Construct {
	s := SharedConstructVars{
		SolLabels: &map[string]*string{
			"app": a.Jss("sol"),
		},
		ConfigMapName:  "sol-cm",
		DeploymentName: "sol",
	}
	service(chart, s)
	configMap(chart, s)
	deployment(chart, s)
	return chart
}
