package main

import (
	"context"
	"dagger/dagger-2-gha/examples/go/internal/dagger"
)

type Go struct{}

func (m *Go) Hello(ctx context.Context, lang string) (string, error) {
	if lang == "fr" {
		return dag.Hello().Hello(ctx, dagger.HelloHelloOpts{
			Greeting: "bonjour",
			Name:     "monde",
		})
	}
	return dag.Hello().Hello(ctx)
}

func (m *Go) ContainerStuff(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.
		Wolfi().
		Container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"ls", "-l"}).
		Stdout(ctx)
}

// Generate a simple workflow configuration
func (m *Go) Main() *dagger.Directory {
	return dag.
		Dagger2Gha().
		OnPush("hello --name=main", dagger.Dagger2GhaOnPushOpts{
			Branches: []string{"main"},
			Module:   "github.com/shykes/hello",
		}).
		OnPullRequest("hello --name='pull request'", dagger.Dagger2GhaOnPullRequestOpts{
			Module: "github.com/shykes/hello",
		}).
		Config()
}
