package config

import (
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
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
