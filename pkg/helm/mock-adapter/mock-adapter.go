package mock_adapter

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	URLsKey = "qa-mock-adapter"
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
	mock, err := e.Fwd.FindPort("qa-mock-adapter:0", "qa-mock-adapter", "serviceport").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	mockInternal, err := e.Fwd.FindPort("qa-mock-adapter:0", "qa-mock-adapter", "serviceport").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	if e.Cfg.InsideK8s {
		urls = append(urls, mockInternal, mockInternal)
	} else {
		urls = append(urls, mock, mockInternal)
	}
	e.URLs[URLsKey] = urls
	log.Info().Str("URL", mock).Msg("Mock adapter local connection")
	log.Info().Str("URL", mockInternal).Msg("Mock adapter remote connection")
	return nil
}

func defaultProps() map[string]interface{} {
	internalRepo := os.Getenv(config.EnvVarInternalDockerRepo)
	mockAdapterRepo := "qa-mock-adapter"
	if internalRepo != "" {
		mockAdapterRepo = fmt.Sprintf("%s/qa-mock-adapter", internalRepo)
	}

	return map[string]interface{}{
		"replicaCount": "1",
		"service": map[string]interface{}{
			"type": "NodePort",
			"port": "6060",
		},
		"app": map[string]interface{}{
			"serverPort": "6060",
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
			"repository": mockAdapterRepo,
			"snapshot":   false,
			"pullPolicy": "IfNotPresent",
		},
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	return NewVersioned("", props)
}

// NewVersioned enables choosing a specific helm chart version
func NewVersioned(helmVersion string, props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustMerge(&dp, props)
	return Chart{
		Name:    "qa-mock-adapter",
		Path:    "chainlink-qa/qa-mock-adapter",
		Values:  &dp,
		Version: helmVersion,
	}
}
