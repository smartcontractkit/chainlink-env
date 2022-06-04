package chainlink

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	AppName              = "chainlink-node"
	NodesLocalURLsKey    = "chainlink_local"
	NodesInternalURLsKey = "chainlink_internal"
	DBsLocalURLsKey      = "chainlink_db"
)

type Props struct{}

type Chart struct {
	Name   string
	Path   string
	Props  *Props
	Values *map[string]interface{}
}

func (m Chart) IsDeployed() bool {
	return true
}

func (m Chart) GetName() string {
	return m.Name
}

func (m Chart) GetPath() string {
	return m.Path
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	pods, err := e.Fwd.Client.ListPods(e.Cfg.Namespace, fmt.Sprintf("app=%s", AppName))
	if err != nil {
		return err
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "node", "access").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[NodesLocalURLsKey] = append(e.URLs[NodesLocalURLsKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("Local connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "node", "access").
			As(client.RemoteConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[NodesInternalURLsKey] = append(e.URLs[NodesInternalURLsKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("Remote (in cluster) connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", AppName, i), "chainlink-db", "postgres").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[DBsLocalURLsKey] = append(e.URLs[DBsLocalURLsKey], n)
		log.Info().Int("Node", i).Str("URL", n).Msg("DB local Connection")
	}
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
		"replicas": "1",
		"env": map[string]interface{}{
			"database_url": "postgresql://postgres:node@0.0.0.0/chainlink?sslmode=disable",
		},
		"chainlink": map[string]interface{}{
			"image": map[string]interface{}{
				"image":   "public.ecr.aws/chainlink/chainlink",
				"version": "1.4.1-root",
			},
			"web_port": "6688",
			"p2p_port": "8090",
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "350m",
					"memory": "1024Mi",
				},
				"limits": map[string]interface{}{
					"cpu":    "350m",
					"memory": "1024Mi",
				},
			},
		},
		"db": map[string]interface{}{
			"stateful": false,
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
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvOverrideVersion(&dp)
	config.MustEnvCodeOverrideMap("CL_VALUES", &dp, props)
	return Chart{
		Name:   "chainlink",
		Path:   "chainlink-qa/chainlink",
		Values: &dp,
	}
}
