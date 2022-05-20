package main

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/chains/ethereum"
	"github.com/smartcontractkit/chainlink-env/chains/solana"
	"github.com/smartcontractkit/chainlink-env/testrunner"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/chainlink"
	"github.com/smartcontractkit/chainlink-env/client"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	s := client.NewSnapshotTest()
	nsCount := 1
	ns := fmt.Sprintf("zclcdk-test-%d", nsCount)
	s.AddCase(
		"Create Solana OCR cluster",
		ns,
		"chainlink-ocr-solana-cluster.k8s.yaml",
		chainlink.NewChart,
		&chainlink.Props{
			Namespace: ns,
			ChainProps: []interface{}{
				&solana.Props{},
			},
			AppVersions: []chainlink.VersionProps{
				{
					Image:     "public.ecr.aws/z0b1w9r9/chainlink",
					Tag:       "develop.d2b337227bc5d4c4b6ee4f605017f205e01e5d16",
					Instances: 5,
				},
			},
		})
	nsCount++
	s.AddCase(
		"Create generic multi-versions cluster",
		ns,
		"chainlink-ocr-cluster-multiver.k8s.yaml",
		chainlink.NewChart,
		&chainlink.Props{
			Namespace: ns,
			ChainProps: []interface{}{
				&ethereum.Props{},
			},
			AppVersions: []chainlink.VersionProps{
				{
					Image:     "public.ecr.aws/z0b1w9r9/chainlink",
					Tag:       "develop.d2b337227bc5d4c4b6ee4f605017f205e01e5d16",
					Instances: 2,
				},
				{
					Image:     "public.ecr.aws/chainlink/chainlink",
					Tag:       "develop.567932f0bc793a5ca1804a6dfa40863793748769",
					Instances: 1,
				},
			},
		})
	nsCount++
	s.AddCase(
		"Create generic OCR 5 node cluster",
		ns,
		"chainlink-ocr-cluster.k8s.yaml",
		chainlink.NewChart,
		&chainlink.Props{
			Namespace: ns,
			ChainProps: []interface{}{
				&ethereum.Props{},
			},
			AppVersions: []chainlink.VersionProps{
				{
					Image:     "public.ecr.aws/z0b1w9r9/chainlink",
					Tag:       "develop.d2b337227bc5d4c4b6ee4f605017f205e01e5d16",
					Instances: 5,
				},
			},
		})
	nsCount++
	s.AddCase(
		"Create generic OCR 5 node cluster with persistence",
		ns,
		"chainlink-ocr-cluster-persistence.k8s.yaml",
		chainlink.NewChart,
		&chainlink.Props{
			Namespace: ns,
			ChainProps: []interface{}{
				&ethereum.Props{},
			},
			AppVersions: []chainlink.VersionProps{
				{
					Image:       "public.ecr.aws/z0b1w9r9/chainlink",
					Tag:         "develop.d2b337227bc5d4c4b6ee4f605017f205e01e5d16",
					Instances:   5,
					Persistence: chainlink.PersistenceProps{Capacity: "2Gi"},
				},
			},
		})
	nsCount++
	s.AddCase(
		"Create generic OCR 5 node cluster with remote test runner",
		ns,
		"chainlink-ocr-cluster-remote-testrunner.k8s.yaml",
		chainlink.NewChart,
		&chainlink.Props{
			Namespace: ns,
			TestRunnerProps: testrunner.Props{
				TestTag:            "@ocr-soak",
				ConfigFileContents: "",
				SlackAPIKey:        "",
				SlackChannel:       "",
				SlackUserID:        "",
				TestBinarySize:     100,
				AccessPort:         8080,
			},
			ChainProps: []interface{}{
				&ethereum.Props{},
			},
			AppVersions: []chainlink.VersionProps{
				{
					Image:       "public.ecr.aws/z0b1w9r9/chainlink",
					Tag:         "develop.d2b337227bc5d4c4b6ee4f605017f205e01e5d16",
					Instances:   5,
					Persistence: chainlink.PersistenceProps{Capacity: "2Gi"},
				},
			},
		})
	if err := s.Run(); err != nil {
		panic(err)
	}
}
