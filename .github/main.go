package main

import (
	"github.com/shykes/gha/.github/internal/dagger"
)

type Github struct{}

// Generate our CI config
// Export to .github at the repository root
// Example: 'dagger call -m .github -o .github'
func (m *Github) Generate() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline(
			"Demo pipeline 1",
			"git --url=https://github.com/$GITHUB_REPOSITORY branch --name=$GITHUB_REF tree glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				Module: "github.com/shykes/core",
			}).
		WithPipeline(
			"Demo pipeline 2",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				SparseCheckout: []string{"misc", "scripts"},
				Module:         "github.com/shykes/core",
			},
		).
		WithPipeline(
			"Demo pipeline 3",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				Module:   "github.com/shykes/core",
				Dispatch: true,
			}).
		// Trigger 'Demo pipeline 1' on:
		//  - push to main branch
		//  - push to any tag
		OnPush(
			[]string{"Demo pipeline 1"},
			dagger.GhaOnPushOpts{
				Branches: []string{"main"},
				Tags:     []string{"*"},
			}).
		// Trigger 'Demo pipeline 1' and 'Demo pipeline 2' on pull request
		OnPullRequest([]string{
			"Demo pipeline 1",
			"Demo pipeline 2",
		}).
		Config().
		Directory(".github")
}
