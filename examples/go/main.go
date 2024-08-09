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
		WithPipeline("hello-english", "hello --lang=en").
		WithPipeline("hello-french", "hello --lang=fr").
		WithPipeline("container-stuff", "container-stuff --source=.").
		Config()
}
