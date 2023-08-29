#!/bin/bash

# Set the path to your service account key file
export GOOGLE_APPLICATION_CREDENTIALS="/home/swizzle_prod_user/.config/gcloud/application_default_credentials.json"
echo "GOOGLE_APPLICATION_CREDENTIALS set to $GOOGLE_APPLICATION_CREDENTIALS"

# Authenticate and fetch an access token
echo "Fetching access token..."
TOKEN=$(gcloud auth print-access-token)
if [[ -z "$TOKEN" ]]; then
    echo "Failed to fetch access token!"
    exit 1
fi
echo "Access token fetched successfully."

# Define GCS URL
GCS_URL="https://storage.cloud.google.com/swizzle_scripts/fermat-linux"

# Download the file and print the response
echo "Attempting to download from $GCS_URL..."
RESPONSE=$(curl -L -H "Authorization: Bearer $TOKEN" -v -o fermat-linux "$GCS_URL" -w '%{http_code}' -s)
if [ "$RESPONSE" == "200" ]; then
    echo "Download successful!"
else
    echo "Download failed with HTTP status code: $RESPONSE"
    exit 1
fi