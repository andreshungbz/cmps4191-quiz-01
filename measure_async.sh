#!/usr/bin/env bash
set -u

api='http://localhost:4000/v1'
start_ns=$(date +%s%N)

submit=$(curl --silent --show-error \
  --write-out $'\n%{time_total}' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data @request.json \
  "$api/reports")

ack_time=${submit##*$'\n'}
body=${submit%$'\n'*}
job_id=$(printf '%s' "$body" | jq -r '.job_id')
polls=0

while true; do
  job=$(curl --silent --show-error "$api/jobs/$job_id")
  status=$(printf '%s' "$job" | jq -r '.job.status')
  polls=$((polls + 1))
  printf 'poll=%d status=%s\n' "$polls" "$status"
  case "$status" in completed|failed) break ;; esac
  sleep 0.25
done

end_ns=$(date +%s%N)
complete_time=$(awk -v s="$start_ns" -v e="$end_ns" \
  'BEGIN { printf "%.3f", (e-s)/1000000000 }')
printf 'job_id=%s\nack_time=%ss\ncomplete_time=%ss\npolls=%d\nfinal=%s\n' \
  "$job_id" "$ack_time" "$complete_time" "$polls" "$status"
