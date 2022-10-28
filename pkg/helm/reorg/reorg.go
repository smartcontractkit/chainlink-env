package reorg

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey            = "geth"
	TXNodesAppLabel    = "geth-ethereum-geth"
	MinerNodesAppLabel = "geth-ethereum-miner-node"
)

type Props struct {
	NetworkName string `envconfig:"network_name"`
	NetworkType string `envconfig:"network_type"`
	Values      map[string]interface{}
}

type Chart struct {
	Name   string
	Path   string
	Props  *Props
	Values *map[string]interface{}
}

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetName() string {
	return m.Name
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetPath() string {
	return m.Path
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	urls := make([]string, 0)
	txPodName := fmt.Sprintf("%s-ethereum-geth:0", m.Props.NetworkName)
	miner1PodName := fmt.Sprintf("%s-ethereum-miner-node:0", m.Props.NetworkName)
	miner2PodName := fmt.Sprintf("%s-ethereum-miner-node:1", m.Props.NetworkName)
	minerPods, err := e.Client.ListPods(e.Cfg.Namespace, fmt.Sprintf("app=%s-ethereum-miner-node", m.Props.NetworkName))
	if err != nil {
		return err
	}
	txNode, err := e.Fwd.FindPort(txPodName, "geth", "ws-rpc").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	urls = append(urls, txNode)
	if len(minerPods.Items) > 0 {
		miner1, err := e.Fwd.FindPort(miner1PodName, "geth-miner", "ws-rpc-miner").As(client.LocalConnection, client.WS)
		if err != nil {
			return err
		}
		miner2, err := e.Fwd.FindPort(miner2PodName, "geth-miner", "ws-rpc-miner").As(client.LocalConnection, client.WS)
		if err != nil {
			return err
		}
		urls = append(urls, miner1, miner2)
		log.Info().Str("URL", miner1).Msg("Geth network (Miner #1)")
		log.Info().Str("URL", miner2).Msg("Geth network (Miner #2)")
	}

	e.URLs[m.Props.NetworkName] = urls
	log.Info().Str("URL", txNode).Msg("Geth network (TX Node)")
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "geth",
		NetworkType: "geth-reorg",
		Values: map[string]interface{}{
			"imagePullPolicy": "IfNotPresent",
			"bootnode": map[string]interface{}{
				"replicas": "2",
				"image": map[string]interface{}{
					"repository": "ethereum/client-go",
					"tag":        "alltools-v1.10.25",
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
					"tag":        "v1.10.25",
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
					"networkId": "1337",
				},
			},
		},
	}
}

func New(props *Props) environment.ConnectedChart {
	targetProps := defaultProps()
	config.MustMerge(targetProps, props)
	config.MustMerge(&targetProps.Values, props.Values)
	return Chart{
		Name:   targetProps.NetworkName,
		Path:   "chainlink-qa/ethereum",
		Values: &targetProps.Values,
		Props:  targetProps,
	}
}
