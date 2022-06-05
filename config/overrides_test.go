package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type Props struct {
	Name string `envconfig:"MY_NAME" yaml:"name"`
}

func TestOverrideCodeEnv(t *testing.T) {
	t.Run("defaults are there", func(t *testing.T) {
		nameBefore := "DefaultName"
		codeProps := &Props{}
		defaultCodeProps := &Props{
			Name: "DefaultName",
		}
		MustEnvCodeOverrideStruct("prefix", defaultCodeProps, codeProps)
		require.Equal(t, nameBefore, defaultCodeProps.Name)
	})
	t.Run("code can override", func(t *testing.T) {
		codeProps := &Props{
			Name: "CodeName",
		}
		defaultCodeProps := &Props{
			Name: "DefaultName",
		}
		MustEnvCodeOverrideStruct("prefix", defaultCodeProps, codeProps)
		require.Equal(t, "CodeName", defaultCodeProps.Name)
	})
	t.Run("env can override code", func(t *testing.T) {
		codeProps := &Props{
			Name: "CodeName",
		}
		defaultCodeProps := &Props{
			Name: "DefaultName",
		}
		overridenName := "EnvName"
		// nolint
		os.Setenv("PREFIX_MY_NAME", overridenName)
		MustEnvCodeOverrideStruct("PREFIX", defaultCodeProps, codeProps)
		require.Equal(t, overridenName, defaultCodeProps.Name)
	})
	t.Run("works with maps too, env vars can be overriden from file", func(t *testing.T) {
		codeProps := map[string]interface{}{
			"name": "code_name",
		}
		defaultCodeProps := map[string]interface{}{
			"name": "default_name",
		}
		// nolint
		os.Setenv("HELM_FILE_VALUES", "overrides.yaml")
		MustEnvCodeOverrideMap("HELM_FILE_VALUES", &defaultCodeProps, codeProps)
		require.Equal(t, "file_override", defaultCodeProps["name"])
	})
	t.Run("env prefix check", func(t *testing.T) {
		codeProps := &Props{
			Name: "CodeName",
		}
		defaultCodeProps := &Props{
			Name: "DefaultName",
		}
		overridenName := "EnvName"
		// nolint
		os.Setenv("PREFIX_SOME_MY_NAME", overridenName)
		MustEnvCodeOverrideStruct("prEfix_sOme", defaultCodeProps, codeProps)
		require.Equal(t, overridenName, defaultCodeProps.Name)
	})
}
