#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <path-to-git-repo> <new-remote-url>"
    exit 1
fi

REPO_PATH=$1
NEW_REMOTE_URL=$2

cd "$REPO_PATH" || { echo "Error: Failed to change to directory $REPO_PATH."; exit 2; }

if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    echo "Error: The directory $REPO_PATH is not a git repository."
    exit 3
fi

echo "Checking current remote settings in $REPO_PATH..."
git remote -v

echo "Removing current remote 'origin'..."
git remote remove origin

echo "Adding new remote 'origin' with URL: $NEW_REMOTE_URL..."
git remote add origin "$NEW_REMOTE_URL"

echo "New remote settings:"
git remote -v

echo "Remote changed successfully!"
