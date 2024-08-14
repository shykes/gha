package main

import (
	"errors"
	"regexp"

	"github.com/shykes/gha/internal/dagger"
)

// PUSH TRIGGER

// Add a trigger to execute a Dagger pipeline on a git push
func (m *Gha) OnPush(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// The Dagger module to load
	// +optional
	// +default="."
	module string,
	// Github secrets to inject into the pipeline environment.
	// For each secret, an env variable with the same name is created.
	// Example: ["PROD_DEPLOY_TOKEN", "PRIVATE_SSH_KEY"]
	// +optional
	secrets []string,
	// Run only on push to specific branches
	// +optional
	branches []string,
	// Run only on push to specific branches
	// +optional
	tags []string,
	// Run only on push to specific paths
	// +optional
	paths []string,
	// Dispatch jobs to the given runner
	// +optional
	runner string,
) *Gha {
	if err := validateSecretNames(secrets); err != nil {
		panic(err) // FIXME
	}
	m.PushTriggers = append(m.PushTriggers, PushTrigger{
		Event: PushEvent{
			Branches: branches,
			Tags:     tags,
			Paths:    paths,
		},
		Pipeline: m.pipeline(command, module, runner, secrets),
	})
	return m
}

type PushTrigger struct {
	Event    PushEvent
	Pipeline Pipeline
}

func (t PushTrigger) asWorkflow() Workflow {
	var workflow = t.Pipeline.asWorkflow()
	workflow.On = WorkflowTriggers{Push: &(t.Event)}
	return workflow
}

func (t PushTrigger) Config(filename string, asJson bool) *dagger.Directory {
	return t.asWorkflow().Config(filename, asJson)
}

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

// DISPATCH TRIGGER

// Add a trigger to execute a Dagger pipeline on a workflow dispatch event
func (m *Gha) OnDispatch(
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
	// Dispatch jobs to the given runner
	// +optional
	runner string,
) *Gha {
	if err := validateSecretNames(secrets); err != nil {
		panic(err) // FIXME
	}
	m.DispatchTriggers = append(m.DispatchTriggers, DispatchTrigger{
		Pipeline: m.pipeline(command, module, runner, secrets),
		Event: WorkflowDispatchEvent{
			Inputs: nil, // FIXME: add inputs, could be pretty dope
		},
	})
	return m
}

type DispatchTrigger struct {
	// When this happens...
	Event WorkflowDispatchEvent
	// ...run this
	Pipeline Pipeline
}

func (t DispatchTrigger) asWorkflow() Workflow {
	var workflow = t.Pipeline.asWorkflow()
	workflow.On = WorkflowTriggers{WorkflowDispatch: &(t.Event)}
	return workflow
}

func (t DispatchTrigger) Config(filename string, asJson bool) *dagger.Directory {
	return t.asWorkflow().Config(filename, asJson)
}

// check if the secret name contains only alphanumeric characters and underscores.
func validateSecretNames(secrets []string) error {
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	for _, secretName := range secrets {
		if !validName.MatchString(secretName) {
			return errors.New("invalid secret name: '" + secretName + "' must contain only alphanumeric characters and underscores")
		}
	}
	return nil
}
