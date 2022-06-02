package mockserver

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey = "mockserver"
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
	mock, err := e.Fwd.FindPort("mockserver:0", "mockserver", "serviceport").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	mockInternal, err := e.Fwd.FindPort("mockserver:0", "mockserver", "serviceport").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	e.URLs[URLsKey] = append(e.URLs[URLsKey], mock)
	e.URLs[URLsKey] = append(e.URLs[URLsKey], mockInternal)
	log.Info().Str("URL", mock).Msg("Mockserver local connection")
	log.Info().Str("URL", mockInternal).Msg("Mockserver remote connection")
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
		"replicaCount": "1",
		"service": map[string]interface{}{
			"type": "NodePort",
			"port": "1080",
		},
		"app": map[string]interface{}{
			"logLevel":               "INFO",
			"serverPort":             "1080",
			"mountedConfigMapName":   "mockserver-config",
			"propertiesFileName":     "mockserver.properties",
			"readOnlyRootFilesystem": "false",
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "200m",
					"memory": "256Mi",
				},
				"limits": map[string]interface{}{
					"cpu":    "200m",
					"memory": "256Mi",
				},
			},
		},
		"image": map[string]interface{}{
			"repository": "mockserver",
			"snapshot":   false,
			"pullPolicy": "IfNotPresent",
		},
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvCodeOverrideMap("GETH_VALUES", &dp, props)
	return Chart{
		Name:   "mockserver",
		Path:   "chainlink-qa/mockserver",
		Values: &dp,
	}
}