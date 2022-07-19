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
func EVMOneNode(config *environment.Config) *environment.Environment {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil))
}

// EVMMinimalLocalBS local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR) + Blockscout
func EVMMinimalLocalBS(config *environment.Config) *environment.Environment {
	return environment.New(config).
		AddChart(blockscout.New(&blockscout.Props{})).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 5,
		}))
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
func EVMReorg(config *environment.Config) *environment.Environment {
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
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 5,
			"env": map[string]interface{}{
				"eth_url": "ws://geth-reorg-ethereum-geth:8546",
			},
		}))
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
				"capacity": "30Gi",
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
