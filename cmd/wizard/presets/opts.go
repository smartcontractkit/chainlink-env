package presets

type ExternalNetworkOpts struct {
	HttpURL string `envconfig:"http_url"`
	WsURL   string `envconfig:"ws_url"`
	ChainID string `envconfig:"chain_id"`
}
