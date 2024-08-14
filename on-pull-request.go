package main

import "github.com/shykes/gha/internal/dagger"

// PULL REQUEST TRIGGER

// Add a trigger to execute a Dagger pipeline on a pull request
func (m *Gha) OnPullRequest(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// Dagger module to load
	// +optional
	// +default="."
	module string,
	// Github secrets to inject into the pipeline environment.
	// For each secret, an env variable with the same name is created.
	// Example: ["PROD_DEPLOY_TOKEN", "PRIVATE_SSH_KEY"]
	// +optional
	secrets []string,
	// +optional
	// Run only for pull requests that target specific branches
	branches []string,
	// Run only for certain types of pull request events
	// See https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#pull_request
	// +optional
	types []string,
	// Dispatch jobs to the given runner
	// +optional
	runner string,
) *Gha {
	if err := validateSecretNames(secrets); err != nil {
		panic(err) // FIXME
	}
	m.PullRequestTriggers = append(m.PullRequestTriggers, PullRequestTrigger{
		Event: PullRequestEvent{
			Branches: branches,
			Types:    types,
		},
		Pipeline: m.pipeline(command, module, runner, secrets),
	})
	return m
}

type PullRequestTrigger struct {
	Event    PullRequestEvent
	Pipeline Pipeline
}

func (t PullRequestTrigger) asWorkflow() Workflow {
	var workflow = t.Pipeline.asWorkflow()
	workflow.On = WorkflowTriggers{PullRequest: &(t.Event)}
	return workflow
}

func (t PullRequestTrigger) Config(filename string, asJson bool) *dagger.Directory {
	return t.asWorkflow().Config(filename, asJson)
}
