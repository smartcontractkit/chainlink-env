package config

import (
	"os"

	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

const (
	EnvVarNamespace            = "ENV_NAMESPACE"
	EnvVarNamespaceDescription = "Namespace name to connect to"
	EnvVarNamespaceExample     = "chainlink-test-epic"

	EnvVarCLImage            = "CHAINLINK_IMAGE"
	EnvVarCLImageDescription = "Chainlink image repository"
	EnvVarCLImageExample     = "public.ecr.aws/chainlink/chainlink"

	EnvVarCLTag            = "CHAINLINK_VERSION"
	EnvVarCLTagDescription = "Chainlink image tag"
	EnvVarCLTagExample     = "1.5.1-root"

	EnvVarUser            = "CHAINLINK_ENV_USER"
	EnvVarUserDescription = "Owner of an environment"
	EnvVarUserExample     = "Satoshi"

	EnvVarCLCommitSha            = "CHAINLINK_COMMIT_SHA"
	EnvVarCLCommitShaDescription = "The sha of the commit that you're running tests on. Mostly used for CI"
	EnvVarCLCommitShaExample     = "${{ github.sha }}"

	EnvVarTestTrigger            = "TEST_TRIGGERED_BY"
	EnvVarTestTriggerDescription = "How the test was triggered, either manual or CI."
	EnvVarTestTriggerExample     = "CI"

	EnvVarLogLevel            = "TEST_LOG_LEVEL"
	EnvVarLogLevelDescription = "Environment logging level"
	EnvVarLogLevelExample     = "info | debug | trace"

	EnvVarSlackKey            = "SLACK_API_KEY"
	EnvVarSlackKeyDescription = "The OAuth Slack API key to report tests results with"
	EnvVarSlackKeyExample     = "xoxb-example-key"

	EnvVarSlackChannel            = "SLACK_CHANNEL"
	EnvVarSlackChannelDescription = "The Slack code for the channel you want to send the notification to"
	EnvVarSlackChannelExample     = "C000000000"

	EnvVarSlackUser            = "SLACK_USER"
	EnvVarSlackUserDescription = "The Slack code for the user you want to notify"
	EnvVarSlackUserExample     = "U000000000"

	EnvVarToleration                 = "K8S_TOLERATION"
	EnvVarTolerationsUserDescription = "Node roles to tolerate"
	EnvVarTolerationsExample         = "foundations"

	EnvVarNodeSelector                = "K8S_NODE_SELECTOR"
	EnvVarNodeSelectorUserDescription = "Node role to deploy to"
	EnvVarNodeSelectorExample         = "foundations"
)

func MustMerge(targetVars interface{}, codeVars interface{}) {
	if err := mergo.Merge(targetVars, codeVars, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func MustEnvOverrideVersion(target interface{}) {
	image := os.Getenv(EnvVarCLImage)
	tag := os.Getenv(EnvVarCLTag)
	if image != "" && tag != "" {
		if err := mergo.Merge(target, map[string]interface{}{
			"chainlink": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   image,
					"version": tag,
				},
			},
		}, mergo.WithOverride); err != nil {
			log.Fatal().Err(err).Send()
		}
	}
}
