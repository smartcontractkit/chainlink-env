package ethereum

import (
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"

	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

type Props struct {
	NetworkName string   `envconfig:"network_name"`
	Simulated   bool     `envconfig:"network_simulated"`
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
	return m.Props.Simulated
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
	if m.Props.Simulated {
		gethLocal, err := e.Fwd.FindPort("geth:0", "geth-network", "ws-rpc").As(client.LocalConnection, client.WS)
		if err != nil {
			return err
		}
		gethInternal, err := e.Fwd.FindPort("geth:0", "geth-network", "ws-rpc").As(client.RemoteConnection, client.WS)
		if err != nil {
			return err
		}
		e.NetworkURLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], gethLocal)
		if e.Cfg.InsideK8s {
			e.NetworkURLs[m.Props.NetworkName] = []string{gethInternal}
		}
		log.Info().Str("Name", "Geth").Str("URLs", gethLocal).Msg("Geth network")
	} else {
		e.NetworkURLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], m.Props.WsURLs...)
		log.Info().Str("Name", m.Props.NetworkName).Strs("URLs", m.Props.WsURLs).Msg("Ethereum network")
	}
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "Simulated Geth",
		Simulated:   true,
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
	targetProps := defaultProps()
	if props == nil {
<<<<<<< HEAD
		props = targetProps
	}
	if err := mergo.MergeWithOverwrite(targetProps, props); err != nil {
		log.Fatal().Err(err).Msg("Error merging ethereum props")
=======
		props = &Props{
			NetworkType: Geth,
		}
>>>>>>> 719651160495ccd5da04957bd97ff748c0434a31
	}
	config.MustEnvCodeOverrideMap("ETHEREUM_VALUES", &targetProps.Values, props.Values)
	targetProps.Simulated = props.Simulated // Mergo has issues with boolean merging for simulated networks
	if targetProps.Simulated {
		return Chart{
			HelmProps: &HelmProps{
				Name:   "geth",
				Path:   "chainlink-qa/geth",
				Values: &targetProps.Values,
			},
			Props: targetProps,
		}
	}
	return Chart{
		Props: targetProps,
	}
}
