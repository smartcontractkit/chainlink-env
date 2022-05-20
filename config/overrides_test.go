package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type Props struct {
	Name string `envconfig:"MY_NAME"`
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
		os.Setenv("PREFIX_MY_NAME", overridenName)
		MustEnvCodeOverrideStruct("PREFIX", defaultCodeProps, codeProps)
		require.Equal(t, overridenName, defaultCodeProps.Name)
	})
	t.Run("env prefix check", func(t *testing.T) {
		codeProps := &Props{
			Name: "CodeName",
		}
		defaultCodeProps := &Props{
			Name: "DefaultName",
		}
		overridenName := "EnvName"
		os.Setenv("PREFIX_SOME_MY_NAME", overridenName)
		MustEnvCodeOverrideStruct("prEfix_sOme", defaultCodeProps, codeProps)
		require.Equal(t, overridenName, defaultCodeProps.Name)
	})
}
