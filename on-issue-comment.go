package main

import "github.com/shykes/gha/internal/dagger"

// Add a trigger to execute a Dagger pipeline on an issue comment
func (m *Gha) OnIssueComment(
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
	// Dispatch jobs to the given runner
	// +optional
	runner string,
	// Use a sparse git checkout, only including the given paths
	// Example: ["src", "tests", "Dockerfile"]
	// +optional
	sparseCheckout []string,
	// Run only for certain types of issue comment events
	// See https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#issue_comment
	// +optional
	types []string,
) *Gha {
	if err := validateSecretNames(secrets); err != nil {
		panic(err) // FIXME
	}
	m.IssueCommentTriggers = append(m.IssueCommentTriggers, IssueCommentTrigger{
		Event: IssueCommentEvent{
			Types: types,
		},
		Pipeline: m.pipeline(command, module, runner, secrets, sparseCheckout),
	})
	return m
}

type IssueCommentTrigger struct {
	// When this happens...
	Event IssueCommentEvent
	// ...do this
	Pipeline Pipeline
}

func (t IssueCommentTrigger) asWorkflow() Workflow {
	var workflow = t.Pipeline.asWorkflow()
	workflow.On = WorkflowTriggers{IssueComment: &(t.Event)}
	return workflow
}

func (t IssueCommentTrigger) Config(filename string, asJson bool) *dagger.Directory {
	return t.asWorkflow().Config(filename, asJson)
}
