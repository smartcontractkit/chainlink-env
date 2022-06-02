package chaos

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/smartcontractkit/chainlink-env/imports/k8s/networkchaos/chaosmeshorg"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

func blankManifest(namespace string) (cdk8s.App, cdk8s.Chart) {
	app := cdk8s.NewApp(&cdk8s.AppProps{
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_APP,
	})
	return app, cdk8s.NewChart(app, a.Str("root"), &cdk8s.ChartProps{
		Namespace: a.Str(namespace),
	})
}

func NewNetworkPartitionExperiment(namespace string, fromApp string, toApp string) (cdk8s.App, string, string) {
	app, root := blankManifest(namespace)
	c := chaosmeshorg.NewNetworkChaos(root, a.Str("experiment"), &chaosmeshorg.NetworkChaosProps{
		Spec: &chaosmeshorg.NetworkChaosSpec{
			Action: chaosmeshorg.NetworkChaosSpecAction_PARTITION,
			Mode:   chaosmeshorg.NetworkChaosSpecMode_ALL,
			Selector: &chaosmeshorg.NetworkChaosSpecSelector{
				LabelSelectors: &map[string]*string{"app": a.Str(fromApp)},
			},
			Direction:       chaosmeshorg.NetworkChaosSpecDirection_BOTH,
			Duration:        a.Str("999h"),
			ExternalTargets: nil,
			Loss: &chaosmeshorg.NetworkChaosSpecLoss{
				Loss: a.Str("100"),
			},
			Target: &chaosmeshorg.NetworkChaosSpecTarget{
				Mode: chaosmeshorg.NetworkChaosSpecTargetMode_ALL,
				Selector: &chaosmeshorg.NetworkChaosSpecTargetSelector{
					LabelSelectors: &map[string]*string{"app": a.Str(toApp)},
				},
			},
		},
	})
	return app, *c.Name(), "networkchaos"
}
