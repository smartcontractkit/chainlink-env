package schema_registry

import (
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

type Props struct {
}

type Chart struct {
	Name   string
	Path   string
	Props  *Props
	Values *map[string]interface{}
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

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	return nil
}

func defaultProps() map[string]interface{} {
	return map[string]interface{}{}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustMerge(&dp, props)
	return Chart{
		Name:   "cp-schema-registry",
		Path:   "chainlink-qa/schema-registry",
		Values: &dp,
	}
}
