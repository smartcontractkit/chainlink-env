package reorg

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey = "geth"
)

type Chart struct {
	Name   string
	Path   string
	Values *map[string]interface{}
}

func (m Chart) GetName() string {
	return m.Name
}

func (m Chart) GetPath() string {
	return m.Path
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	fullName := fmt.Sprintf("%s-ethereum-geth:0", m.Name)
	geth1tx, err := e.Fwd.FindPort(fullName, "geth", "ws-rpc").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	e.URLs[URLsKey] = append(e.URLs[URLsKey], geth1tx)
	log.Info().Str("URL", geth1tx).Msg("Geth network one (TX Node)")
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
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
	}
}

func New(name string, props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvCodeOverrideMap("REORG_VALUES", &dp, props)
	return Chart{
		Name:   name,
		Path:   "chainlink-qa/ethereum",
		Values: &dp,
	}
}
