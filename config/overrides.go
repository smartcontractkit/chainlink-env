package config

import (
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// MustEnvOverrideStruct used when you need to override a struct with `envconfig` fields from environment variables
func MustEnvOverrideStruct(envPrefix string, s interface{}) {
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Struct", s).
		Msg("Overriding struct with ENV vars for envPrefix")
	if err := envconfig.Process(envPrefix, s); err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Struct", s).
		Msg("Done overriding struct with ENV vars for envPrefix")
}

// MustEnvCodeOverrideStruct used when you need to override in order
// ENV_VARS -> Code defined struct fields -> Sane defaults if struct is nil
func MustEnvCodeOverrideStruct(envPrefix string, defaults interface{}, s interface{}) {
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Code", defaults).
		Interface("Struct", s).
		Msg("Done overriding struct with code default vars for prefix")
	if err := mergo.Merge(defaults, s, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Code", defaults).
		Interface("Struct", s).
		Msg("Done overriding struct with code default vars for prefix")
	MustEnvOverrideStruct(envPrefix, defaults)
}
