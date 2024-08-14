package main

import "github.com/shykes/gha/internal/dagger"

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
	// Use a sparse git checkout, only including the given paths
	// Example: ["src", "tests", "Dockerfile"]
	// +optional
	sparseCheckout []string,
) *Gha {
	if err := validateSecretNames(secrets); err != nil {
		panic(err) // FIXME
	}
	m.DispatchTriggers = append(m.DispatchTriggers, DispatchTrigger{
		Pipeline: m.pipeline(command, module, runner, secrets, sparseCheckout),
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
