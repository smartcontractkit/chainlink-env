package main

import (
	"github.com/smartcontractkit/chainlink-env/chaos"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/reorg"
	"time"
)

func main() {
	// creates non-controlled re-org on geth
	e := environment.New(nil)
	err := e.
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(reorg.New(
			"geth-reorg",
			map[string]interface{}{
				"geth": map[string]interface{}{
					"genesis": map[string]interface{}{
						"networkId": "1337",
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
	if err != nil {
		panic(err)
	}
	expID, err := e.Chaos.Run(chaos.NewNetworkPartitionExperiment(
		e.Cfg.Namespace,
		reorg.TXNodesAppLabel,
		reorg.MinerNodesAppLabel,
	))
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Minute)
	if err = e.Chaos.Stop(expID); err != nil {
		panic(err)
	}
}
