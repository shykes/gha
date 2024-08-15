package main

import (
	"github.com/shykes/gha/examples/go/internal/dagger"
)

type Examples struct{}

// Access Github secrets
func (m *Examples) Gha_Secrets() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline(
			"deploy docs",
			"deploy-docs --source=. --password env:$DOCS_SERVER_PASSWORD",
			dagger.GhaWithPipelineOpts{
				Dispatch: true,
				Secrets:  []string{"DOCS_SERVER_PASSWORD"},
			}).
		Config()
}

// Access github context information magically injected as env variables
func (m *Examples) Gha_GithubContext() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline("lint all branches", "lint --source=${GITHUB_REPOSITORY_URL}#${GITHUB_REF}").
		OnPush([]string{"lint all branches"}).
		Config()
}

// Compose a pipeline from an external module, instead of the one embedded in the repo.
func (m *Examples) Gha_CustomModule() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline(
			"say hello",
			"hello --name=$GITHUB_REPOSITORY_OWNER",
			dagger.GhaWithPipelineOpts{
				Module: "github.com/shykes/hello",
			}).
		Config()
}

// Call the repo's 'build()' dagger function on push to main
func (m *Examples) GhaOnPush() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline(
			"build and publish app container from main",
			"publish --source=. --registry-user=$REGISTRY_USER --registry-password=$REGISTRY_PASSWORD",
			dagger.GhaWithPipelineOpts{
				Secrets: []string{
					"REGISTRY_USER", "REGISTRY_PASSWORD",
				},
			}).
		OnPush([]string{"build and publish app container from main"},
			dagger.GhaOnPushOpts{
				Branches: []string{"main"},
			}).
		Config()
}

// Call integration tests on pull requests
func (m *Examples) GhaOnPullRequest() *dagger.Directory {
	return dag.
		Gha().
		WithPipeline("test pull requests", "test --all --source=.").
		OnPullRequest([]string{"test pull requests"}).
		Config()
}
