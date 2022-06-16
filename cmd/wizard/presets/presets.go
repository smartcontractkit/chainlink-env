package presets

import (
	cfg "github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/cdk8s/blockscout"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/reorg"
)

// EVMOneNode local development Chainlink deployment
func EVMOneNode(config *environment.Config) error {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(nil)).
		Run()
}

// EVMMinimalLocalBS local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR) + Blockscout
func EVMMinimalLocalBS(config *environment.Config) error {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddChart(blockscout.New(&blockscout.Props{})).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(map[string]interface{}{
			"replicas": 5,
		})).
		Run()
}

// EVMMinimalLocal local development Chainlink deployment,
// 1 bootstrap + 4 oracles (minimal requirements for OCR)
func EVMMinimalLocal(config *environment.Config) error {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(map[string]interface{}{
			"replicas": 5,
		})).
		Run()
}

// MultiNetwork local development Chainlink deployment for multiple networks
func MultiNetwork(config *environment.Config, opts *MultiNetworkOpts) error {
	cfg.MustEnvOverrideStruct("", opts)
	e := environment.New(config)
	e.AddHelm(mockservercfg.New(nil)).AddHelm(mockserver.New(nil))
	for _, net := range opts.Networks {
		e.AddHelm(ethereum.New(&ethereum.Props{
			NetworkName: net.Name,
			NetworkType: net.Type,
			HttpURLs:    net.HttpURLs,
			WsURLs:      net.WsURLs,
		}))
	}
	// TODO: make proper configuration for all networks after config refactoring,
	// TODO: configuration for 1+ networks will change soon to TOML
	clVars := map[string]interface{}{
		"env": map[string]interface{}{
			"eth_http_url": opts.Networks[0].HttpURLs[0],
			"eth_url":      opts.Networks[0].WsURLs[0],
			"eth_chain_id": opts.Networks[0].ChainID,
		},
	}
	return e.AddHelm(chainlink.New(clVars)).Run()
}

// EVMReorg deployment for two Ethereum networks re-org test
func EVMReorg(config *environment.Config) error {
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
		AddHelm(chainlink.New(map[string]interface{}{
			"replicas": 5,
			"env": map[string]interface{}{
				"eth_url": "ws://geth-reorg-ethereum-geth:8546",
			},
		})).
		Run()
}

// EVMSoak deployment for a long running soak tests
func EVMSoak(config *environment.Config) error {
	return environment.New(config).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(&ethereum.Props{
			NetworkType: ethereum.Geth,
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
		AddHelm(chainlink.New(map[string]interface{}{
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
		})).
		Run()
}
