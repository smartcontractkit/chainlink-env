package presets

import (
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/reorg"
)

// EVMOneNode local development Chainlink deployment
func EVMOneNode(config *environment.Config) (*environment.Environment, error) {
	c, err := chainlink.NewDeployment(1, nil)
	if err != nil {
		return nil, err
	}
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelmCharts(c), nil
}

// EVMMinimalLocalBS local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR) + Blockscout
func EVMMinimalLocalBS(config *environment.Config) (*environment.Environment, error) {
	c, err := chainlink.NewDeployment(5, nil)
	if err != nil {
		return nil, err
	}
	return environment.New(config).
		AddChart(blockscout.New(&blockscout.Props{})).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelmCharts(c), nil
}

// EVMMinimalLocal local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR)
func EVMMinimalLocal(config *environment.Config) *environment.Environment {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 5,
		}))
}

// EVMReorg deployment for two Ethereum networks re-org test
func EVMReorg(config *environment.Config) (*environment.Environment, error) {
	var clToml = `[[EVM]]
ChainID = '1337'
FinalityDepth = 200

[[EVM.Nodes]]
Name = 'geth'
WSURL = 'ws://geth-ethereum-geth:8546'
HTTPURL = 'http://geth-ethereum-geth:8544'

[EVM.HeadTracker]
HistoryDepth = 400`
	c, err := chainlink.NewDeployment(5, map[string]interface{}{
		"toml": clToml,
	})
	if err != nil {
		return nil, err
	}
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(reorg.New(&reorg.Props{
			NetworkName: "geth",
			NetworkType: "geth-reorg",
			Values: map[string]interface{}{
				"geth": map[string]interface{}{
					"genesis": map[string]interface{}{
						"networkId": "1337",
					},
				},
			},
		})).
		AddHelm(reorg.New(&reorg.Props{
			NetworkName: "geth-2",
			NetworkType: "geth-reorg",
			Values: map[string]interface{}{
				"geth": map[string]interface{}{
					"genesis": map[string]interface{}{
						"networkId": "2337",
					},
				},
			},
		})).
		AddHelmCharts(c), nil
}

// EVMSoak deployment for a long running soak tests
func EVMSoak(config *environment.Config) *environment.Environment {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(&ethereum.Props{
			Simulated: true,
			Values: map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "2048Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "2048Mi",
					},
				},
			},
		})).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 5,
			"db": map[string]interface{}{
				"stateful": true,
				"capacity": "1Gi",
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "250m",
						"memory": "256Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "250m",
						"memory": "256Mi",
					},
				},
			},
			"chainlink": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "2048Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "2048Mi",
					},
				},
			},
		}))
}
