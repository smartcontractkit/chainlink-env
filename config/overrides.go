package config

import (
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"reflect"
)

// MustEnvOverrideStruct used when you need to override a struct with `envconfig` fields from environment variables
func MustEnvOverrideStruct(envPrefix string, s interface{}) {
	log.Trace().
		Str("TargetKind", reflect.ValueOf(s).Kind().String()).
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
func MustEnvCodeOverrideStruct(envPrefix string, targetVars interface{}, codeVars interface{}) {
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Code", targetVars).
		Interface("Struct", codeVars).
		Msg("Overriding struct with code default vars")
	if err := mergo.Merge(targetVars, codeVars, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Trace().
		Str("Prefix", envPrefix).
		Interface("Code", targetVars).
		Interface("Struct", codeVars).
		Msg("Done overriding struct with code default vars")
	MustEnvOverrideStruct(envPrefix, targetVars)
}
