package common

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/presets"
	"github.com/stretchr/testify/require"
)

const (
	TestEnvType = "chainlink-env-test"
)

var (
	testSelector = fmt.Sprintf("envType=%s", TestEnvType)
)

func GetTestEnvConfig(t *testing.T) *environment.Config {
	return &environment.Config{
		NamespacePrefix: TestEnvType,
		Labels:          []string{testSelector},
		Test:            t,
	}
}

func TestMultiStageMultiManifestConnection(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)

	ethChart := ethereum.New(nil)
	ethNetworkName := ethChart.GetProps().(*ethereum.Props).NetworkName

	// we adding the same chart with different index and executing multi-stage deployment
	// connections should be renewed
	e := environment.New(testEnvConfig)
	err := e.AddHelm(ethChart).
		AddHelm(chainlink.New(0, nil)).
		Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
	require.Len(t, e.URLs[chainlink.NodesLocalURLsKey], 1)
	require.Len(t, e.URLs[chainlink.NodesInternalURLsKey], 1)
	require.Len(t, e.URLs[chainlink.DBsLocalURLsKey], 1)
	require.Len(t, e.URLs, 7)

	err = e.AddHelm(chainlink.New(1, nil)).
		Run()
	require.NoError(t, err)
	require.Len(t, e.URLs[chainlink.NodesLocalURLsKey], 2)
	require.Len(t, e.URLs[chainlink.NodesInternalURLsKey], 2)
	require.Len(t, e.URLs[chainlink.DBsLocalURLsKey], 2)
	require.Len(t, e.URLs, 7)

	urls := make([]string, 0)
	urls = append(urls, e.URLs[chainlink.NodesLocalURLsKey]...)
	urls = append(urls, e.URLs[ethNetworkName+"_http"]...)

	r := resty.New()
	for _, u := range urls {
		log.Info().Str("URL", u).Send()
		res, err := r.R().Get(u)
		require.NoError(t, err)
		require.Equal(t, "200 OK", res.Status())
	}
}

func TestConnectWithoutManifest(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": 1,
		}))
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})

	testEnvConfig.NoManifestUpdate = true
	testEnvConfig.Namespace = e.Cfg.Namespace
	err = environment.New(testEnvConfig).
		Run()
	require.NoError(t, err)
	url, err := e.Fwd.FindPort("chainlink-0:0", "node", "access").As(client.LocalConnection, client.HTTP)
	require.NoError(t, err)
	urlGeth, err := e.Fwd.FindPort("geth:0", "geth-network", "http-rpc").As(client.LocalConnection, client.HTTP)
	require.NoError(t, err)
	r := resty.New()
	res, err := r.R().Get(url)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
	res, err = r.R().Get(urlGeth)
	require.NoError(t, err)
	require.Equal(t, "200 OK", res.Status())
}

func Test5NodesSoakEnvironmentWithPVCs(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMSoak(testEnvConfig)
	err := e.Run()
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
	require.NoError(t, err)
}

func TestWithSingleNodeEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMOneNode(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
}

func TestMinResources5NodesEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMMinimalLocal(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
}

func TestMinResources5NodesEnvWithBlockscout(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMMinimalLocalBS(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
}

func Test5NodesPlus2MiningGethsReorgEnv(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := presets.EVMReorg(testEnvConfig)
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
}

func TestMultipleInstancesOfTheSameType(t *testing.T) {
	t.Parallel()
	testEnvConfig := GetTestEnvConfig(t)
	e := environment.New(testEnvConfig).
		AddHelm(ethereum.New(nil)).
		AddHelm(chainlink.New(0, nil)).
		AddHelm(chainlink.New(1, nil))
	err := e.Run()
	require.NoError(t, err)
	if e.WillUseRemoteRunner() {
		return
	}
	t.Cleanup(func() {
		// nolint
		e.Shutdown()
	})
}

// func TestWithChaos(t *require.TestingT) {
// 	t.Parallel()

// 	testCases := map[string]struct {
// 		networkChart environment.ConnectedChart
// 		clChart      environment.ConnectedChart
// 		chaosFunc    chaos.ManifestFunc
// 		chaosProps   *chaos.Props
// 	}{
// 		// see ocr_chaos.test.go for comments
// 		"pod-chaos-fail-minority-nodes": {
// 			ethereum.New(nil),
// 			chainlink.New(0, nil),
// 			chaos.NewFailPods,
// 			&chaos.Props{
// 				LabelsSelector: &map[string]*string{"chaosGroupMinority": a.Str("1")},
// 				DurationStr:    "1m",
// 			},
// 		},
// 	}
// }
