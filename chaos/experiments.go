package chaos

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	networkChaos "github.com/smartcontractkit/chainlink-env/imports/k8s/networkchaos/chaosmeshorg"
	podChaos "github.com/smartcontractkit/chainlink-env/imports/k8s/podchaos/chaosmeshorg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
	"strconv"
)

var (
	FOREVER      = a.Str("999h")
	experimentNr = 0
)

type ManifestFunc func(namespace string, props *Props) (cdk8s.App, string, string)

type Props struct {
	LabelsSelector *map[string]*string
	ContainerNames *[]*string
	DurationStr    string
	FromLabels     *map[string]*string
	ToLabels       *map[string]*string
}

func blankManifest(namespace string) (cdk8s.App, cdk8s.Chart) {
	app := cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	return app, cdk8s.NewChart(app, a.Str("root"), &cdk8s.ChartProps{
		Namespace: a.Str(namespace),
	})
}

func NewKillPods(namespace string, props *Props) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	experimentNr = experimentNr + 1
	c := podChaos.NewPodChaos(root, a.Str("experiment-"+strconv.Itoa(experimentNr)), &podChaos.PodChaosProps{
		Spec: &podChaos.PodChaosSpec{
			Action: podChaos.PodChaosSpecAction_POD_KILL,
			Mode:   podChaos.PodChaosSpecMode_ALL,
			Selector: &podChaos.PodChaosSpecSelector{
				LabelSelectors: props.LabelsSelector,
			},
			Duration: FOREVER,
		},
	})
	return app, *c.Name(), "podchaos"
}

func NewFailPods(namespace string, props *Props) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	experimentNr = experimentNr + 1
	c := podChaos.NewPodChaos(root, a.Str("experiment-"+strconv.Itoa(experimentNr)), &podChaos.PodChaosProps{
		Spec: &podChaos.PodChaosSpec{
			Action: podChaos.PodChaosSpecAction_POD_FAILURE,
			Mode:   podChaos.PodChaosSpecMode_ALL,
			Selector: &podChaos.PodChaosSpecSelector{
				LabelSelectors: props.LabelsSelector,
			},
			Duration: a.Str(props.DurationStr),
		},
	})
	return app, *c.Name(), "podchaos"
}

func NewFailContainers(namespace string, props *Props) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	experimentNr = experimentNr + 1
	c := podChaos.NewPodChaos(root, a.Str("experiment-"+strconv.Itoa(experimentNr)), &podChaos.PodChaosProps{
		Spec: &podChaos.PodChaosSpec{
			Action: podChaos.PodChaosSpecAction_POD_KILL,
			Mode:   podChaos.PodChaosSpecMode_ALL,
			Selector: &podChaos.PodChaosSpecSelector{
				LabelSelectors: props.LabelsSelector,
			},
			ContainerNames: props.ContainerNames,
			Duration:       FOREVER,
		},
	})
	return app, *c.Name(), "podchaos"
}

func NewContainerKill(namespace string, props *Props) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	experimentNr = experimentNr + 1
	c := podChaos.NewPodChaos(root, a.Str("experiment-"+strconv.Itoa(experimentNr)), &podChaos.PodChaosProps{
		Spec: &podChaos.PodChaosSpec{
			Action: podChaos.PodChaosSpecAction_POD_KILL,
			Mode:   podChaos.PodChaosSpecMode_ALL,
			Selector: &podChaos.PodChaosSpecSelector{
				LabelSelectors: props.LabelsSelector,
			},
			Duration: FOREVER,
		},
	})
	return app, *c.Name(), "podchaos"
}

func NewNetworkPartition(namespace string, props *Props) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	experimentNr = experimentNr + 1
	c := networkChaos.NewNetworkChaos(root, a.Str("experiment-"+strconv.Itoa(experimentNr)), &networkChaos.NetworkChaosProps{
		Spec: &networkChaos.NetworkChaosSpec{
			Action: networkChaos.NetworkChaosSpecAction_PARTITION,
			Mode:   networkChaos.NetworkChaosSpecMode_ALL,
			Selector: &networkChaos.NetworkChaosSpecSelector{
				LabelSelectors: props.FromLabels,
			},
			Direction:       networkChaos.NetworkChaosSpecDirection_BOTH,
			Duration:        a.Str(props.DurationStr),
			ExternalTargets: nil,
			Loss: &networkChaos.NetworkChaosSpecLoss{
				Loss: a.Str("100"),
			},
			Target: &networkChaos.NetworkChaosSpecTarget{
				Mode: networkChaos.NetworkChaosSpecTargetMode_ALL,
				Selector: &networkChaos.NetworkChaosSpecTargetSelector{
					LabelSelectors: props.ToLabels,
				},
			},
		},
	})
	return app, *c.Name(), "networkchaos"
}
