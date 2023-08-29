#!/bin/bash

# Check if the correct number of arguments is provided
# Expecting the base64 encoded string of the new service account key as an argument
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <base64-encoded-service-account-key>"
    exit 1
fi

ENCODED_KEY=$1
DEFAULT_CREDENTIALS_PATH="$HOME/.config/gcloud/application_default_credentials.json"
TEMP_KEY_PATH="/tmp/temp_service_account_key.json"

echo "Decoding the base64 encoded service account key..."
echo "$ENCODED_KEY" | base64 -d > "$TEMP_KEY_PATH"
if [ $? -ne 0 ]; then
    echo "Error: Failed to decode the base64 string."
    exit 2
fi

echo "Replacing the application_default_credentials.json file..."
if cp "$TEMP_KEY_PATH" "$DEFAULT_CREDENTIALS_PATH"; then
    echo "Successfully replaced the application_default_credentials.json file."
else
    echo "Error: Failed to replace the application_default_credentials.json file."
    rm -f "$TEMP_KEY_PATH"
    exit 3
fi

echo "Authenticating with the new service account key..."
if gcloud auth activate-service-account --key-file="$TEMP_KEY_PATH"; then
    echo "Successfully authenticated with the new service account key."
else
    echo "Error: Failed to authenticate with the new service account key."
    rm -f "$TEMP_KEY_PATH"
    exit 4
fi

rm -f "$TEMP_KEY_PATH"
echo "gcloud auth configuration updated successfully!"
