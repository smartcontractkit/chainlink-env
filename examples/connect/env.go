package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"os"
	"time"
)

func myEnv() *environment.Environment {
	return environment.New(&environment.Config{
		Labels: []string{fmt.Sprintf("envType=testenv")},
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil))
}

func main() {
	// example of quick usage to debug env, removed on SIGINT
	//os.Setenv("CHAINLINK_IMAGE", "ddd")
	//os.Setenv("CHAINLINK_VERSION", "aaa")
	e := myEnv()
	err := e.Run()
	if err != nil {
		panic(err)
	}
	os.Setenv("NAMESPACE_NAME", e.Cfg.Namespace)
	err = e.Run()
	// nolint
	defer e.Shutdown()
	if err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Minute)
}
