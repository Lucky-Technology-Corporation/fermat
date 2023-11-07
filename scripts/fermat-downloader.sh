#!/bin/bash

# Whether we're restoring from a snapshot or initializing from scratch
SNAPSHOT_MODE=false

for arg in "$@"
do
    case $arg in
        --snapshot-mode)
        SNAPSHOT_MODE=true
        shift
        ;;
        *)
        shift
        ;;
    esac
done

if [ "$SNAPSHOT_MODE" = "true" ]; then
    echo "[INFO] Running in snapshot mode..."
else
    echo "[INFO] Running in initialization mode..."
fi

# Ensure environment variables are set
if [[ -z "$SERVICE_ACCOUNT_JSON_BASE64" || -z "$BUCKET_NAME" || -z "$GCP_PROJECT" ]]; then
    echo "[ERROR] One or more required environment variables (SERVICE_ACCOUNT_JSON_BASE64, BUCKET_NAME, GCP_PROJECT) are missing."
    exit 1
fi

# Check for required commands
for cmd in gcloud curl git; do
    if ! command -v "$cmd" > /dev/null; then
        echo "[ERROR] Required command '$cmd' not found."
        exit 1
    fi
done

# Decoding and saving service account credentials
DEFAULT_SERVICE_ACCOUNT_FILE="$HOME/.config/gcloud/application_default_credentials.json"
FERMAT_SERVICE_ACCOUNT_FILE="$HOME/.config/gcloud/fermat-keys.json"

echo "[INFO] Decoding and saving service account credentials..."

SERVICE_ACCOUNT_JSON=$(echo "$SERVICE_ACCOUNT_JSON_BASE64" | base64 --decode)
if [[ $? -ne 0 ]]; then
    echo "[ERROR] Failed to decode service account credentials."
    exit 1
fi

echo $SERVICE_ACCOUNT_JSON > "$DEFAULT_SERVICE_ACCOUNT_FILE"
echo $SERVICE_ACCOUNT_JSON > "$FERMAT_SERVICE_ACCOUNT_FILE"

echo "[SUCCESS] Service account credentials saved to $FERMAT_SERVICE_ACCOUNT_FILE."
echo "[SUCCESS] Set default account credentials to fermat service account"

# Activating service account for gcloud
echo "[INFO] Setting up gcloud auth for service account..."
if gcloud auth activate-service-account --key-file="$FERMAT_SERVICE_ACCOUNT_FILE"; then
    echo "[SUCCESS] Service account activated successfully."
else
    echo "[ERROR] Failed to activate service account!"
    exit 1
fi

# Fetching access token
echo "[INFO] Fetching access token..."
ACCESS_TOKEN=$(gcloud auth print-access-token)
if [[ -z "$ACCESS_TOKEN" ]]; then
    echo "[ERROR] Failed to fetch access token!"
    exit 1
fi
echo "[SUCCESS] Access token fetched successfully."

# Downloading from GCS
GCS_URL="https://storage.googleapis.com/$BUCKET_NAME/fermat-linux"
echo "[INFO] Attempting to download from $GCS_URL..."
HTTP_RESPONSE=$(curl -L -H "Authorization: Bearer $ACCESS_TOKEN" -o fermat-linux "$GCS_URL" -w '%{http_code}' -s)
if [ "$HTTP_RESPONSE" = "200" ]; then
    echo "[SUCCESS] Download successful!"
    chmod +x fermat-linux
    echo "[INFO] The file 'fermat-linux' has been made executable."
else
    echo "[ERROR] Download failed with HTTP status code: $HTTP_RESPONSE"
    exit 1
fi

if [ "$SNAPSHOT_MODE" = "false" ]; then
    # Cloning from Google Cloud Source Repositories
    REPO_DIR="$HOME/code"
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
