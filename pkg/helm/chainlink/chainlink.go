package chainlink

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	AppName              = "chainlink"
	NodesLocalURLsKey    = "chainlink_local"
	NodesInternalURLsKey = "chainlink_internal"
	DBsLocalURLsKey      = "chainlink_db"
)

type Props struct{}

type Chart struct {
	Name    string
	Index   int
	Path    string
	Version string
	Props   *Props
	Values  *map[string]interface{}
}

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetName() string {
	return m.Name
}

func (m Chart) GetPath() string {
	return m.Path
}

func (m Chart) GetVersion() string {
	return m.Version
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	// fetching all apps with label app=chainlink-${deploymentIndex}:${instanceIndex}
	pods, err := e.Fwd.Client.ListPods(e.Cfg.Namespace, fmt.Sprintf("app=%s", m.Name))
	if err != nil {
		return err
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", m.Name, i), "node", "access").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[NodesLocalURLsKey] = append(e.URLs[NodesLocalURLsKey], n)
		log.Info().Str("Deployment", m.Name).Int("Node", i).Str("URL", n).Msg("Local connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", m.Name, i), "node", "access").
			As(client.RemoteConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[NodesInternalURLsKey] = append(e.URLs[NodesInternalURLsKey], n)
		if e.Cfg.InsideK8s {
			e.URLs[NodesLocalURLsKey] = e.URLs[NodesInternalURLsKey]
		}
		log.Info().Str("Deployment", m.Name).Int("Node", i).Str("URL", n).Msg("Remote (in cluster) connection")
	}
	for i := 0; i < len(pods.Items); i++ {
		n, err := e.Fwd.FindPort(fmt.Sprintf("%s:%d", m.Name, i), "chainlink-db", "postgres").
			As(client.LocalConnection, client.HTTP)
		if err != nil {
			return err
		}
		e.URLs[DBsLocalURLsKey] = append(e.URLs[DBsLocalURLsKey], n)
		log.Info().Str("Deployment", m.Name).Int("Node", i).Str("URL", n).Msg("DB local Connection")
	}
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
		"replicas": "1",
		"env": map[string]interface{}{
			"CL_DATABASE_URL": "postgresql://postgres:verylongdatabasepassword@0.0.0.0/chainlink?sslmode=disable",
		},
		"chainlink": map[string]interface{}{
			"image": map[string]interface{}{
				"image":   "public.ecr.aws/chainlink/chainlink",
				"version": "1.9.0",
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

func New(index int, props map[string]interface{}) environment.ConnectedChart {
	return NewVersioned(index, "", props)
}

// NewVersioned enables you to select a specific helm chart version
func NewVersioned(index int, helmVersion string, props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvOverrideVersion(&dp)
	config.MustMerge(&dp, props)
	return Chart{
		Index:   index,
		Name:    fmt.Sprintf("%s-%d", AppName, index),
		Path:    "chainlink-qa/chainlink",
		Version: helmVersion,
		Values:  &dp,
	}
}
