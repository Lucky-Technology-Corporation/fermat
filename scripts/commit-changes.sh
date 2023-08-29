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

# Push the changes using gcloud source repos
gcloud source repos push "${REPO_NAME}" -- git-dir="${REPO_DIR}/.git"
gcloud_exit_status=$?

if [[ $gcloud_exit_status -eq 0 ]]; then
  printf "[SUCCESS] Changes pushed successfully to %s with commit message: %s.\n" "${REPO_NAME}" "${COMMIT_MESSAGE}"
else
  printf "[ERROR] Failed to push changes to %s.\n" "${REPO_NAME}"
  exit 1
fi

