#!/bin/bash

: "${MONGO_USERNAME:=default_username}"
: "${MONGO_PASSWORD:=default_password}"

export MONGO_USERNAME
export MONGO_PASSWORD

while true; do
    echo "$(date) - Starting fermat-linux..."
    ./fermat-linux
    echo "$(date) - fermat-linux stopped unexpectedly. Restarting in 5 seconds..."
    sleep 2
done
