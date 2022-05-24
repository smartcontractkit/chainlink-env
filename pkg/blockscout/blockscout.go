package blockscout

import (
	"fmt"
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"github.com/smartcontractkit/chainlink-env/pkg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

type Props struct{}

// vars some shared labels/selectors and names that must match in resources
type vars struct {
	Labels        *map[string]*string
	BaseName      string
	ConfigMapName string
	Props         *Props
}

func service(chart cdk8s.Chart, vars vars) {
	k8s.NewKubeService(chart, a.Str(fmt.Sprintf("%s-service", vars.BaseName)), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(vars.BaseName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("explorer"),
					Port:       a.Num(4000),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(4000)),
				},
			},
			Selector: vars.Labels,
		},
	})
}

func postgresContainer(p vars) *k8s.Container {
	return &k8s.Container{
		Name:  a.Str(fmt.Sprintf("%s-db", p.BaseName)),
		Image: a.Str("postgres:13.6"),
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("postgres"),
				ContainerPort: a.Num(5432),
			},
		},
		Env: &[]*k8s.EnvVar{
			a.EnvVarStr("POSTGRES_PASSWORD", "postgres"),
			a.EnvVarStr("POSTGRES_DB", "blockscout"),
		},
		LivenessProbe: &k8s.Probe{
			Exec: &k8s.ExecAction{
				Command: pkg.PGIsReadyCheck()},
			InitialDelaySeconds: a.Num(60),
			PeriodSeconds:       a.Num(60),
		},
		ReadinessProbe: &k8s.Probe{
			Exec: &k8s.ExecAction{
				Command: pkg.PGIsReadyCheck()},
			InitialDelaySeconds: a.Num(2),
			PeriodSeconds:       a.Num(2),
		},
		Resources: a.ContainerResources("1000m", "2048Mi", "1000m", "2048Mi"),
	}
}

func deployment(chart cdk8s.Chart, shared vars) {
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
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							container(shared),
							postgresContainer(shared),
						},
					},
				},
			},
		})
}

func container(shared vars) *k8s.Container {
	return &k8s.Container{
		Name:            a.Str(fmt.Sprintf("%s-node", shared.BaseName)),
		Image:           a.Str("f4hrenh9it/blockscout:v1"),
		ImagePullPolicy: a.Str("Always"),
		Command:         &[]*string{a.Str(`/bin/bash`)},
		Args: &[]*string{
			a.Str("-c"),
			a.Str("mix ecto.create && mix ecto.migrate && mix phx.server"),
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("explorer"),
				ContainerPort: a.Num(4000),
			},
		},
		Env: &[]*k8s.EnvVar{
			a.EnvVarStr("MIX_ENV", "prod"),
			a.EnvVarStr("ECTO_USE_SSL", "'false'"),
			a.EnvVarStr("COIN", "DAI"),
			a.EnvVarStr("ETHEREUM_JSONRPC_VARIANT", "geth"),
			a.EnvVarStr("ETHEREUM_JSONRPC_HTTP_URL", "http://geth:8544"),
			a.EnvVarStr("ETHEREUM_JSONRPC_WS_URL", "ws://geth:8546"),
			a.EnvVarStr("DATABASE_URL", "postgresql://postgres:@localhost:5432/blockscout?ssl=false"),
		},
		Resources: a.ContainerResources("300m", "2048Mi", "300m", "2048Mi"),
	}
}

func NewChart(chart cdk8s.Chart, props *Props) cdk8s.Chart {
	s := vars{
		Labels: &map[string]*string{
			"app": a.Str("blockscout"),
		},
		ConfigMapName: "blockscout-cm",
		BaseName:      "blockscout",
		Props:         props,
	}
	service(chart, s)
	deployment(chart, s)
	return chart
}
