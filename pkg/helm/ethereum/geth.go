package ethereum

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	Geth     = "geth"
	External = "external"
)

type Props struct {
	NetworkName string   `envconfig:"network_name"`
	NetworkType string   `envconfig:"network_type"`
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
	switch m.Props.NetworkType {
	case Geth:
		return true
	case External:
		return false
	default:
		log.Fatal().Msg("unknown network type")
		return false
	}
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
	switch m.Props.NetworkType {
	case Geth:
		gethLocal, err := e.Fwd.FindPort("geth:0", "geth-network", "ws-rpc").As(client.LocalConnection, client.WS)
		if err != nil {
			return err
		}
		gethInternal, err := e.Fwd.FindPort("geth:0", "geth-network", "ws-rpc").As(client.RemoteConnection, client.WS)
		if err != nil {
			return err
		}
		e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], gethLocal)
		internalName := fmt.Sprintf("%s_internal", m.Props.NetworkName)
		e.URLs[internalName] = append(e.URLs[internalName], gethInternal)
		log.Info().Str("Name", "Geth").Str("URLs", gethLocal).Msg("Geth network")
	case External:
		e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkType], m.Props.WsURLs...)
		log.Info().Str("Name", m.Props.NetworkName).Strs("URLs", m.Props.WsURLs).Msg("Ethereum network")
	}
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "geth",
		NetworkType: Geth,
		Values: map[string]interface{}{
			"replicas": "1",
			"geth": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   "ethereum/client-go",
					"version": "v1.10.17",
				},
			},
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "1000m",
					"memory": "768Mi",
				},
				"limits": map[string]interface{}{
					"cpu":    "1000m",
					"memory": "768Mi",
				},
			},
		},
	}
}

func New(props *Props) environment.ConnectedChart {
	if props == nil {
		props = &Props{}
	}
	targetProps := defaultProps()
	config.MustEnvCodeOverrideStruct("ETHEREUM", targetProps, props)
	config.MustEnvCodeOverrideMap("ETHEREUM_VALUES", &targetProps.Values, props.Values)
	switch targetProps.NetworkType {
	case Geth:
		return Chart{
			HelmProps: &HelmProps{
				Name:   "geth",
				Path:   "chainlink-qa/geth",
				Values: &targetProps.Values,
			},
			Props: targetProps,
		}
	case External:
		return Chart{
			Props: targetProps,
		}
	default:
		log.Fatal().Msg("unknown Ethereum network type")
		return nil
	}
}
