#!/bin/bash

# Activating service account for gcloud
echo "[INFO] Setting up gcloud auth for service account..."
if gcloud auth activate-service-account --key-file=~/.config/gcloud/application_default_credentials.json; then
    echo "[SUCCESS] Service account activated successfully."
else
    echo "[ERROR] Failed to activate service account!"
    exit 1
fi

# Fetching access token
echo "[INFO] Fetching access token..."
TOKEN=$(gcloud auth print-access-token)
if [[ -z "$TOKEN" ]]; then
    echo "[ERROR] Failed to fetch access token!"
    exit 1
fi
echo "[SUCCESS] Access token fetched successfully."

# Downloading from GCS
GCS_URL="https://storage.googleapis.com/swizzle_scripts/fermat-linux"
echo "[INFO] Attempting to download from $GCS_URL..."
RESPONSE=$(curl -L -H "Authorization: Bearer $TOKEN" -o fermat-linux "$GCS_URL" -w '%{http_code}' -s)
if [ "$RESPONSE" == "200" ]; then
    echo "[SUCCESS] Download successful!"
    chmod +x fermat-linux
    echo "[INFO] The file 'fermat-linux' has been made executable."
else
    echo "[ERROR] Download failed with HTTP status code: $RESPONSE"
    exit 1
fi

# Cloning from Google Cloud Source Repositories
REPO_DIR="$HOME/code"
REPO_NAME="swizzle-webserver-template"

echo "[INFO] Attempting to clone '$REPO_NAME' from Google Cloud Source Repositories..."
mkdir -p "$REPO_DIR"
if gcloud source repos clone $REPO_NAME "$REPO_DIR" --project=swizzle-prod --quiet; then
    echo "[SUCCESS] '$REPO_NAME' has been successfully cloned to $REPO_DIR."
else
    echo "[ERROR] Failed to clone '$REPO_NAME'."
    exit 1
fi

# Configuring git username
gitUsername="${GIT_USERNAME:-Swizzle User}"
echo "[INFO] Setting GIT_USERNAME to $gitUsername."
git config --global user.name "$gitUsername"
if [ $? -ne 0 ]; then
    echo "[ERROR] Error setting git username"
    exit 1
else
    echo "[SUCCESS] Git username set successfully."
fi

# Configuring git email
gitEmail="${GIT_EMAIL:-default@swizzle.co}"
echo "[INFO] Setting GIT_EMAIL to $gitEmail."
git config --global user.email "$gitEmail"
if [ $? -ne 0 ]; then
    echo "[ERROR] Error setting git email"
    exit 1
else
    echo "[SUCCESS] Git email set successfully."
fi
