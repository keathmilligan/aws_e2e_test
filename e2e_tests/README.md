# End-to-End Tests for AWS E2E API Service

This directory contains end-to-end tests for verifying that the API service deployed to AWS is available and working correctly.

## Overview

The tests verify:

1. The health endpoint is accessible and returns the expected response
2. The message creation endpoint works correctly
3. The message retrieval endpoint works correctly and returns the created message

## Prerequisites

- Go 1.18 or later
- Access to the deployed API service

## Running the Tests

You can run the tests in several ways:

### Using the Makefile

```bash
# Run the tests
make test API_URL=https://your-api-url.com

# Or set the API_URL environment variable
export API_URL="https://your-api-url.com"
make test

# Install dependencies
make deps

# Show help
make help
```

### Using the shell script

```bash
./run_tests.sh --api-url https://your-api-url.com
```

### Directly using Go

```bash
go test -v -api-url="https://your-api-url.com"
```

The API URL can also be provided via the `API_URL` environment variable:

```bash
export API_URL="https://your-api-url.com"
./run_tests.sh
```

## GitHub Actions Integration

The tests are automatically run as part of the CI/CD pipeline after each successful deployment. The GitHub Actions workflow:

1. Runs after the deployment workflow completes successfully
2. Gets the API URL from the CloudFormation stack outputs
3. Runs the end-to-end tests against the deployed API

You can also manually trigger the tests from the GitHub Actions UI by providing an API URL.

## Test Details

### Health Check Test

Verifies that the `/health` endpoint returns a 200 OK response with the expected JSON payload.

### Message Creation and Retrieval Test

1. Creates a new message with a unique text
2. Retrieves all messages
3. Verifies that the created message is included in the list of messages

## Distributed System Considerations

The API service is deployed with multiple instances for high availability. This introduces some considerations for testing:

1. **Eventual Consistency**: When a message is created in one instance, there might be a slight delay before it's visible to all instances. The tests include a short delay (2 seconds) after creating a message to account for this eventual consistency.

2. **Persistent Storage**: The API uses DynamoDB for persistent storage to ensure data is consistent across all instances.

3. **Load Balancing**: Requests might be routed to different instances, so a message created in one request might need to be retrieved by a different instance in a subsequent request.

## Troubleshooting

If the tests fail, check:

1. The API service is running and accessible from the GitHub Actions runner
2. The API URL is correct
3. The security groups allow traffic from the GitHub Actions runner
4. The health endpoint is working correctly
5. The message endpoints are working correctly
6. If the message creation/retrieval test fails:
   - Check if there's a DynamoDB table provisioned and accessible
   - Verify the ECS task has the correct IAM permissions to access DynamoDB
   - Increase the delay after message creation if needed (in main_test.go)
   - Check CloudWatch logs for any errors related to DynamoDB access