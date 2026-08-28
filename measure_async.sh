#!/usr/bin/env bash

# Require that variables be set before use.
set -u

api='http://localhost:4000/v1'

# Record the start time in nanoseconds.
start_ns=$(date +%s%N)

# Submit HTTP POST request to generate a report.
submit=$(curl --silent --show-error \
  --write-out $'\n%{time_total}' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data @request.json \
  "$api/reports")

# Extract ack_time value, reponse body, and the job_id from the response.
ack_time=${submit##*$'\n'}
body=${submit%$'\n'*}
job_id=$(printf '%s' "$body" | jq -r '.job_id')

# Continuously poll and count the number of polls until the job is completed or failed.
polls=0
while true; do
  job=$(curl --silent --show-error "$api/jobs/$job_id")
  status=$(printf '%s' "$job" | jq -r '.job.status')
  polls=$((polls + 1))
  printf 'poll=%d status=%s\n' "$polls" "$status"
  case "$status" in completed|failed) break ;; esac
  sleep 0.25 # every 250ms
done

# Record the end time in nanoseconds.
end_ns=$(date +%s%N)

# Calculate the total time taken for the job to complete.
complete_time=$(awk -v s="$start_ns" -v e="$end_ns" \
  'BEGIN { printf "%.3f", (e-s)/1000000000 }')

# Print the job_id, ack_time, complete_time, number of polls, and final status.
printf 'job_id=%s\nack_time=%ss\ncomplete_time=%ss\npolls=%d\nfinal=%s\n' \
  "$job_id" "$ack_time" "$complete_time" "$polls" "$status"
