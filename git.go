package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	repoName   = "swizzle-webserver-template"
	repoDir    = "~/code"
	defaultMsg = "swizzle automatic commit"
)

type RequestBody struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	CommitMessage string `json:"commitMessage"`
	Tag           string `json:"tag,omitempty"`
}

func commitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	var body RequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open repository: %s", err), http.StatusInternalServerError)
		return
	}

	// Ensure on master branch
	headRef, err := repo.Head()
	if err != nil {
		http.Error(w, "Could not fetch HEAD", http.StatusInternalServerError)
		return
	}
	if headRef.Name().Short() != "master" {
		http.Error(w, "Not on master branch", http.StatusForbidden)
		return
	}

	worktree, err := repo.Worktree()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get worktree: %s", err), http.StatusInternalServerError)
		return
	}

	_, err = worktree.Add(".")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add changes to staging area: %s", err), http.StatusInternalServerError)
		return
	}

	gitName := body.Name
	if gitName == "" {
		gitName = "Default Swizzle Committer"
	}

	signature := &object.Signature{
		Name:  gitName,
		Email: body.Email,
		When:  time.Now(),
	}

	commitHash, err := worktree.Commit(body.CommitMessage, &git.CommitOptions{
		Author:    signature,
		Committer: signature,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to commit changes: %s", err), http.StatusInternalServerError)
		return
	}

	// Create tag if provided
	if body.Tag != "" {
		_, err = repo.CreateTag(body.Tag, commitHash, &git.CreateTagOptions{
			Tagger:  signature,
			Message: body.CommitMessage,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create tag: %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Send success response
	w.Write([]byte("Commit successful!"))
}

func commitPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		CommitMsg string `json:"commit_msg"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	commitMsg := requestBody.CommitMsg
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("%s: %s", defaultMsg, time.Now())
	}

	repo, err := git.PlainOpen(repoDir)
	CheckIfError(w, err)

	workTree, err := repo.Worktree()
	CheckIfError(w, err)

	_, err = workTree.Add(".")
	CheckIfError(w, err)

	_, err = workTree.Commit(commitMsg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "Your Name",       // TODO: Replace with your name
			Email: "you@example.com", // TODO: Replace with your email
			When:  time.Now(),
		},
	})
	CheckIfError(w, err)

	err = workTree.Checkout(&git.CheckoutOptions{
		Branch: "refs/heads/production",
	})
	CheckIfError(w, err)

	// Get the commit we want to merge.
	masterRef, err := repo.Reference("refs/heads/master", true)
	CheckIfError(w, err)

	_, err = repo.CommitObject(masterRef.Hash())
	CheckIfError(w, err)

	w.Write([]byte("Operation completed successfully"))
}

func CheckIfError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type CommandRunner struct {
	err      error
	dir      string
	output   string
	exitCode int
}

func (runner *CommandRunner) Run(name string, args ...string) {
	if runner.err != nil {
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = runner.dir

	output, err := cmd.CombinedOutput()

	runner.output = string(output)
	runner.exitCode = cmd.ProcessState.ExitCode()

	if err != nil {
		cmdParts := append([]string{name}, args...)
		cmdString := strings.Join(cmdParts, " ")
		runner.err = fmt.Errorf("Failed while running command: '%s'. Error: %v. Command Output:\n%s", cmdString, err, string(runner.output))
	}
}

type PushProductionStatus string

const (
	BuildTriggered PushProductionStatus = "BUILD_TRIGGERED"
	NoChanges      PushProductionStatus = "NO_CHANGES"
	Failed         PushProductionStatus = "FAILED"
)

type PushProductionResponse struct {
	Status  PushProductionStatus `json:"status"`
	Message string               `json:"message,omitempty"`
	Commit  string               `json:"commit,omitempty"`
}

func pushProduction(w http.ResponseWriter, r *http.Request) {
	commitMessage := fmt.Sprintf("swizzle commit production: %s", time.Now().Format(time.RFC3339))

	runner := &CommandRunner{dir: "code"}
	runner.Run("git", "add", ".")
	runner.Run("git", "commit", "-m", commitMessage)

	// Reset the runner because we always want to try a push even if the previous commands errored out.
	// This is because 'git commit -m "..."' will return exit code 1 if there is nothing to commit.
	// However, just because master is up-to-date doesn't mean that our release branch is so we should
	// try doing a push anyways.
	runner = &CommandRunner{dir: "code"}
	runner.Run("git", "push", "-o", "nokeycheck", "origin", "master:release")

	if runner.err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteJSONResponseWithHeader(w, http.StatusInternalServerError, &PushProductionResponse{
			Status:  Failed,
			Message: runner.err.Error(),
		})
		return
	}

	// There were no changes so we should notify the caller that no build will be triggered
	not_build_triggered := strings.Contains(runner.output, "Everything up-to-date")

	// This shouldn't fail, but just in case we'll check for an error
	runner.Run("git", "rev-parse", "HEAD")
	if runner.err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteJSONResponseWithHeader(w, http.StatusInternalServerError, &PushProductionResponse{
			Status:  Failed,
			Message: runner.err.Error(),
		})
		return
	}

	commit := strings.TrimSpace(runner.output)
	status := BuildTriggered
	if not_build_triggered {
		status = NoChanges
	}

	WriteJSONResponse(w, &PushProductionResponse{
		Status: status,
		Commit: commit,
	})
}

func mergeMasterIntoRelease() error {
	cmd := exec.Command("git", "checkout", "release")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "merge", "master", "--no-rebase", "-m", "merge to release")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "checkout", "master")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "merge", "release", "--no-rebase", "-m", "housekeeping: keep branch up to date")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
