package ethereum

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey         = "geth"
	InternalURLsKey = "geth_internal"
)

const (
	Geth             = "geth"
	ExternalEthereum = "external"
)

type Props struct {
	NetworkType string `envconfig:"network_type"`
	HttpURL     string `envconfig:"http_url"`
	WsURL       string `envconfig:"ws_url"`
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

func (m Chart) IsDeployed() bool {
	switch m.Props.NetworkType {
	case Geth:
		return true
	case ExternalEthereum:
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
		e.URLs[URLsKey] = append(e.URLs[URLsKey], gethLocal)
		e.URLs[InternalURLsKey] = append(e.URLs[InternalURLsKey], gethInternal)
		log.Info().Str("URL", gethLocal).Msg("Geth network")
	case ExternalEthereum:
		e.URLs[URLsKey] = append(e.URLs[URLsKey], m.Props.WsURL)
		log.Info().Str("URL", m.Props.WsURL).Msg("Ethereum network")
	}
	return nil
}

func defaultProps() *Props {
	return &Props{
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
	case ExternalEthereum:
		return Chart{
			Props: targetProps,
		}
	default:
		log.Fatal().Msg("unknown Ethereum network type")
		return nil
	}
}
