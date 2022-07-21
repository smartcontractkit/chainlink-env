package starknet

import (
	"github.com/smartcontractkit/chainlink-env/environment"
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
