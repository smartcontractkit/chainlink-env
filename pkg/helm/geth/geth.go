package geth

import (
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
	geth1tx, err := e.Fwd.FindPort("geth:0", "geth-network", "ws-rpc").As(client.LocalConnection, client.WS)
	if err != nil {
		return err
	}
	e.URLs[URLsKey] = append(e.URLs[URLsKey], geth1tx)
	log.Info().Str("URL", geth1tx).Msg("Geth network local connection")
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
		"replicas": "1",
		"geth": map[string]interface{}{
			"image": map[string]interface{}{
				"image":   "ethereum/client-go",
				"version": "v1.10.17",
			},
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "200m",
				"memory": "528Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "200m",
				"memory": "528Mi",
			},
		},
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvCodeOverrideMap("GETH_VALUES", &dp, props)
	return Chart{
		Name:   "geth",
		Path:   "chainlink-qa/geth",
		Values: &dp,
	}
}
