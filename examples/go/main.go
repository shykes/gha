package main

import (
	"context"

	"github.com/shykes/gha/examples/go/internal/dagger"
)

type Examples struct{}

func (m *Examples) Hello(ctx context.Context, lang string) (string, error) {
	if lang == "fr" {
		return dag.Hello().Hello(ctx, dagger.HelloHelloOpts{
			Greeting: "bonjour",
			Name:     "monde",
		})
	}
	return dag.Hello().Hello(ctx)
}

func (m *Examples) ContainerStuff(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.
		Wolfi().
		Container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"ls", "-l"}).
		Stdout(ctx)
}

// Generate a simple configuration triggered by git push on main
func (m *Examples) GhaOnPush() *dagger.Directory {
	return dag.
		Gha().
		OnPush("hello --name=main", dagger.GhaOnPushOpts{
			Branches: []string{"main"},
			Module:   "github.com/shykes/hello",
		}).
		Config()
}

// Generate a simple configuration triggered by a pull request
func (m *Examples) GhaOnPullRequest() *dagger.Directory {
	return dag.
		Gha().
		OnPullRequest("hello --name='pull request'", dagger.GhaOnPullRequestOpts{
			Module: "github.com/shykes/hello",
		}).
		Config()
}

// Access Github secrets
func (m *Examples) Gha_Secrets() *dagger.Directory {
	return dag.
		Gha().
		OnDispatch("deploy-docs", dagger.GhaOnDispatchOpts{
			Secrets: []string{"DOCS_SERVER_PASSWORD"},
		}).
		Config()
}

// Access github context information magically injected as env variables
func (m *Examples) Gha_GithubContext() *dagger.Directory {
	return dag.
		Gha().
		OnDispatch(
			"git --url=https://github.com/$GITHUB_REPOSITORY branch --name=$GITHUB_REF tree glob --pattern=*",
			dagger.GhaOnDispatchOpts{
				Module: "github.com/shykes/core",
			}).
		Config()
}
