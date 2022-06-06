package presets_test

import (
	"github.com/smartcontractkit/chainlink-env/cmd/wizard/presets"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/logging"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestOverrideCodeEnv(t *testing.T) {
	t.Run("networks override with base64", func(t *testing.T) {
		logging.Init()
		codeProps := &presets.MultiNetworkOpts{
			Networks: presets.Networks{
				{
					Name:     "codeName",
					Type:     "external",
					HttpURLs: []string{"defaultHttp"},
					WsURLs:   []string{"defaultWS"},
				},
			},
		}
		defaultCodeProps := &presets.MultiNetworkOpts{
			Networks: presets.Networks{
				{
					Name:     "default",
					Type:     "external",
					HttpURLs: []string{"defaultHttp"},
					WsURLs:   []string{"defaultWS"},
				},
			},
		}
		cfgBase64 := "base64:LSBuYW1lOiBmaWxlTmFtZQogIHR5cGU6IGZpbGVUeXBlCiAgaHR0cF91cmxzOgogICAgLSAiZmlsZVVSTDEiCiAgd3NfdXJsczoKICAgIC0gImZpbGVVUkwyIgogIGNoYWluX2lkOiAzMzc=\n"
		// nolint
		os.Setenv("NETWORKS_CONFIG_FILE", cfgBase64)
		config.MustEnvCodeOverrideStruct("", defaultCodeProps, codeProps)
		require.Equal(t, "fileName", defaultCodeProps.Networks[0].Name)
		require.Equal(t, "fileURL1", defaultCodeProps.Networks[0].HttpURLs[0])
	})
	t.Run("networks override with file", func(t *testing.T) {
		logging.Init()
		codeProps := &presets.MultiNetworkOpts{
			Networks: presets.Networks{
				{
					Name:     "codeName",
					Type:     "external",
					HttpURLs: []string{"codeHttp"},
					WsURLs:   []string{"codeWs"},
				},
			},
		}
		defaultCodeProps := &presets.MultiNetworkOpts{
			Networks: presets.Networks{
				{
					Name:     "default",
					Type:     "external",
					HttpURLs: []string{"defaultHttp"},
					WsURLs:   []string{"defaultWS"},
				},
			},
		}
		path := "networks.yaml"
		// nolint
		os.Setenv("NETWORKS_CONFIG_FILE", path)
		config.MustEnvCodeOverrideStruct("", defaultCodeProps, codeProps)
		require.Equal(t, "fileName", defaultCodeProps.Networks[0].Name)
		require.Equal(t, "fileURL1", defaultCodeProps.Networks[0].HttpURLs[0])
	})
}
