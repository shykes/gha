package main

import (
	"dagger/dagger-2-gha/examples/go/internal/dagger"
)

type Go struct{}

// Generate a simple workflow configuration
func (m *Go) Config() *dagger.Directory {
	return dag.
		Dagger2Gha().
		OnPullRequest("hello", dagger.Dagger2GhaOnPullRequestOpts{
			Module:     "github.com/shykes/hello",
			NoCheckout: true,
		}).
		OnPullRequest("hello", dagger.Dagger2GhaOnPullRequestOpts{
			Args: []string{
				"--greeting", "bonjour",
				"--name", "monde",
			},
			Module:     "github.com/shykes/hello",
			NoCheckout: true,
		}).
		OnPullRequest("container", dagger.Dagger2GhaOnPullRequestOpts{
			Module: "github.com/shykes/daggerverse/wolfi",
			Args: []string{
				"with-mounted-directory", "--source=.", "--path=/src",
				"with-workdir", "--path=/src",
				"with-exec", "--args=ls,-l",
				"stdout",
			},
		}).
		Config()
}
