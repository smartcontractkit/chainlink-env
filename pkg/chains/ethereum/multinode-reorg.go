package ethereum

import (
	cdk8s "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/smartcontractkit/chainlink-env/config"
	a "github.com/smartcontractkit/chainlink-env/pkg/alias"
)

type ReorgProps struct {
}

func defaultReorgProps() *ReorgProps {
	return &ReorgProps{}
}

func NewEthereumReorg(chart cdk8s.Chart, props *ReorgProps, constructName, namespaceName, networkID string) cdk8s.Chart {
	defaultProps := defaultReorgProps()
	config.MustEnvCodeOverrideStruct("", defaultProps, props)
	cdk8s.NewHelm(chart, a.Str(constructName), &cdk8s.HelmProps{
		Chart: a.Str("/Users/f4hrenh9it/go/src/helmenv/environment/charts/geth-reorg"),
		// to properly generate .Release.Namespace, internal vars of Helm
		HelmFlags: &[]*string{
			a.Str("--namespace"),
			a.Str(namespaceName),
		},
		ReleaseName: a.Str(constructName),
		Values: &map[string]interface{}{
			"imagePullPolicy": "IfNotPresent",
			"bootnode": map[string]interface{}{
				"replicas": "2",
				"image": map[string]interface{}{
					"repository": "ethereum/client-go",
					"tag":        "alltools-v1.10.6",
				},
			},
			"bootnodeRegistrar": map[string]interface{}{
				"replicas": "1",
				"image": map[string]interface{}{
					"repository": "jpoon/bootnode-registrar",
					"tag":        "v1.0.0",
				},
			},
			"geth": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": "ethereum/client-go",
					"tag":        "v1.10.17",
				},
				"tx": map[string]interface{}{
					"replicas": "1",
					"service": map[string]interface{}{
						"type": "ClusterIP",
					},
				},
				"miner": map[string]interface{}{
					"replicas": "2",
					"account": map[string]interface{}{
						"secret": "",
					},
				},
				"genesis": map[string]interface{}{
					"networkId": networkID,
				},
			},
		},
	})
	return chart
}
