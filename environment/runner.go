package environment

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
	"github.com/stretchr/testify/require"
)

type Chart struct {
	Props *Props
}

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetName() string {
	return "remote-test-runner"
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetPath() string {
	return ""
}

func (m Chart) GetVersion() string {
	return ""
}

func (m Chart) GetValues() *map[string]interface{} {
	return nil
}

func (m Chart) ExportData(e *Environment) error {
	return nil
}

func NewRunner(props *Props) func(root cdk8s.Chart) ConnectedChart {
	setupShutdown(props)
	return func(root cdk8s.Chart) ConnectedChart {
		c := &Chart{
			Props: props,
		}
		role(root, props)
		job(root, props)
		return c
	}
}

type Props struct {
	BaseName         string
	TargetNamespace  string
	Labels           *map[string]*string
	Image            string
	Test             *testing.T
	Environment      *Environment
	fundReturnFailed bool
}

func role(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeRole(
		chart,
		a.Str(fmt.Sprintf("%s-role", props.BaseName)),
		&k8s.KubeRoleProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(props.BaseName),
			},
			Rules: &[]*k8s.PolicyRule{
				{
					ApiGroups: &[]*string{
						a.Str(""), // this empty line is needed or k8s get really angry
						a.Str("apps"),
						a.Str("batch"),
						a.Str("core"),
						a.Str("networking.k8s.io"),
						a.Str("storage.k8s.io"),
						a.Str("policy"),
						a.Str("chaos-mesh.org"),
						a.Str("monitoring.coreos.com"),
					},
					Resources: &[]*string{
						a.Str("*"),
					},
					Verbs: &[]*string{
						a.Str("*"),
					},
				},
			},
		})
	k8s.NewKubeRoleBinding(
		chart,
		a.Str(fmt.Sprintf("%s-role-binding", props.BaseName)),
		&k8s.KubeRoleBindingProps{
			RoleRef: &k8s.RoleRef{
				ApiGroup: a.Str("rbac.authorization.k8s.io"),
				Kind:     a.Str("Role"),
				Name:     a.Str("remote-test-runner"),
			},
			Metadata: nil,
			Subjects: &[]*k8s.Subject{
				{
					Kind:      a.Str("ServiceAccount"),
					Name:      a.Str("default"),
					Namespace: a.Str(props.TargetNamespace),
				},
			},
		},
	)
}

func job(chart cdk8s.Chart, props *Props) {
	k8s.NewKubeJob(
		chart,
		a.Str(fmt.Sprintf("%s-job", props.BaseName)),
		&k8s.KubeJobProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(props.BaseName),
			},
			Spec: &k8s.JobSpec{
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: props.Labels,
					},
					Spec: &k8s.PodSpec{
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							container(props),
						},
						RestartPolicy: a.Str("Never"),
					},
				},
				ActiveDeadlineSeconds: nil,
				BackoffLimit:          a.Num(0),
			},
		})
}

func container(props *Props) *k8s.Container {
	cpu := os.Getenv(config.EnvVarRemoteRunnerCpu)
	if cpu == "" {
		cpu = "1000m"
	}
	mem := os.Getenv(config.EnvVarRemoteRunnerMem)
	if mem == "" {
		mem = "1024Mi"
	}
	return &k8s.Container{
		Name:            a.Str(fmt.Sprintf("%s-node", props.BaseName)),
		Image:           a.Str(props.Image),
		ImagePullPolicy: a.Str("Always"),
		Env:             jobEnvVars(props),
		Resources:       a.ContainerResources(cpu, mem, cpu, mem),
	}
}

func jobEnvVars(props *Props) *[]*k8s.EnvVar {
	cdk8sVars := make([]*k8s.EnvVar, 0)
	cdk8sVars = append(cdk8sVars, a.EnvVarStr(config.EnvVarNamespace, props.TargetNamespace))
	cdk8sVars = append(cdk8sVars, a.EnvVarStr("TEST_NAME", props.Test.Name()))
	cdk8sVars = append(cdk8sVars, a.EnvVarStr(config.EnvVarInsideK8s, "true"))
	cdk8sVars = append(cdk8sVars, a.EnvVarStr(config.EnvVarNoManifestUpdate, strconv.FormatBool(props.Environment.Cfg.NoManifestUpdate)))
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			if strings.HasPrefix(e[:i], config.EnvVarPrefix) {
				withoutPrefix := strings.Replace(e[:i], config.EnvVarPrefix, "", 1)
				cdk8sVars = append(cdk8sVars, a.EnvVarStr(withoutPrefix, e[i+1:]))
			}
		}
	}
	// add important variables
	lookups := []string{
		config.EnvVarCLImage,
		config.EnvVarCLTag,
		config.EnvVarCLCommitSha,
		config.EnvVarLogLevel,
		config.EnvVarTestTrigger,
		config.EnvVarToleration,
		config.EnvVarSlackChannel,
		config.EnvVarSlackKey,
		config.EnvVarSlackUser,
		config.EnvVarPyroscopeKey,
		config.EnvVarPyroscopeEnvironment,
		config.EnvVarPyroscopeServer,
		config.EnvVarUser,
		config.EnvVarNodeSelector,
		config.EnvVarSelectedNetworks,
		config.EnvVarDBURL,
		config.EnvVarEVMKeys,
	}
	for _, k := range lookups {
		v, success := os.LookupEnv(k)
		if success && len(v) > 0 {
			cdk8sVars = append(cdk8sVars, a.EnvVarStr(k, v))
		}
	}

	return &cdk8sVars
}

func setupShutdown(props *Props) {
	// setup test cleanup
	// if not in detached mode
	// and not using an existing environment
	if !props.Environment.Cfg.DetachRunner && !props.Environment.Cfg.NoManifestUpdate {
		if props.Test != nil {
			props.Test.Cleanup(func() {
				if props.fundReturnFailed {
					err := props.Environment.Shutdown()
					require.NoError(props.Test, err)
				}
			})
		}
	}
}
