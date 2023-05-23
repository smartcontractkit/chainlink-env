package main

import (
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

func main() {
	chainlinkChart, err := chainlink.New(0, nil)
	if err != nil {
		panic(err)
	}
	err = environment.New(&environment.Config{
		Labels:            []string{"type=construction-in-progress"},
		NamespacePrefix:   "new-environment",
		KeepConnection:    true,
		RemoveOnInterrupt: true,
	}).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlinkChart).
		Run()
	if err != nil {
		panic(err)
	}
}
