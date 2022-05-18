package client

import (
	"fmt"
	"os"
	"path/filepath"

	prompt "github.com/c-bata/go-prompt"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/rs/zerolog/log"
)

func dialogue(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "y", Description: "Write new chart and change snapshot"},
		{Text: "n", Description: "Abort changes"},
		{Text: "d", Description: "Deploy newly generated manifest"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

type Snapshots struct {
	Client    *K8sClient
	Snapshots []string
	TestCases []func() TestData
}

func NewSnapshotTest() *Snapshots {
	return &Snapshots{
		Client:    NewK8sClient(),
		Snapshots: make([]string, 0),
		TestCases: make([]func() TestData, 0),
	}
}

func (s *Snapshots) loadSnapshot(path string) string {
	p := filepath.Join(DistRoot, path)
	snap, err := os.ReadFile(p)
	if err != nil {
		log.Warn().Str("Path", p).Msg("No snapshot found")
	}
	return string(snap)
}

type TestData struct {
	TestName         string
	Namespace        string
	Props            interface{}
	ChartProps       cdk8s.ChartProps
	Check            ManifestOutput
	NewChart         string
	ManifestFilename string
	OldChart         string
}

func (s *Snapshots) AddCase(
	name string,
	namespace string,
	manifestFilename string,
	generator func(props interface{}) (cdk8s.App, ManifestOutput),
	props interface{},
) {
	s.TestCases = append(s.TestCases, func() TestData {
		app, checkData := generator(props)
		chart := app.SynthYaml().(string)
		return TestData{
			TestName:         name,
			Props:            props,
			Namespace:        namespace,
			Check:            checkData,
			NewChart:         chart,
			ManifestFilename: manifestFilename,
			OldChart:         s.loadSnapshot(manifestFilename),
		}
	})
}

func (s *Snapshots) DiffMyers(testcase TestData) error {
	edits := myers.ComputeEdits(span.URIFromPath(testcase.ManifestFilename), testcase.OldChart, testcase.NewChart)
	fmt.Print(gotextdiff.ToUnified(testcase.ManifestFilename, testcase.ManifestFilename, testcase.OldChart, edits))
	return nil
}

func (s *Snapshots) changesDialogue(testcase TestData) error {
	if err := s.DiffMyers(testcase); err != nil {
		return err
	}
	log.Warn().Msg("Snapshot has changed, apply changes? (y/n/d)")
	choice := prompt.Input("> ", dialogue)
	switch choice {
	case "y":
		p := filepath.Join(DistRoot, testcase.ManifestFilename)
		log.Info().Str("Chart", p).Msg("Chart saved")
		if err := os.WriteFile(p, []byte(testcase.NewChart), os.ModePerm); err != nil {
			return err
		}
	case "n":
		log.Warn().Msg("Aborting process")
		os.Exit(0)
	case "d":
		log.Info().Str("Manifest", testcase.ManifestFilename).Str("Namespace", testcase.Namespace).Msg("Deploying")
		if err := s.Client.Create(testcase.NewChart); err != nil {
			return err
		}
		if err := s.Client.CheckReady(testcase.Check); err != nil {
			return err
		}
		// nolint
		defer s.Client.RemoveNamespace(testcase.Namespace)
		log.Warn().Msg("Check deployment in Lens, all is fine? (y/n)")
		c := prompt.Input("> ", dialogue)
		updatedChartPath := filepath.Join(DistRoot, testcase.ManifestFilename)
		switch c {
		case "y":
			log.Info().Str("Chart", testcase.ManifestFilename).Msg("Chart updated")
			if err := os.WriteFile(updatedChartPath, []byte(testcase.NewChart), os.ModePerm); err != nil {
				return err
			}
		case "n":
			log.Warn().Str("Chart", updatedChartPath).Msg("New Chart aborted")
			return nil
		}
	}
	return nil
}

func (s *Snapshots) Run() error {
	for _, generator := range s.TestCases {
		testcase := generator()
		log.Info().
			Str("Deployment", testcase.ManifestFilename).
			Interface("Props", testcase.Props).
			Msg("Checking deployment")
		if testcase.NewChart != testcase.OldChart {
			if err := s.changesDialogue(testcase); err != nil {
				return err
			}
		}
	}
	return nil
}
