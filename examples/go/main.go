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

// Generate a simple configuration triggered by git push on main, and pull requests
// 1. Prints "hello, main!" on push to main
// 2. Prints "hello, pull request!" on new pull request
func (m *Examples) Gha() *dagger.Directory {
	return dag.
		Gha().
		OnPush("hello --name=main", dagger.GhaOnPushOpts{
			Branches: []string{"main"},
			Module:   "github.com/shykes/hello",
		}).
		OnPullRequest("hello --name='pull request'", dagger.GhaOnPullRequestOpts{
			Module: "github.com/shykes/hello",
		}).
		Config(dagger.GhaConfigOpts{
			Prefix: "example-",
		})
}

// Generate a configuration with a "dispatch" workflow,
// that can be triggered manually, independently of any platform event.
func (m *Examples) Gha_OnDispatch() *dagger.Directory {
	return dag.
		Gha().
		OnDispatch("deploy-docs", dagger.GhaOnDispatchOpts{
			Secrets: []string{"DOCS_SERVER_PASSWORD"},
		}).
		Config(dagger.GhaConfigOpts{
			Prefix: "example-",
		})
}
