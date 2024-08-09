package main

import (
	"dagger/dagger-2-gha/internal/dagger"
	"encoding/json"
)

type Workflow struct {
	Name string            `json:"name,omitempty" yaml:"name,omitempty"`
	On   WorkflowTriggers  `json:"on" yaml:"on"`
	Jobs map[string]Job    `json:"jobs" yaml:"jobs"`
	Env  map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

func (w Workflow) Config(filename string) *dagger.Directory {
	contents, err := json.MarshalIndent(w, "", " ")
	if err != nil {
		panic(err)
	}
	return dag.
		Directory().
		WithNewFile(".github/workflows/"+filename, string(contents))
}

type WorkflowTriggers struct {
	Push             *PushEvent             `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequest      *PullRequestEvent      `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	Schedule         []ScheduledEvent       `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	WorkflowDispatch *WorkflowDispatchEvent `json:"workflow_dispatch,omitempty" yaml:"workflow_dispatch,omitempty"`
	// Other event types can be added here...
}

type PushEvent struct {
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type PullRequestEvent struct {
	Types    []string `json:"types,omitempty" yaml:"types,omitempty"`
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type ScheduledEvent struct {
	Cron string `json:"cron" yaml:"cron"`
}

type WorkflowDispatchEvent struct {
	Inputs map[string]DispatchInput `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

type DispatchInput struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Default     string `json:"default,omitempty" yaml:"default,omitempty"`
}

type Job struct {
	RunsOn         string            `json:"runs-on" yaml:"runs-on"`
	Needs          []string          `json:"needs,omitempty" yaml:"needs,omitempty"`
	Steps          []JobStep         `json:"steps" yaml:"steps"`
	Env            map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Strategy       *Strategy         `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	TimeoutMinutes int               `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty"`
	// Other job-specific fields can be added here...
}

type JobStep struct {
	Name           string            `json:"name,omitempty" yaml:"name,omitempty"`
	Uses           string            `json:"uses,omitempty" yaml:"uses,omitempty"`
	Run            string            `json:"run,omitempty" yaml:"run,omitempty"`
	With           map[string]string `json:"with,omitempty" yaml:"with,omitempty"`
	Env            map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	TimeoutMinutes int               `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty"`
	Shell          string            `json:"shell,omitempty" yaml:"shell,omitempty"`
	// Other step-specific fields can be added here...
}

type Strategy struct {
	Matrix      map[string][]string `json:"matrix,omitempty" yaml:"matrix,omitempty"`
	MaxParallel int                 `json:"max-parallel,omitempty" yaml:"max-parallel,omitempty"`
	FailFast    bool                `json:"fail-fast,omitempty" yaml:"fail-fast,omitempty"`
}
