package reorg

import (
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
	txNode, err := e.Fwd.FindPort("geth-ethereum-geth:0", "geth", "ws-rpc").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	miner1, err := e.Fwd.FindPort("geth-ethereum-miner-node:0", "geth-miner", "ws-rpc-miner").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	miner2, err := e.Fwd.FindPort("geth-ethereum-miner-node:1", "geth-miner", "ws-rpc-miner").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	urls = append(urls, txNode, miner1, miner2)
	e.URLs[m.Props.NetworkName] = urls
	log.Info().Str("URL", txNode).Msg("Geth network (TX Node)")
	log.Info().Str("URL", miner1).Msg("Geth network (Miner #1)")
	log.Info().Str("URL", miner2).Msg("Geth network (Miner #2)")
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
					"networkId": "1337",
				},
			},
		},
	}
}

func New(props *Props) environment.ConnectedChart {
	targetProps := defaultProps()
	config.MustEnvCodeOverrideStruct("ETHEREUM", targetProps, props)
	config.MustEnvCodeOverrideMap("ETHEREUM_VALUES", &targetProps.Values, props.Values)
	return Chart{
		Name:   targetProps.NetworkName,
		Path:   "chainlink-qa/ethereum",
		Values: &targetProps.Values,
		Props:  targetProps,
	}
}
