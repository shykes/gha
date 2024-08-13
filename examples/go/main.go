package main

import (
	"context"
	"dagger/dagger-2-gha/examples/go/internal/dagger"
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
func (m *Examples) Dagger2Gha() *dagger.Directory {
	return dag.
		Dagger2Gha().
		OnPush("hello --name=main", dagger.Dagger2GhaOnPushOpts{
			Branches: []string{"main"},
			Module:   "github.com/shykes/hello",
		}).
		OnPullRequest("hello --name='pull request'", dagger.Dagger2GhaOnPullRequestOpts{
			Module: "github.com/shykes/hello",
		}).
		Config(dagger.Dagger2GhaConfigOpts{
			Prefix: "example-",
		})
}

// Generate a configuration with a "dispatch" workflow,
// that can be triggered manually, independently of any platform event.
func (m *Examples) Dagger2Gha_OnDispatch() *dagger.Directory {
	return dag.
		Dagger2Gha().
		OnDispatch("deploy-docs").
		Config(dagger.Dagger2GhaConfigOpts{
			Prefix: "example-",
		})
}
