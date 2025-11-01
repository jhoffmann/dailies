#!/bin/bash

# Parse command line arguments
TAGS=""
while getopts "t:" opt; do
  case $opt in
  t)
    TAGS="$OPTARG"
    ;;
  \?)
    echo "Invalid option: -$OPTARG" >&2
    exit 1
    ;;
  esac
done

# Build API URL
API_URL="http://localhost:9002/api/tasks?completed=false"
if [ -n "$TAGS" ]; then
  API_URL="${API_URL}&tag=${TAGS}"
fi

# Query API with httpie (suppress output on error)
RESPONSE=$(http --check-status --ignore-stdin --timeout=2 "$API_URL" 2>/dev/null)

# Exit silently if httpie failed (API offline or error)
if [ $? -ne 0 ]; then
  exit 0
fi

# Get task count
TASK_COUNT=$(echo "$RESPONSE" | jq 'length')

# Exit silently if no tasks found
if [ "$TASK_COUNT" -eq 0 ]; then
  exit 0
fi

# Build tooltip with task names and ages
# Strip both nanoseconds AND timezone offset to get basic ISO8601 format
TOOLTIP=$(echo "$RESPONSE" | jq -r '.[] | 
  (.updated_at | sub("\\.[0-9]+(Z|[+-][0-9]{2}:[0-9]{2})$"; "Z") | fromdateiso8601) as $updated |
  ((now - $updated) / 3600 | floor) as $hours |
  "\(.name) (\($hours)h)"' | paste -sd '\n')

# Generate JSON output using jo
jo text=" ${TASK_COUNT}" tooltip="${TOOLTIP}" class="[]" percentage=100
