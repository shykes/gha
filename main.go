// Manage Github Actions configurations with Dagger
//
// Daggerizing your CI makes your YAML configurations smaller, but they still exist,
// and they're still a pain to maintain by hand.
//
// This module aims to finish the job, by letting you generate your remaining
// YAML configuration from a Dagger pipeline, written in your favorite language.
package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/shykes/gha/internal/dagger"
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
) *Gha {
	return &Gha{Settings: Settings{
		PublicToken:   publicToken,
		NoTraces:      noTraces,
		DaggerVersion: daggerVersion,
		StopEngine:    stopEngine,
		AsJson:        asJson,
		Runner:        runner,
	}}
}

type Gha struct {
	// +private
	Pipelines []*Pipeline
	// +private
	Settings Settings
}

type Settings struct {
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
	// +private
	PullRequestConcurrency string
}

// Validate a Github Actions configuration (best effort)
func (m *Gha) Validate(ctx context.Context, repo *dagger.Directory) (*Gha, error) {
	for _, p := range m.Pipelines {
		if err := p.Check(ctx, repo); err != nil {
			return m, err
		}
	}
	return m, nil
}

// Export the configuration to a .github directory
func (m *Gha) Config(
	// Prefix to use for generated workflow filenames
	// +optional
	prefix string,
) *dagger.Directory {
	dir := dag.Directory()
	for _, p := range m.Pipelines {
		dir = dir.WithDirectory(".", p.Config())
	}
	return dir
}

// Add a pipeline
func (m *Gha) WithPipeline(
	// Pipeline name
	name string,
	// The Dagger command to execute
	// Example 'build --source=.'
	command string,
	// The Dagger module to load
	// +optional
	module string,
	// Dispatch jobs to the given runner
	// +optional
	runner string,
	// Github secrets to inject into the pipeline environment.
	// For each secret, an env variable with the same name is created.
	// Example: ["PROD_DEPLOY_TOKEN", "PRIVATE_SSH_KEY"]
	// +optional
	secrets []string,
	// Use a sparse git checkout, only including the given paths
	// Example: ["src", "tests", "Dockerfile"]
	// +optional
	sparseCheckout []string,
	// (DEPRECATED) allow this pipeline to be manually "dispatched"
	// +optional
	// +deprecated
	dispatch bool,
	// Disable manual "dispatch" of this pipeline
	// +optional
	noDispatch bool,
	// Enable lfs on git checkout
	// +optional
	lfs bool,
	// The maximum number of minutes to run the pipeline before killing the process
	// +optional
	timeoutMinutes int,
	// Run the pipeline on any issue comment activity
	// +optional
	onIssueComment bool,
	// +optional
	onIssueCommentCreated bool,
	// +optional
	onIssueCommentEdited bool,
	// +optional
	onIssueCommentDeleted bool,
	// Run the pipeline on any pull request activity
	// +optional
	onPullRequest bool,
	// Configure this pipeline's concurrency for each PR.
	// This is triggered when the pipeline is scheduled concurrently on the same PR.
	//   - allow: all instances are allowed to run concurrently
	//   - queue: new instances are queued, and run sequentially
	//   - preempt: new instances run immediately, older ones are canceled
	// Possible values: "allow", "preempt", "queue"
	// +optional
	// +default="allow"
	pullRequestConcurrency string,
	// +optional
	onPullRequestBranches []string,
	// +optional
	onPullRequestPaths []string,
	// +optional
	onPullRequestAssigned bool,
	// +optional
	onPullRequestUnassigned bool,
	// +optional
	onPullRequestLabeled bool,
	// +optional
	onPullRequestUnlabeled bool,
	// +optional
	onPullRequestOpened bool,
	// +optional
	onPullRequestEdited bool,
	// +optional
	onPullRequestClosed bool,
	// +optional
	onPullRequestReopened bool,
	// +optional
	onPullRequestSynchronize bool,
	// +optional
	onPullRequestConverted_to_draft bool,
	// +optional
	onPullRequestLocked bool,
	// +optional
	onPullRequestUnlocked bool,
	// +optional
	onPullRequestEnqueued bool,
	// +optional
	onPullRequestDequeued bool,
	// +optional
	onPullRequestMilestoned bool,
	// +optional
	onPullRequestDemilestoned bool,
	// +optional
	onPullRequestReadyForReview bool,
	// +optional
	onPullRequestReviewRequested bool,
	// +optional
	onPullRequestReviewRequestRemoved bool,
	// +optional
	onPullRequestAutoMergeEnabled bool,
	// +optional
	onPullRequestAutoMergeDisabled bool,
	// Run the pipeline on any git push
	// +optional
	onPush bool,
	// Run the pipeline on git push to the specified tags
	// +optional
	onPushTags []string,
	// Run the pipeline on git push to the specified branches
	// +optional
	onPushBranches []string,
) *Gha {
	p := &Pipeline{
		Name:           name,
		Command:        command,
		Module:         module,
		Secrets:        secrets,
		SparseCheckout: sparseCheckout,
		LFS:            lfs,
		Settings:       m.Settings,
	}
	if !noDispatch {
		p.Triggers.WorkflowDispatch = &WorkflowDispatchEvent{}
	}
	if pullRequestConcurrency != "" {
		p.Settings.PullRequestConcurrency = pullRequestConcurrency
	}
	if runner != "" {
		p.Settings.Runner = runner
	}
	if onIssueComment {
		p.OnIssueComment(nil)
	}
	if onIssueCommentCreated {
		p.OnIssueComment([]string{"created"})
	}
	if onIssueCommentDeleted {
		p.OnIssueComment([]string{"deleted"})
	}
	if onIssueCommentEdited {
		p.OnIssueComment([]string{"edited"})
	}
	if onPullRequest {
		p.OnPullRequest(nil, nil, nil)
	}
	if onPullRequestBranches != nil {
		p.OnPullRequest(nil, onPullRequestBranches, nil)
	}
	if onPullRequestPaths != nil {
		p.OnPullRequest([]string{"paths"}, nil, onPullRequestPaths)
	}
	if onPullRequestAssigned {
		p.OnPullRequest([]string{"assigned"}, nil, nil)
	}
	if onPullRequestUnassigned {
		p.OnPullRequest([]string{"unassigned"}, nil, nil)
	}
	if onPullRequestLabeled {
		p.OnPullRequest([]string{"labeled"}, nil, nil)
	}
	if onPullRequestUnlabeled {
		p.OnPullRequest([]string{"unlabeled"}, nil, nil)
	}
	if onPullRequestOpened {
		p.OnPullRequest([]string{"opened"}, nil, nil)
	}
	if onPullRequestEdited {
		p.OnPullRequest([]string{"edited"}, nil, nil)
	}
	if onPullRequestClosed {
		p.OnPullRequest([]string{"closed"}, nil, nil)
	}
	if onPullRequestReopened {
		p.OnPullRequest([]string{"reopened"}, nil, nil)
	}
	if onPullRequestSynchronize {
		p.OnPullRequest([]string{"synchronize"}, nil, nil)
	}
	if onPullRequestConverted_to_draft {
		p.OnPullRequest([]string{"converted-to-draft"}, nil, nil)
	}
	if onPullRequestLocked {
		p.OnPullRequest([]string{"locked"}, nil, nil)
	}
	if onPullRequestUnlocked {
		p.OnPullRequest([]string{"unlocked"}, nil, nil)
	}
	if onPullRequestEnqueued {
		p.OnPullRequest([]string{"enqueued"}, nil, nil)
	}
	if onPullRequestDequeued {
		p.OnPullRequest([]string{"dequeued"}, nil, nil)
	}
	if onPullRequestMilestoned {
		p.OnPullRequest([]string{"milestoned"}, nil, nil)
	}
	if onPullRequestDemilestoned {
		p.OnPullRequest([]string{"demilestoned"}, nil, nil)
	}
	if onPullRequestReadyForReview {
		p.OnPullRequest([]string{"ready-for-review"}, nil, nil)
	}
	if onPullRequestReviewRequested {
		p.OnPullRequest([]string{"review-requested"}, nil, nil)
	}
	if onPullRequestReviewRequestRemoved {
		p.OnPullRequest([]string{"review-request-removed"}, nil, nil)
	}
	if onPullRequestAutoMergeEnabled {
		p.OnPullRequest([]string{"auto-merge-enabled"}, nil, nil)
	}
	if onPullRequestAutoMergeDisabled {
		p.OnPullRequest([]string{"auto-merge-disabled"}, nil, nil)
	}
	if onPush {
		p.OnPush(nil, nil)
	}
	if onPushBranches != nil {
		p.OnPush(onPushBranches, nil)
	}
	if onPushTags != nil {
		p.OnPush(nil, onPushTags)
	}
	m.Pipelines = append(m.Pipelines, p)
	return m
}

func (p *Pipeline) OnIssueComment(
	// Run only for certain types of issue comment events
	// See https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#issue_comment
	// +optional
	types []string,
) *Pipeline {
	if p.Triggers.IssueComment == nil {
		p.Triggers.IssueComment = &IssueCommentEvent{}
	}
	p.Triggers.IssueComment.Types = append(p.Triggers.IssueComment.Types, types...)
	return p
}

// Add a trigger to execute a Dagger pipeline on a pull request
func (p *Pipeline) OnPullRequest(
	// Run only for certain types of pull request events
	// See https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#pull_request
	// +optional
	types []string,
	// Run only for pull requests that target specific branches
	// +optional
	branches []string,
	// Run only for pull requests that target specific paths
	// +optional
	paths []string,
) *Pipeline {
	if p.Triggers.PullRequest == nil {
		p.Triggers.PullRequest = &PullRequestEvent{}
	}
	p.Triggers.PullRequest.Types = append(p.Triggers.PullRequest.Types, types...)
	p.Triggers.PullRequest.Branches = append(p.Triggers.PullRequest.Branches, branches...)
	p.Triggers.PullRequest.Paths = append(p.Triggers.PullRequest.Paths, paths...)
	return p
}

// Add a trigger to execute a Dagger pipeline on a git push
func (p *Pipeline) OnPush(
	// Run only on push to specific branches
	// +optional
	branches []string,
	// Run only on push to specific tags
	// +optional
	tags []string,
) *Pipeline {
	if p.Triggers.Push == nil {
		p.Triggers.Push = &PushEvent{}
	}
	p.Triggers.Push.Branches = append(p.Triggers.Push.Branches, branches...)
	p.Triggers.Push.Tags = append(p.Triggers.Push.Tags, tags...)
	return p
}

// Lookup a pipeline
func (m *Gha) pipeline(name string) *Pipeline {
	for _, p := range m.Pipelines {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// A Dagger pipeline to be called from a Github Actions configuration
type Pipeline struct {
	// +private
	Name string
	// +private
	Module string
	// +private
	Command string
	// +private
	Secrets []string
	// +private
	SparseCheckout []string
	// +private
	LFS bool
	// +private
	Settings Settings
	// +private
	Triggers WorkflowTriggers
}

func (p *Pipeline) Config() *dagger.Directory {
	return p.asWorkflow().Config(p.workflowFilename(), p.Settings.AsJson)
}

func (p *Pipeline) concurrency() *WorkflowConcurrency {
	setting := p.Settings.PullRequestConcurrency
	if setting == "" || setting == "allow" {
		return nil
	}
	if (setting != "queue") && (setting != "preempt") {
		panic("Unsupported value for 'pullRequestConcurrency': " + setting)
	}
	concurrency := &WorkflowConcurrency{
		// If in a pull request: concurrency group is unique to workflow + head branch
		// If NOT in a pull request: concurrency group is unique to run ID -> no grouping
		Group: "${{ github.workflow }}-${{ github.head_ref || github.run_id }}",
	}
	if setting == "preempt" {
		concurrency.CancelInProgress = true
	}
	return concurrency
}

func (p *Pipeline) checkSecretNames() error {
	// check if the secret name contains only alphanumeric characters and underscores.
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	for _, secretName := range p.Secrets {
		if !validName.MatchString(secretName) {
			return errors.New("invalid secret name: '" + secretName + "' must contain only alphanumeric characters and underscores")
		}
	}
	return nil
}

func (p *Pipeline) checkCommandAndModule(ctx context.Context, repo *dagger.Directory) error {
	script := "dagger call"
	if p.Module != "" {
		script = script + " -m '" + p.Module + "' "
	}
	script = script + p.Command + " --help"
	_, err := dag.
		Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"dagger", "bash"},
		}).
		WithMountedDirectory("/src", repo).
		WithWorkdir("/src").
		WithExec(
			[]string{"bash", "-c", script},
			dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true},
		).
		Sync(ctx)
	return err
}

// Check that the pipeline is valid, in a best effort way
func (p *Pipeline) Check(
	ctx context.Context,
	// +defaultPath="/"
	repo *dagger.Directory,
) error {
	if err := p.checkSecretNames(); err != nil {
		return err
	}
	if err := p.checkCommandAndModule(ctx, repo); err != nil {
		return err
	}
	return nil
}

// Generate a GHA workflow from a Dagger pipeline definition.
// The workflow will have no triggers, they should be filled separately.
func (p *Pipeline) asWorkflow() Workflow {
	steps := []JobStep{
		p.checkoutStep(),
		p.installDaggerStep(),
		p.warmEngineStep(),
		p.callDaggerStep(),
	}
	if p.Settings.StopEngine {
		steps = append(steps, p.stopEngineStep())
	}
	return Workflow{
		Name:        p.Name,
		On:          p.Triggers,
		Concurrency: p.concurrency(),
		Jobs: map[string]Job{
			p.jobID(): Job{
				// The job name is used by the "required checks feature" in branch protection rules
				Name:   p.Name,
				RunsOn: p.Settings.Runner,
				Steps:  steps,
				Outputs: map[string]string{
					"stdout": "${{ steps.exec.outputs.stdout }}",
					"stderr": "${{ steps.exec.outputs.stderr }}",
				},
			},
		},
	}
}

func (p *Pipeline) workflowFilename() string {
	var name string
	// Convert to lowercase
	name = strings.ToLower(p.Name)
	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	name = re.ReplaceAllString(name, "-")
	// Trim leading and trailing hyphens
	name = strings.Trim(name, "-")
	// Add the .yml extension
	return name + ".yml"
}

func (p *Pipeline) jobID() string {
	return "dagger"
}

func (p *Pipeline) checkoutStep() JobStep {
	step := JobStep{
		Name: "Checkout",
		Uses: "actions/checkout@v4",
		With: map[string]string{},
	}
	if p.SparseCheckout != nil {
		// Include common dagger paths in the checkout, to make
		// sure local modules work by default
		// FIXME: this is only a guess, we need the 'source' field of dagger.json
		//  to be sure.
		sparseCheckout := append(p.SparseCheckout, "dagger.json", ".dagger", "dagger", "ci")
		step.With["sparse-checkout"] = strings.Join(sparseCheckout, "\n")
	}
	if p.LFS {
		step.With["lfs"] = "true"
	}
	return step
}

func (p *Pipeline) warmEngineStep() JobStep {
	return p.bashStep("warm-engine", nil)
}

func (p *Pipeline) installDaggerStep() JobStep {
	return p.bashStep("install-dagger", map[string]string{
		"DAGGER_VERSION": p.Settings.DaggerVersion,
	})
}

func (p *Pipeline) callDaggerStep() JobStep {
	env := map[string]string{}
	// Inject dagger command
	env["COMMAND"] = "dagger call -q " + p.Command
	// Inject user-defined secrets
	for _, secretName := range p.Secrets {
		env[secretName] = fmt.Sprintf("${{ secrets.%s }}", secretName)
	}
	// Inject module name
	if p.Module != "" {
		env["DAGGER_MODULE"] = p.Module
	}
	// Inject Dagger Cloud token
	if !p.Settings.NoTraces {
		if p.Settings.PublicToken != "" {
			env["DAGGER_CLOUD_TOKEN"] = p.Settings.PublicToken
			// For backwards compatibility with older engines
			env["_EXPERIMENTAL_DAGGER_CLOUD_TOKEN"] = p.Settings.PublicToken
		} else {
			env["DAGGER_CLOUD_TOKEN"] = "${{ secrets.DAGGER_CLOUD_TOKEN }}"
			// For backwards compatibility with older engines
			env["_EXPERIMENTAL_DAGGER_CLOUD_TOKEN"] = "${{ secrets.DAGGER_CLOUD_TOKEN }}"
		}
	}
	// Inject Github context keys
	// github.ref becomes $GITHUB_REF, etc.
	for _, key := range githubContextKeys {
		env["GITHUB_"+strings.ToUpper(key)] = fmt.Sprintf("${{ github.%s }}", key)
	}
	return p.bashStep("exec", env)
}

func (p *Pipeline) stopEngineStep() JobStep {
	return p.bashStep("scripts/stop-engine.sh", nil)
}

// Return a github actions step which executes the script embedded at scripts/<filename>.sh
// The script must be checked in with the module source code.
func (p *Pipeline) bashStep(id string, env map[string]string) JobStep {
	filename := "scripts/" + id + ".sh"
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
		ID:    id,
		Shell: "bash",
		Run:   script,
		Env:   env,
	}
}
