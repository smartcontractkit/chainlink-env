package remotetestrunner

import (
	"os"

	"github.com/rs/zerolog/log"
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

func (m Chart) GetProps() interface{} {
	return m.Props
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
	slackKey := os.Getenv(config.EnvVarSlackKey)
	slackChannel := os.Getenv(config.EnvVarSlackChannel)
	slackUser := os.Getenv(config.EnvVarSlackUser)
	if slackKey == "" {
		log.Warn().Msg("SLACK_API_KEY not set, the test won't be able to report results to Slack")
	}
	if slackChannel == "" {
		log.Warn().Msg("SLACK_CHANNEL not set, the test won't be able to report results to Slack")
	}
	if slackUser == "" {
		log.Warn().Msg("SLACK_USER not set, the test may not be able to report results to Slack")
	}
	log.Info().Str("API Key", slackKey).Str("Channel", slackChannel).Str("User", slackUser).Msg("Using Slack Creds")
	return map[string]interface{}{
		"remote_test_runner": map[string]interface{}{
			"slack_api":     slackKey,
			"slack_channel": slackChannel,
			"slack_user_id": slackUser,
			"access_port":   8080,
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "2000m",
				"memory": "1536Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "2000m",
				"memory": "1536Mi",
			},
		},
	}
}

func New(props map[string]interface{}) environment.ConnectedChart {
	dp := defaultProps()
	config.MustMerge(&dp, props)
	return Chart{
		Name:   "remote-test-runner",
		Path:   "chainlink-qa/remote-test-runner",
		Values: &dp,
	}
}

// NewLocal Use a local chart path instead of a remote one, very useful for debugging chart changes
func NewLocal(props map[string]interface{}, chartPath string) environment.ConnectedChart {
	dp := defaultProps()
	config.MustMerge(&dp, props)
	return Chart{
		Name:   "remote-test-runner",
		Path:   chartPath,
		Values: &dp,
	}
}
