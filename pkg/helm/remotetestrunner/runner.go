package remotetestrunner

import (
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/environment"
)

const (
	URLsKey = "remote-test-runner"
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

func defaultProps() map[string]interface{} {
	return map[string]interface{}{
		"remote_test_runner": map[string]interface{}{
			"test_name":        "@soak-ocr",
			"slack_api":        "default",
			"slack_channel":    "default",
			"slack_user_id":    "default",
			"remote_test_size": 0,
			"access_port":      8080,
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "250m",
				"memory": "512Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "512Mi",
			},
		},
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustEnvCodeOverrideMap("REMOTE_TEST_RUNNER_VALUES", &dp, props)
	return Chart{
		Name:   "remote-test-runner",
		Path:   "chainlink-qa/remote-test-runner",
		Values: &dp,
	}
}
