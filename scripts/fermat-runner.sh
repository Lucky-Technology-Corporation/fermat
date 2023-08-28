#!/bin/bash

# Check and set environment variables if needed
[ -z "$MONGO_USERNAME" ] && export MONGO_USERNAME=usertest123
[ -z "$MONGO_PASSWORD" ] && export MONGO_PASSWORD=passtest123

while true; do
    ./fermat-linux
    echo "fermat-linux crashed with exit code $?. Respawning..." >&2
    sleep 1
done
