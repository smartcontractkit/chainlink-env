package pkg

import a "github.com/smartcontractkit/chainlink-env/pkg/alias"

type ResourcesMode int

const (
	MinimalLocalResourcesMode ResourcesMode = iota
	SoakResourcesMode
)

func PGIsReadyCheck() *[]*string {
	return &[]*string{
		a.Str("pg_isready"),
		a.Str("-U"),
		a.Str("postgres"),
	}
}
