package main

import (
	"github.com/shykes/gha/.github/internal/dagger"
)

type Github struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Github) Generate() *dagger.Directory {
	return dag.
		Gha().
		OnPush(
			"Demo pipeline 1",
			"git --url=https://github.com/$GITHUB_REPOSITORY branch --name=$GITHUB_REF tree glob --pattern=*",
			dagger.GhaOnPushOpts{
				Module: "github.com/shykes/core",
			}).
		OnPush(
			"Demo pipeline 2",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaOnPushOpts{
				SparseCheckout: []string{"misc", "scripts"},
				Module:         "github.com/shykes/core",
			},
		).
		OnDispatch(
			"Demo pipeline 3",
			"directory with-directory --path=. --source=. glob --pattern=*",
			dagger.GhaOnDispatchOpts{
				Module: "github.com/shykes/core",
			},
		).
		Config().
		Directory(".github")
}
