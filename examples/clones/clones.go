package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

func main() {
	// Multiple environments of the same type/chart
	chainlinkChart1, err := chainlink.New(0, map[string]interface{}{
		"chainlink": map[string]interface{}{
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu": "344m",
				},
				"limits": map[string]interface{}{
					"cpu": "344m",
				},
			},
		},
		"db": map[string]interface{}{
			"stateful": "true",
			"capacity": "1Gi",
		},
	})
	if err != nil {
		panic(err)
	}
	chainlinkChart2, err := chainlink.New(1,
		map[string]interface{}{
			"chainlink": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu": "577m",
					},
					"limits": map[string]interface{}{
						"cpu": "577m",
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}
	err = environment.New(&environment.Config{
		Labels:            []string{fmt.Sprintf("envType=%s", pkg.EnvTypeEVM5)},
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart1).
		AddHelm(chainlinkChart2).
		Run()
	if err != nil {
		panic(err)
	}
}
