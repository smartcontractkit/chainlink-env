package main

import (
	"time"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/examples/deployment_part"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
)

func main() {
	e := environment.New(&environment.Config{
		NamespacePrefix:   "adding-new-deployment-part",
		TTL:               3 * time.Hour,
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).
		AddHelm(deployment_part.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 5,
			"env": map[string]interface{}{
				"CL_DEV": "false",
				"CL_CONFIG": `
OCR.Enabled = false

[OCR2]
Enabled = true

P2P.V1.ListenPort = 0
[P2P.V2]
Enabled = true
ListenAddresses = ['0.0.0.0:6690']
DeltaDial = '5s'
`,
			},
		}))
	if err := e.Run(); err != nil {
		panic(err)
	}
}
