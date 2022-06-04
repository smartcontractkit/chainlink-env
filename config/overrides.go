package config

import (
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
)

const (
	EnvVarCLImage            = "CHAINLINK_IMAGE"
	EnvVarCLImageDescription = "Chainlink image repository"
	EnvVarCLImageExample     = "public.ecr.aws/chainlink/chainlink"

	EnvVarCLTag            = "CHAINLINK_TAG"
	EnvVarCLTagDescription = "Chainlink image tag"
	EnvVarCLTagExample     = "1.4.0-root"

	EnvVarUser            = "CHAINLINK_ENV_USER"
	EnvVarUserDescription = "Owner of an environment"
	EnvVarUserExample     = "Satoshi"

	EnvVarLogLevel            = "ENV_LOG_LEVEL"
	EnvVarLogLevelDescription = "Environment logging level"
	EnvVarLogLevelExample     = "info | debug | trace"
)

// MustEnvOverrideStruct used when you need to override a struct with `envconfig` fields from environment variables
func MustEnvOverrideStruct(envPrefix string, s interface{}) {
	if err := envconfig.Process(envPrefix, s); err != nil {
		log.Fatal().Err(err).Send()
	}
}

// MustEnvCodeOverrideStruct used when you need to override in order
// ENV_VARS -> Code defined struct fields -> Sane defaults if struct is nil
func MustEnvCodeOverrideStruct(envPrefix string, targetVars interface{}, codeVars interface{}) {
	if err := mergo.Merge(targetVars, codeVars, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
	MustEnvOverrideStruct(envPrefix, targetVars)
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

// MustEnvCodeOverrideMap used when overriding helm charts both from env and code
func MustEnvCodeOverrideMap(envVarName string, target, src interface{}) {
	var fileVars map[string]interface{}
	if err := mergo.Merge(target, src, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
	os.Getenv(envVarName)
	fp := os.Getenv(envVarName)
	if os.Getenv(envVarName) != "" {
		path, err := filepath.Abs(fp)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		d, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		if err := yaml.Unmarshal(d, &fileVars); err != nil {
			log.Fatal().Err(err).Send()
		}
		if err := mergo.Merge(target, fileVars, mergo.WithOverride); err != nil {
			log.Fatal().Err(err).Send()
		}
	}
}
