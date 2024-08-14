package main

import (
	"encoding/json"

	"github.com/shykes/gha/internal/dagger"
	"gopkg.in/yaml.v3"
)

const (
	genHeader = "# This file was generated. See https://daggerverse.dev/mod/github.com/shykes/gha"
)

var (
	// List of keys available as '${{github.KEY}}'
	// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/contexts#github-context
	githubContextKeys = []string{
		// The name of the action currently running, or the id of a step. GitHub removes special characters, and uses the name __run when the current step runs a script without an id. If you use the same action more than once in the same job, the name will include a suffix with the sequence number with underscore before it. For example, the first script you run will have the name __run, and the second script will be named __run_2. Similarly, the second invocation of actions/checkout will be actionscheckout2.
		"action",
		// The path where an action is located. This property is only supported in composite actions. You can use this path to access files located in the same repository as the action, for example by changing directories to the path:  cd ${{ github.action_path }} .
		"action_path",
		// For a step executing an action, this is the ref of the action being executed. For example, v2. Do not use in the run keyword. To make this context work with composite actions, reference it within the env context of the composite action.
		"action_ref",
		// For a step executing an action, this is the owner and repository name of the action. For example, actions/checkout. Do not use in the run keyword. To make this context work with composite actions, reference it within the env context of the composite action.
		"action_repository",
		// For a composite action, the current result of the composite action.
		"action_status",
		// The username of the user that triggered the initial workflow run. If the workflow run is a re-run, this value may differ from github.triggering_actor. Any workflow re-runs will use the privileges of github.actor, even if the actor initiating the re-run (github.triggering_actor) has different privileges.
		"actor",
		// The account ID of the person or app that triggered the initial workflow run. For example, 1234567. Note that this is different from the actor username.
		"actor_id",
		// The URL of the GitHub REST API.
		"api_url",
		// The base_ref or target branch of the pull request in a workflow run. This property is only available when the event that triggers a workflow run is either pull_request or pull_request_target.
		"base_ref",
		// Path on the runner to the file that sets environment variables from workflow commands. This file is unique to the current step and is a different file for each step in a job. For more information, see "Workflow commands for GitHub Actions."
		"env",
		// The name of the event that triggered the workflow run.
		"event_name",
		// The path to the file on the runner that contains the full event webhook payload.
		"event_path",
		// The URL of the GitHub GraphQL API.
		"graphql_url",
		// The head_ref or source branch of the pull request in a workflow run. This property is only available when the event that triggers a workflow run is either pull_request or pull_request_target.
		"head_ref",
		// The job_id of the current job. Note: This context property is set by the Actions runner, and is only available within the execution steps of a job. Otherwise, the value of this property will be null.
		"job",
		// Path on the runner to the file that sets system PATH variables from workflow commands. This file is unique to the current step and is a different file for each step in a job. For more information, see "Workflow commands for GitHub Actions."
		"path",
		// The fully-formed ref of the branch or tag that triggered the workflow run. For workflows triggered by push, this is the branch or tag ref that was pushed. For workflows triggered by pull_request, this is the pull request merge branch. For workflows triggered by release, this is the release tag created. For other triggers, this is the branch or tag ref that triggered the workflow run. This is only set if a branch or tag is available for the event type. The ref given is fully-formed, meaning that for branches the format is refs/heads/<branch_name>, for pull requests it is refs/pull/<pr_number>/merge, and for tags it is refs/tags/<tag_name>. For example, refs/heads/feature-branch-1.
		"ref",
		// The short ref name of the branch or tag that triggered the workflow run. This value matches the branch or tag name shown on GitHub. For example, feature-branch-1. For pull requests, the format is <pr_number>/merge.
		"ref_name",
		// true if branch protections or rulesets are configured for the ref that triggered the workflow run.
		"ref_protected",
		// The type of ref that triggered the workflow run. Valid values are branch or tag.
		"ref_type",
		// The owner and repository name. For example, octocat/Hello-World.
		"repository",
		// The ID of the repository. For example, 123456789. Note that this is different from the repository name.
		"repository_id",
		// The repository owner's username. For example, octocat.
		"repository_owner",
		// The repository owner's account ID. For example, 1234567. Note that this is different from the owner's name.
		"repository_owner_id",
		// The Git URL to the repository. For example, git://github.com/octocat/hello-world.git.
		"repositoryUrl",
		// The number of days that workflow run logs and artifacts are kept.
		"retention_days",
		// A unique number for each workflow run within a repository. This number does not change if you re-run the workflow run.
		"run_id",
		// A unique number for each run of a particular workflow in a repository. This number begins at 1 for the workflow's first run, and increments with each new run. This number does not change if you re-run the workflow run.
		"run_number",
		// A unique number for each attempt of a particular workflow run in a repository. This number begins at 1 for the workflow run's first attempt, and increments with each re-run.
		"run_attempt",
		// The source of a secret used in a workflow. Possible values are None, Actions, Codespaces, or Dependabot.
		"secret_source",
		// The URL of the GitHub server. For example: https://github.com.
		"server_url",
		// The commit SHA that triggered the workflow. The value of this commit SHA depends on the event that triggered the workflow. For more information, see "Events that trigger workflows." For example, ffac537e6cbbf934b08745a378932722df287a53.
		"sha",
		// A token to authenticate on behalf of the GitHub App installed on your repository. This is functionally equivalent to the GITHUB_TOKEN secret. For more information, see "Automatic token authentication." Note: This context property is set by the Actions runner, and is only available within the execution steps of a job. Otherwise, the value of this property will be null.
		"token",
		// The username of the user that initiated the workflow run. If the workflow run is a re-run, this value may differ from github.actor. Any workflow re-runs will use the privileges of github.actor, even if the actor initiating the re-run (github.triggering_actor) has different privileges.
		"triggering_actor",
		// The name of the workflow. If the workflow file doesn't specify a name, the value of this property is the full path of the workflow file in the repository.
		"workflow",
		// The ref path to the workflow. For example, octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch.
		"workflow_ref",
		// The commit SHA for the workflow file.
		"workflow_sha",
		// The default working directory on the runner for steps, and the default location of your repository when using the checkout action.
		"workspace",
	}
)

type Workflow struct {
	Name string            `json:"name,omitempty" yaml:"name,omitempty"`
	On   WorkflowTriggers  `json:"on" yaml:"on"`
	Jobs map[string]Job    `json:"jobs" yaml:"jobs"`
	Env  map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

// Generate an overlay config directory for this workflow
func (w Workflow) Config(
	// Filename of the workflow file under .github/workflows/
	filename string,
	// Encode the workflow as JSON, which is valid YAML
	asJson bool,
) *dagger.Directory {
	var (
		contents []byte
		err      error
	)
	if asJson {
		contents, err = json.MarshalIndent(w, "", " ")
	} else {
		contents, err = yaml.Marshal(w)
	}
	if err != nil {
		panic(err)
	}
	return dag.
		Directory().
		WithNewFile(".github/workflows/"+filename, genHeader+"\n"+string(contents))
}

type WorkflowTriggers struct {
	Push             *PushEvent             `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequest      *PullRequestEvent      `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	Schedule         []ScheduledEvent       `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	WorkflowDispatch *WorkflowDispatchEvent `json:"workflow_dispatch,omitempty" yaml:"workflow_dispatch,omitempty"`
	IssueComment     *IssueCommentEvent     `json:"issue_comment,omitempty" yaml:"issue_comment,omitempty"`
}

type PushEvent struct {
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type PullRequestEvent struct {
	Types    []string `json:"types,omitempty" yaml:"types,omitempty"`
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type ScheduledEvent struct {
	Cron string `json:"cron" yaml:"cron"`
}

type WorkflowDispatchEvent struct {
	// FIXME: The Dagger API can't serialize maps
	// Inputs map[string]DispatchInput `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

type IssueCommentEvent struct {
	Types []string `json:"types,omitempty" yaml:"types,omitempty"`
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
	Outputs        map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type JobStep struct {
	Name           string            `json:"name,omitempty" yaml:"name,omitempty"`
	ID             string            `json:"id,omitempty" yaml:"id,omitempty"`
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
