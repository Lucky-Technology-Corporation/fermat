#!/bin/bash

export GOOGLE_APPLICATION_CREDENTIALS="/home/swizzle_prod_user/.config/gcloud/application_default_credentials.json"

GCS_URL="https://storage.cloud.google.com/swizzle_scripts/fermat"
TOKEN=$(gcloud auth print-access-token)

if curl -H "Authorization: Bearer $TOKEN" -o fermat-linux "$GCS_URL"; then
    echo "Download successful!"
else
    echo "Download failed!"
fi
