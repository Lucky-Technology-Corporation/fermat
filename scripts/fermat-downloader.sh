#!/bin/bash

# Authenticate and fetch an access token
echo "Fetching access token..."
TOKEN=$(gcloud auth print-access-token)
if [[ -z "$TOKEN" ]]; then
    echo "Failed to fetch access token!"
    exit 1
fi
echo "Access token fetched successfully."

# Define GCS URL for direct object access
GCS_URL="https://storage.googleapis.com/swizzle_scripts/fermat-linux"

# Download the file and print the response
echo "Attempting to download from $GCS_URL..."
RESPONSE=$(curl -L -H "Authorization: Bearer $TOKEN" -o fermat-linux "$GCS_URL" -w '%{http_code}' -s)
if [ "$RESPONSE" == "200" ]; then
    echo "Download successful!"

    # Make the downloaded file executable
    chmod +x fermat-linux
    echo "The file 'fermat-linux' has been made executable."
else
    echo "Download failed with HTTP status code: $RESPONSE"
    exit 1
fi
