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
}

func (m *Gha) Check(ctx context.Context, repo *dagger.Directory) (*Gha, error) {
	for _, p := range m.Pipelines {
		if err := p.Check(ctx, repo); err != nil {
			return m, err
		}
	}
	return m, nil
}

// Generate a github config directory, usable as an overlay on the repository root
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

// Register a new dagger pipeline, with no triggers
// Use the pipeline name to attach triggers in follow-up calls
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
	if runner != "" {
		p.Settings.Runner = runner
	}
	m.Pipelines = append(m.Pipelines, p)
	return m
}

// Add a trigger to execute a Dagger pipeline on an issue comment
func (m *Gha) OnIssueComment(
	// Pipelines to trigger
	pipelines []string,
	// Run only for certain types of issue comment events
	// See https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#issue_comment
	// +optional
	types []string,
) (*Gha, error) {
	for _, name := range pipelines {
		p := m.pipeline(name)
		if p == nil {
			return m, fmt.Errorf("pipeline not found: %s", name)
		}
		if p.Triggers.IssueComment == nil {
			p.Triggers.IssueComment = &IssueCommentEvent{}
		}
		p.Triggers.IssueComment.Types = append(p.Triggers.IssueComment.Types, types...)
	}
	return m, nil
}

// Add a trigger to execute a Dagger pipeline on a pull request
func (m *Gha) OnPullRequest(
	// Pipelines to trigger
	pipelines []string,
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
) (*Gha, error) {
	for _, name := range pipelines {
		p := m.pipeline(name)
		if p == nil {
			return m, fmt.Errorf("pipeline not found: %s", name)
		}
		if p.Triggers.PullRequest == nil {
			p.Triggers.PullRequest = &PullRequestEvent{}
		}
		p.Triggers.PullRequest.Types = append(p.Triggers.PullRequest.Types, types...)
		p.Triggers.PullRequest.Branches = append(p.Triggers.PullRequest.Branches, branches...)
		p.Triggers.PullRequest.Paths = append(p.Triggers.PullRequest.Paths, paths...)
	}
	return m, nil
}

// Add a trigger to execute a Dagger pipeline on a git push
func (m *Gha) OnPush(
	// Pipelines to trigger
	pipelines []string,
	// Run only on push to specific branches
	// +optional
	branches []string,
	// Run only on push to specific tags
	// +optional
	tags []string,
	// Run only on push to specific paths
	// +optional
	paths []string,
) (*Gha, error) {
	for _, name := range pipelines {
		p := m.pipeline(name)
		if p == nil {
			return m, fmt.Errorf("pipeline not found: %s", name)
		}
		if p.Triggers.Push == nil {
			p.Triggers.Push = &PushEvent{}
		}
		p.Triggers.Push.Branches = append(p.Triggers.Push.Branches, branches...)
		p.Triggers.Push.Tags = append(p.Triggers.Push.Tags, tags...)
		p.Triggers.Push.Paths = append(p.Triggers.Push.Paths, paths...)
	}
	return m, nil
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
		Name: p.Name,
		On:   p.Triggers,
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
