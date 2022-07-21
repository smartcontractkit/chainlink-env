package starknet

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey = "devnet"
)

type Props struct {
	NetworkName string   `envconfig:"network_name"`
	HttpURLs    []string `envconfig:"http_url"`
	WsURLs      []string `envconfig:"ws_url"`
	Values      map[string]interface{}
}

type HelmProps struct {
	Name   string
	Path   string
	Values *map[string]interface{}
}

type Chart struct {
	HelmProps *HelmProps
	Props     *Props
}

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetName() string {
	return m.HelmProps.Name
}

func (m Chart) GetPath() string {
	return m.HelmProps.Path
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.HelmProps.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	urls := make([]string, 0)
	devnet, err := e.Fwd.FindPort("devnet:0", "devnet", "serviceport").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	devnetInternal, err := e.Fwd.FindPort("devnet:0", "devnet", "serviceport").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	if e.Cfg.InsideK8s {
		urls = append(urls, devnetInternal, devnetInternal)
	} else {
		urls = append(urls, devnet, devnetInternal)
	}
	e.URLs[URLsKey] = urls
	log.Info().Str("URL", devnet).Msg("Devnet local connection")
	log.Info().Str("URL", devnetInternal).Msg("Devnet remote connection")
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "devnet",
		Values: map[string]interface{}{
			"replicas": "1",
			"devnet": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   "shardlabs/starknet-devnet",
					"version": "v0.2.6",
				},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
				},
				"seed":      "123",
				"real_node": "false",
			},
		},
	}
}

func New(props *Props) environment.ConnectedChart {
	if props == nil {
		props = defaultProps()
	}
	return Chart{
		HelmProps: &HelmProps{
			Name:   "devnet",
			Path:   "chainlink-qa/starknet",
			Values: &props.Values,
		},
		Props: props,
	}
}
