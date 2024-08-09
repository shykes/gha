// Generate Github Actions configurations from Dagger pipelines
package main

import (
	"dagger/dagger-2-gha/internal/dagger"
	"fmt"
	"strconv"
	"strings"
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

// Add a Dagger pipeline to be called on pull request to the given branch
func (m *Dagger2Gha) OnPullRequest(
	// +optional
	// +default="main"
	branch string,
	// +optional
	// +default="."
	module string,
	function string,
	// +optional
	args []string,
	// +optional
	noCheckout bool,
) *Dagger2Gha {
	m.Pipelines = append(m.Pipelines, Pipeline{
		OnPullRequestBranch: branch,
		Module:              module,
		Function:            function,
		Args:                args,
		Checkout:            !noCheckout,
		DaggerVersion:       m.DaggerVersion,
	})
	return m
}

func (m *Dagger2Gha) Config() *dagger.Directory {
	dir := dag.Directory()
	for i, pipeline := range m.Pipelines {
		dir = dir.WithFile(
			fmt.Sprintf(".github/workflows/%d.yml", i+1),
			pipeline.Config().File(),
		)
	}
	return dir
}

type Pipeline struct {
	OnPullRequestBranch string
	Module              string
	Function            string
	Args                []string
	Checkout            bool
	DaggerVersion       string
}

func (p *Pipeline) Name() string {
	return strings.Join(append([]string{p.Function}, p.Args...), " ")
}

func (p *Pipeline) Config() Workflow {
	var steps []Step
	if p.Checkout {
		steps = append(steps, Step{
			Name: "Checkout",
			Uses: "actions/checkout@v4",
		})
	}
	steps = append(steps, Step{
		Name: "Call Dagger",
		Uses: "dagger/dagger-for-github@v6",
		With: map[string]string{
			"version": "latest",
			"module":  p.Module,
			"args":    shellCommand(append([]string{p.Function}, p.Args...)),
		},
	})
	return Workflow{
		Name: p.Name(),
		On: WorkflowTriggers{
			PullRequest: &PullRequestEvent{
				Branches: []string{p.OnPullRequestBranch},
			},
		},
		Jobs: map[string]Job{
			"main": Job{
				Steps:  steps,
				RunsOn: "ubuntu-latest",
			},
		},
	}
}

func shellCommand(args []string) string {
	var escapedArgs []string
	for _, arg := range args {
		escapedArgs = append(escapedArgs, strconv.Quote(arg))
	}
	return strings.Join(escapedArgs, " ")
}
