#!/bin/bash

trap "echo 'Script stopped by user.'; exit" SIGINT

counter=0

while true; do
  timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  echo '{"message": "Push log message to stderr","datetime": "'$timestamp'","counter": '$counter'}' >&2
  ((counter++))
  sleep 5
done