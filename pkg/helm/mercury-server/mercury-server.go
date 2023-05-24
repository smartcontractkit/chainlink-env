package mercury_server

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
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
	httpLocal, err := e.Fwd.FindPort("mercury-server-rest:0", "rest-mercury-server", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	httpRemote, err := e.Fwd.FindPort("mercury-server-rest:0", "rest-mercury-server", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	wsLocal, err := e.Fwd.FindPort("mercury-server-ws:0", "ws-mercury-server", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	wsRemote, err := e.Fwd.FindPort("mercury-server-ws:0", "ws-mercury-server", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	wsrpcLocal, err := e.Fwd.FindPort("mercury-server-wsrpc:0", "wsrpc-mercury-server", "wsrpc").As(client.LocalConnection, client.WSS)
	if err != nil {
		return err
	}
	wsrpcRemote, err := e.Fwd.FindPort("mercury-server-wsrpc:0", "wsrpc-mercury-server", "wsrpc").As(client.RemoteConnection, client.WSS)
	if err != nil {
		return err
	}
	if e.Cfg.InsideK8s {
		urls = append(urls, httpLocal, httpLocal)
		urls = append(urls, wsLocal, wsLocal)
		urls = append(urls, wsrpcLocal, wsrpcLocal)
	} else {
		urls = append(urls, httpRemote, httpLocal)
		urls = append(urls, wsRemote, wsLocal)
		urls = append(urls, wsrpcRemote, wsrpcLocal)
	}
	log.Info().Str("URL", httpLocal).Msg("mercury-server http local connection")
	log.Info().Str("URL", httpRemote).Msg("mercury-server http remote connection")
	log.Info().Str("URL", wsLocal).Msg("mercury-server ws local connection")
	log.Info().Str("URL", wsRemote).Msg("mercury-server ws remote connection")
	log.Info().Str("URL", wsrpcLocal).Msg("mercury-server wsrpc local connection")
	log.Info().Str("URL", wsrpcRemote).Msg("mercury-server wsrpc remote connection")

	dbProps := (*m.Values)["postgresql"].(map[string]interface{})
	isLocalDbEnabled, ok := dbProps["enabled"].(bool)
	if ok && isLocalDbEnabled {
		dbLocal, err := e.Fwd.FindPort("mercury-server-db:0", "postgresql", "tcp-postgresql").As(client.LocalConnection, client.POSTGRESQL)
		if err != nil {
			return err
		}
		dbRemote, err := e.Fwd.FindPort("mercury-server-db:0", "postgresql", "tcp-postgresql").As(client.RemoteConnection, client.POSTGRESQL)
		if err != nil {
			return err
		}
		if e.Cfg.InsideK8s {
			urls = append(urls, dbLocal, dbLocal)
		} else {
			urls = append(urls, dbRemote, dbLocal)
		}
		log.Info().Str("URL", dbLocal).Msg("mercury-server-db local connection")
		log.Info().Str("URL", dbRemote).Msg("mercury-server-db remote connection")
	}
	e.URLs[m.Name] = urls

	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{}
}

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
