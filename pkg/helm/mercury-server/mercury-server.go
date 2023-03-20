package mercury_server

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey = "mercury-server"
)

type Props struct {
}

type Chart struct {
	Name    string
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
	urls := make([]string, 0)
	httpLocal, err := e.Fwd.FindPort("mercury-server:0", "mercury-server", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	httpRemote, err := e.Fwd.FindPort("mercury-server:0", "mercury-server", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	wsrpcLocal, err := e.Fwd.FindPort("mercury-server:0", "mercury-server", "ws-rpc").As(client.LocalConnection, client.WSS)
	if err != nil {
		return err
	}
	wsrpcRemote, err := e.Fwd.FindPort("mercury-server:0", "mercury-server", "ws-rpc").As(client.RemoteConnection, client.WSS)
	if err != nil {
		return err
	}
	if e.Cfg.InsideK8s {
		urls = append(urls, httpLocal, httpLocal)
		urls = append(urls, wsrpcLocal, wsrpcLocal)
	} else {
		urls = append(urls, httpRemote, httpLocal)
		urls = append(urls, wsrpcRemote, wsrpcLocal)

	}
	e.URLs[URLsKey] = urls
	log.Info().Str("URL", httpLocal).Msg("mercury-server local connection")
	log.Info().Str("URL", httpRemote).Msg("mercury-server remote connection")
	log.Info().Str("URL", wsrpcLocal).Msg("mercury-server wsrpc local connection")
	log.Info().Str("URL", wsrpcRemote).Msg("mercury-server wsrpc remote connection")

	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{}
}

// NewVersioned enables choosing a specific helm chart version

func New(path string, helmVersion string, props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustMerge(&dp, props)
	return Chart{
		Name:    "mercury-server",
		Path:    path,
		Values:  &dp,
		Version: helmVersion,
	}
}
