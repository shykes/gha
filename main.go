// Generate Github Actions configurations from Dagger pipelines
package main

import (
	"dagger/dagger-2-gha/internal/dagger"
)

func New(
	// Public Dagger Cloud token, for open-source projects. DO NOT PASS YOUR PRIVATE DAGGER CLOUD TOKEN!
	// This is for a special "public" token which can safely be shared publicly.
	// To get one, contact support@dagger.io
	// +optional
	publicToken string,
	// Dagger version to run in the Github Actions pipelines
	// +optional
	// +default="latest"
	daggerVersion string,
) *Dagger2Gha {
	return &Dagger2Gha{
		PublicToken:   publicToken,
		DaggerVersion: daggerVersion,
	}
}

type Dagger2Gha struct {
	// +private
	Pipelines []Pipeline
	// +private
	PublicToken string
	// +private
	DaggerVersion string
}

func (m *Dagger2Gha) WithPipeline(
	// Name of the Github Action
	name string,
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// +optional
	// +default="."
	module string,
) *Dagger2Gha {
	m.Pipelines = append(m.Pipelines, Pipeline{
		DaggerVersion: m.DaggerVersion,
		PublicToken:   m.PublicToken,
		Name:          name,
		Command:       command,
		Module:        module,
	})
	return m
}

// Generate a github config directory, usable as an overlay on the repository root
func (m *Dagger2Gha) Config() *dagger.Directory {
	dir := dag.Directory()
	for _, p := range m.Pipelines {
		dir = dir.WithDirectory(".", p.GithubAction().Config())
	}
	return dir
}

type Pipeline struct {
	// +private
	DaggerVersion string
	// +private
	PublicToken string
	Name        string
	Module      string
	Command     string
}

func (p *Pipeline) GithubAction() Action {
	var env = make(map[string]string)
	if p.PublicToken != "" {
		env["DAGGER_CLOUD_TOKEN"] = p.PublicToken
	}
	action := Action{
		Name: p.Name,
		Runs: Runs{
			Using: "composite",
			Steps: []CompositeActionStep{
				CompositeActionStep{
					Name: "Checkout",
					Uses: "actions/checkout@v4",
				},
				CompositeActionStep{
					Name: "Dagger",
					Uses: "dagger/dagger-for-github@v6",
					With: map[string]string{
						"version": p.DaggerVersion,
						"command": p.Command,
						"module":  p.Module,
					},
					Env: env,
				},
			},
		},
	}

	return action
}
