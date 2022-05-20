package testrunner

import (
	"fmt"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
)

type Props struct {
	TestTag            string
	ConfigFileContents string
	SlackAPIKey        string
	SlackChannel       string
	SlackUserID        string
	TestBinarySize     float64
	AccessPort         float64
}

func DefaultTestrunnerProps() *Props {
	return &Props{}
}

// SharedConstructVars some shared labels/selectors and names that must match in resources
type SharedConstructVars struct {
	Labels        *map[string]*string
	BaseName      string
	ConfigMapName string
	Props         *Props
}

func configMap(chart cdk8s.Chart, shared SharedConstructVars) {
	k8s.NewKubeConfigMap(chart, a.Str(shared.ConfigMapName), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(shared.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Str(shared.ConfigMapName),
			},
		},
		Data: &map[string]*string{
			"test-env.json": a.Str(shared.Props.ConfigFileContents),
			"init.sh": a.Str(`#!/bin/sh

    echo "Installing dependencies"
    apk add build-base

    echo "Waiting for $TEST_FILE to start being copied"
    until [ -f $TEST_FILE ]
    do
      sleep 1
    done
    echo "Found $TEST_FILE"

    CURRENT_FILE_SIZE=$(stat -c%s "$TEST_FILE")
    until [ "$CURRENT_FILE_SIZE" -eq "$TEST_FILE_SIZE" ]
    do
      CURRENT_FILE_SIZE=$(stat -c%s "$TEST_FILE")
      echo "Copied $CURRENT_FILE_SIZE/$TEST_FILE_SIZE bytes"
      sleep 1
    done

    chmod +x $TEST_FILE
    echo "File details"
    stat $TEST_FILE
    echo "-----------"
    file $TEST_FILE
    echo "-----------"
    ldd $TEST_FILE

    echo "Copied $TEST_FILE Successfully! Running test..."
    $TEST_FILE --ginkgo.timeout=0 --ginkgo.v --ginkgo.focus {{ .Values.remote_test_runner.test_name }}`),
		},
	})
}

func service(chart cdk8s.Chart, shared SharedConstructVars) {
	k8s.NewKubeService(chart, a.Str(fmt.Sprintf("%s-service", shared.BaseName)), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(shared.BaseName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("access"),
					Port:       a.Num(shared.Props.AccessPort),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(shared.Props.AccessPort)),
				},
			},
			Selector: shared.Labels,
		},
	})
}

func deployment(chart cdk8s.Chart, shared SharedConstructVars) {
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

func container(shared SharedConstructVars) *k8s.Container {
	return &k8s.Container{
		Name:            a.Str(shared.BaseName),
		Image:           a.Str(fmt.Sprintf("%s:%s", "ethereum/client-go", "v1.10.17")),
		ImagePullPolicy: a.Str("Always"),
		Command: &[]*string{
			a.Str(`sh`),
			a.Str(`./root/init.sh`),
		},
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Str(shared.ConfigMapName),
				MountPath: a.Str("/root/init.sh"),
				SubPath:   a.Str("init.sh"),
			},
			{
				Name:      a.Str(shared.ConfigMapName),
				MountPath: a.Str("/root/test-env.json"),
				SubPath:   a.Str("test-env.json"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("access"),
				ContainerPort: a.Num(shared.Props.AccessPort),
			},
		},
		Env: &[]*k8s.EnvVar{
			a.EnvVarStr("ENVIRONMENT_FILE", "/root/test-env.json"),
			a.EnvVarStr("SLACK_API", shared.Props.SlackAPIKey),
			a.EnvVarStr("SLACK_CHANNEL", shared.Props.SlackChannel),
			a.EnvVarStr("SLACK_USER_ID", shared.Props.SlackUserID),
			a.EnvVarStr("ACCESS_PORT", fmt.Sprintf("%f", shared.Props.AccessPort)),
			a.EnvVarStr("GOOS", "linux"),
			a.EnvVarStr("GOARCH", "amd64"),
			a.EnvVarStr("FRAMEWORK_CONFIG_FILE", "/root/framework.yaml"),
			a.EnvVarStr("NETWORKS_CONFIG_FILE", "/root/networks.yaml"),
			a.EnvVarStr("TEST_FILE", "/root/remote.test"),
			a.EnvVarStr("TEST_FILE_SIZE", fmt.Sprintf("%f", shared.Props.TestBinarySize)),
		},
		Resources: a.ContainerResources("200m", "528Mi", "200m", "528Mi"),
	}
}

func NewTestRunnerChart(chart cdk8s.Chart, props *Props) cdk8s.Chart {
	s := SharedConstructVars{
		Labels: &map[string]*string{
			"app": a.Str("testrunner"),
		},
		ConfigMapName: "testrunner-cm",
		BaseName:      "testrunner",
		Props:         props,
	}
	service(chart, s)
	configMap(chart, s)
	deployment(chart, s)
	return chart
}
