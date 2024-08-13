// Generate Github Actions configurations from Dagger pipelines
//
// Daggerizing your CI makes your YAML configurations smaller, but they still exist,
// and they're still a pain to maintain by hand.
//
// This module aims to finish the job, by letting you generate your remaining
// YAML configuration from a Dagger pipeline, written in your favorite language.
package main

import (
	"context"
	"dagger/dagger-2-gha/internal/dagger"
	"fmt"
	"strings"
)

func New(
	// Disable sending traces to Dagger Cloud
	// +optional
	noTraces bool,
	// Public Dagger Cloud token, for open-source projects. DO NOT PASS YOUR PRIVATE DAGGER CLOUD TOKEN!
	// This is for a special "public" token which can safely be shared publicly.
	// To get one, contact support@dagger.io
	// +optional
	publicToken string,
	// Dagger version to run in the Github Actions pipelines
	// +optional
	// +default="latest"
	daggerVersion string,
	// Explicitly stop the Dagger Engine after completing the pipeline
	// +optional
	stopEngine bool,
	// Encode all files as JSON (which is also valid YAML)
	// +optional
	asJson bool,
	// Configure a default runner for all workflows
	// See https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/using-self-hosted-runners-in-a-workflow
	// +optional
	// +default="ubuntu-latest"
	runner string,
) *Dagger2Gha {
	return &Dagger2Gha{
		PublicToken:   publicToken,
		NoTraces:      noTraces,
		DaggerVersion: daggerVersion,
		StopEngine:    stopEngine,
		AsJson:        asJson,
		Runner:        runner,
	}
}

type Dagger2Gha struct {
	// +private
	PushTriggers []PushTrigger
	// +private
	PullRequestTriggers []PullRequestTrigger
	// +private
	DispatchTriggers []DispatchTrigger
	// +private
	PublicToken string
	// +private
	DaggerVersion string
	// +private
	NoTraces bool
	// +private
	StopEngine bool
	// +private
	AsJson bool
	// +private
	Runner string
}

// Add a trigger to execute a Dagger pipeline on a git push
func (m *Dagger2Gha) OnPush(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// The Dagger module to load
	// +optional
	// +default="."
	module string,
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
) *Dagger2Gha {
	m.PushTriggers = append(m.PushTriggers, PushTrigger{
		Event: PushEvent{
			Branches: branches,
			Tags:     tags,
			Paths:    paths,
		},
		Pipeline: m.pipeline(command, module, runner),
	})
	return m
}

// Add a trigger to execute a Dagger pipeline on a pull request
func (m *Dagger2Gha) OnPullRequest(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// Dagger module to load
	// +optional
	// +default="."
	module string,
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
) *Dagger2Gha {
	m.PullRequestTriggers = append(m.PullRequestTriggers, PullRequestTrigger{
		Event: PullRequestEvent{
			Branches: branches,
			Types:    types,
		},
		Pipeline: m.pipeline(command, module, runner),
	})
	return m
}

// Add a trigger to execute a Dagger pipeline on a workflow dispatch event
func (m *Dagger2Gha) OnDispatch(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// Dagger module to load
	// +optional
	// +default="."
	module string,
	// Dispatch jobs to the given runner
	// +optional
	runner string,
) *Dagger2Gha {
	m.DispatchTriggers = append(m.DispatchTriggers, DispatchTrigger{
		Pipeline: m.pipeline(command, module, runner),
		Event: WorkflowDispatchEvent{
			Inputs: nil, // FIXME: add inputs, could be pretty dope
		},
	})
	return m
}

func (m *Dagger2Gha) pipeline(
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	module string,
	runner string,
) Pipeline {
	p := Pipeline{
		DaggerVersion: m.DaggerVersion,
		PublicToken:   m.PublicToken,
		NoTraces:      m.NoTraces,
		StopEngine:    m.StopEngine,
		AsJson:        m.AsJson,
		Runner:        m.Runner,
		Command:       command,
		Module:        module,
	}
	if runner != "" {
		p.Runner = runner
	}
	return p
}

// A Dagger pipeline to be called from a Github Actions configuration
type Pipeline struct {
	// +private
	DaggerVersion string
	// +private
	PublicToken string
	// +private
	Module string
	// +private
	Command string
	// +private
	NoTraces bool
	// +private
	StopEngine bool
	// +private
	AsJson bool
	// +private
	Runner string
}

// Generate a github config directory, usable as an overlay on the repository root
func (m *Dagger2Gha) Config(
	// Prefix to use for generated workflow filenames
	// +optional
	prefix string,
) *dagger.Directory {
	dir := dag.Directory()
	for i, t := range m.PushTriggers {
		filename := fmt.Sprintf("%spush-%d.yml", prefix, i+1)
		dir = dir.WithDirectory(".", t.Config(filename, m.AsJson))
	}
	for i, t := range m.PullRequestTriggers {
		filename := fmt.Sprintf("%spr-%d.yml", prefix, i+1)
		dir = dir.WithDirectory(".", t.Config(filename, m.AsJson))
	}
	for i, t := range m.DispatchTriggers {
		filename := fmt.Sprintf("%sdispatch-%d.yml", prefix, i+1)
		dir = dir.WithDirectory(".", t.Config(filename, m.AsJson))
	}
	return dir
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

func (p *Pipeline) Name() string {
	return strings.SplitN(p.Command, " ", 2)[0]
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

// Generate a GHA workflow from a Dagger pipeline definition.
// The workflow will have no triggers, they should be filled separately.
func (p *Pipeline) asWorkflow() Workflow {
	workflow := Workflow{
		Name: p.Command,
		On:   WorkflowTriggers{}, // Triggers intentionally left blank
		Jobs: map[string]Job{
			"dagger": Job{
				RunsOn: p.Runner,
				Steps: []JobStep{
					p.checkoutStep(),
					p.installDaggerStep(),
					p.callDaggerStep(),
				},
			},
		},
	}
	return workflow
}

func (p *Pipeline) checkoutStep() JobStep {
	return JobStep{
		Name: "Checkout",
		Uses: "actions/checkout@v4",
	}
}

func (p *Pipeline) installDaggerStep() JobStep {
	return p.bashStep("scripts/install-dagger.sh", map[string]string{
		"DAGGER_VERSION": p.DaggerVersion,
	})
}

func (p *Pipeline) callDaggerStep() JobStep {
	step := JobStep{
		Name:  "dagger call",
		Shell: "bash",
		Run:   "dagger call " + p.Command,
		Env:   map[string]string{},
	}
	if p.Module != "" {
		step.Env["DAGGER_MODULE"] = p.Module
	}
	if !p.NoTraces {
		if p.PublicToken != "" {
			step.Env["DAGGER_CLOUD_TOKEN"] = p.PublicToken
			// For backwards compatibility with older engines
			step.Env["_EXPERIMENTAL_DAGGER_CLOUD_TOKEN"] = p.PublicToken
		} else {
			step.Env["DAGGER_CLOUD_TOKEN"] = "${{ secrets.DAGGER_CLOUD_TOKEN }}"
			// For backwards compatibility with older engines
			step.Env["_EXPERIMENTAL_DAGGER_CLOUD_TOKEN"] = "${{ secrets.DAGGER_CLOUD_TOKEN }}"
		}
	}
	return step
}

func (p *Pipeline) stopEngineStep() JobStep {
	return p.bashStep("scripts/stop-engine.sh", nil)
}

// Return a github actions step which executes the script embedded at <filename>.
// The script must be checked in with the module source code.
func (p *Pipeline) bashStep(filename string, env map[string]string) JobStep {
	script, err := dag.
		CurrentModule().
		Source().
		File(filename).
		Contents(context.Background())
	if err != nil {
		// We skip error checking for simplicity
		// (don't want to plumb error checking everywhere)
		panic(err)
	}
	return JobStep{
		Name:  filename,
		Shell: "bash",
		Run:   script,
		Env:   env,
	}
}
