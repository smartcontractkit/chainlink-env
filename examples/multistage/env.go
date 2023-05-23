package main

import (
	"time"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
)

func main() {
	chainlinkChart, err := chainlink.New(0, nil)
	if err != nil {
		panic(err)
	}
	e := environment.New(&environment.Config{TTL: 20 * time.Minute})
	err = e.
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart).
		Run()
	if err != nil {
		panic(err)
	}
	// deploy another part
	e.Cfg.KeepConnection = true
	chainlinkChart2, err := chainlink.New(1, nil)
	if err != nil {
		panic(err)
	}
	err = e.
		AddHelm(chainlinkChart2).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		Run()
	defer func() {
		errr := e.Shutdown()
		panic(errr)
	}()
	if err != nil {
		panic(err)
	}
}
