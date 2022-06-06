package presets

import (
	"github.com/smartcontractkit/chainlink-env/config"
	"strings"
)

type MultiNetworkOpts struct {
	Networks Networks `envconfig:"NETWORKS_CONFIG_FILE"`
}

type Networks []Network

func (m *Networks) Decode(path string) error {
	if strings.HasPrefix(path, "base64:") {
		path = strings.Replace(path, "base64:", "", -1)
		return config.UnmarshalYAMLBase64(path, m)
	}
	return config.UnmarshalYAMLFile(path, m)
}

type Network struct {
	Name     string   `envconfig:"name" yaml:"name"`
	Type     string   `envconfig:"type" yaml:"type"`
	HttpURLs []string `envconfig:"http_urls" yaml:"http_urls"`
	WsURLs   []string `envconfig:"ws_urls" yaml:"ws_urls"`
	ChainID  string   `envconfig:"chain_id" yaml:"chain_id"`
}
