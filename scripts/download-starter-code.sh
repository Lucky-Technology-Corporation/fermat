#!/bin/bash

REPO_DIR="$HOME/code"

if [ ! -d "$REPO_DIR" ]; then
    # Cloning from Google Cloud Source Repositories
    REPO_NAME="swizzle-webserver-template"
    echo "[INFO] Attempting to clone '$REPO_NAME' from Google Cloud Source Repositories..."
    mkdir -p "$REPO_DIR"
    if gcloud source repos clone "$REPO_NAME" "$REPO_DIR" --project="$GCP_PROJECT" --quiet; then
        echo "[SUCCESS] '$REPO_NAME' has been successfully cloned to $REPO_DIR."
    else
        echo "[ERROR] Failed to clone '$REPO_NAME'."
        exit 1
    fi

    # Configuring git username
    gitUsername="${GIT_USERNAME:-Swizzle User}"
    echo "[INFO] Setting GIT_USERNAME to $gitUsername."
    if ! git config --global user.name "$gitUsername"; then
        echo "[ERROR] Error setting git username"
        exit 1
    fi
    echo "[SUCCESS] Git username set successfully."

    # Configuring git email
    gitEmail="${GIT_EMAIL:-default@swizzle.co}"
    echo "[INFO] Setting GIT_EMAIL to $gitEmail."
    if ! git config --global user.email "$gitEmail"; then
        echo "[ERROR] Error setting git email"
        exit 1
    fi
    echo "[SUCCESS] Git email set successfully."

    # Delete git history
    cd "$REPO_DIR"

    if [ -d ".git" ]; then
        echo "[INFO] Removing existing Git history..."
        rm -rf .git

        echo "[INFO] Initializing a new Git repository..."
        git init
        git add .
        git commit -m "Initial commit"

        echo "[SUCCESS] New Git history has been initialized in $REPO_DIR."
    else
        echo "[ERROR] .git directory not found. Are you sure this is a Git repository?"
        exit 1
    fi
fi
