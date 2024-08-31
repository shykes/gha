package main

import (
	"github.com/shykes/gha/.github/internal/dagger"
)

type Github struct{}

// Generate our CI config
// Export to .github at the repository root
// Example: 'dagger call -m .github -o .github'
func (m *Github) Generate(
	// +optional
	// +defaultPath="/"
	// +ignore=["!.github"]
	repository *dagger.Directory,
) *dagger.Directory {
	return dag.
		Gha(dagger.GhaOpts{
			DaggerVersion: "latest",
			Repository:    repository,
		}).
		WithPipeline(
			"Deploy docs",
			"deploy-docs --token $NETLIFY_TOKEN",
			dagger.GhaWithPipelineOpts{
				Secrets:     []string{"NETLIFY_TOKEN"},
				OnPushTags:  []string{"deploy-docs"},
				Permissions: []dagger.GhaPermission{dagger.ReadContents},
			},
		).
		WithPipeline(
			"Demo pipeline 1",
			"git --url=https://github.com/$GITHUB_REPOSITORY branch --name=$GITHUB_REF tree glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				Module:         "github.com/shykes/core",
				OnPullRequest:  true,
				OnPushBranches: []string{"main"},
				OnPushTags:     []string{"*"},
			}).
		WithPipeline(
			"Demo pipeline 2",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				SparseCheckout: []string{"misc", "scripts"},
				Module:         "github.com/shykes/core",
				OnPullRequest:  true,
				OnPushBranches: []string{"main"},
				OnPushTags:     []string{"*"},
			},
		).
		WithPipeline(
			"Demo pipeline 3",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				Module:   "github.com/shykes/core",
				Dispatch: true,
			}).
		WithPipeline(
			"Schedule pipeline",
			"directory with-directory --path=. --directory=. glob --pattern=*",
			dagger.GhaWithPipelineOpts{
				Module:     "github.com/shykes/core",
				OnSchedule: []string{"*/20 * * * *"}, // run every 20 minutes.
			},
		).
		Config()
}
