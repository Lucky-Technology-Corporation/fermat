#!/bin/bash

# Define repository details
REPO_NAME="swizzle-webserver-template"
REPO_DIR="$HOME/code"

# Check if a commit message was provided, else use default
if [ "$#" -eq 1 ]; then
  COMMIT_MESSAGE="$1"
else
  TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")
  COMMIT_MESSAGE="default swizzle commit: $TIMESTAMP"
fi

# Navigate to the repository directory
cd "$REPO_DIR" || {
  echo "[ERROR] Failed to navigate to $REPO_DIR"
  exit 1
}

# Add all changes to the staging area
git add .

# Commit the changes
git commit -m "$COMMIT_MESSAGE"

# Push the master branch using gcloud source repos
gcloud source repos push "${REPO_NAME}" -- git-dir="${REPO_DIR}/.git" --branch master
gcloud_exit_status=$?

if [[ $gcloud_exit_status -ne 0 ]]; then
  printf "[ERROR] Failed to push changes to master branch of %s.\n" "${REPO_NAME}"
  exit 1
fi

# Switch to production branch
git checkout production
if [[ $? -ne 0 ]]; then
  printf "[ERROR] Failed to checkout production branch.\n"
  exit 1
fi

# Merge master into production
git merge master --no-ff -m "Merge master into production"
if [[ $? -ne 0 ]]; then
  printf "[ERROR] Failed to merge master into production branch.\n"
  exit 1
fi

# Create a tag for the release
RELEASE_TAG="release-$(date "+%Y%m%d%H%M%S")"
git tag "$RELEASE_TAG"

# Push the production branch and the new tag using gcloud source repos
gcloud source repos push "${REPO_NAME}" -- git-dir="${REPO_DIR}/.git" --branch production
gcloud source repos push "${REPO_NAME}" -- git-dir="${REPO_DIR}/.git" --tags
gcloud_exit_status=$?

if [[ $gcloud_exit_status -ne 0 ]]; then
  printf "[ERROR] Failed to push changes and/or tags to production branch of %s.\n" "${REPO_NAME}"
  exit 1
fi

# Switch back to master branch
git checkout master
if [[ $? -ne 0 ]]; then
  printf "[ERROR] Failed to checkout master branch.\n"
  exit 1
fi

# Merge production (which includes the merge commit) back into master
git merge production --no-ff -m "re-merge production into master after release"
if [[ $? -ne 0 ]]; then
  printf "[ERROR] Failed to re-merge production into master branch.\n"
  exit 1
fi

# Push the master branch with the new merge commit
gcloud source repos push "${REPO_NAME}" -- git-dir="${REPO_DIR}/.git" --branch master
printf "[SUCCESS] Changes and release tag pushed successfully to %s with commit message: %s.\n" "${REPO_NAME}" "${COMMIT_MESSAGE}"
