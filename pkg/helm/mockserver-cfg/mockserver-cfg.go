package mockserver_cfg

import (
	"github.com/smartcontractkit/chainlink-env/environment"
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
	return nil
}

func New(props map[string]interface{}) environment.ConnectedChart {
	return Chart{
		Name:   "mockserver-cfg",
		Path:   "chainlink-qa/mockserver-config",
		Values: &props,
	}
}
