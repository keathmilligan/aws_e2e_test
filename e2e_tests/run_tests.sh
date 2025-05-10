#!/bin/bash

# Script to run end-to-end tests against the deployed API service

# Default values
API_URL=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --api-url)
      API_URL="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Check if API_URL is provided
if [ -z "$API_URL" ]; then
  # Try to get from environment variable
  if [ -z "$API_URL" ]; then
    echo "Error: API URL must be provided via --api-url flag or API_URL environment variable"
    echo "Usage: $0 --api-url https://your-api-url.com"
    exit 1
  fi
fi

echo "Running end-to-end tests against API at: $API_URL"

# Change to the e2e_tests directory
cd "$(dirname "$0")"

# Run the tests
go test -v -api-url="$API_URL"