# AWS End-to-End POC

This project demonstrates a full-stack application deployed on AWS using CloudFormation, ECS, and ALB.

## Project Structure

- `api/` - Go API service using Gin framework
- `frontend/` - React/TypeScript frontend application
- `iac/` - Infrastructure as Code (CloudFormation templates)
- `e2e_tests/` - End-to-end tests for the deployed API service

## Architecture

The application consists of:

1. A Go API service deployed as an ECS Fargate task
2. An Application Load Balancer (ALB) in front of the ECS service
3. A DynamoDB table for persistent message storage
4. A React frontend application
5. CloudFormation templates for infrastructure provisioning
6. GitHub Actions workflows for CI/CD

## API Service

The API service provides:

- A health check endpoint (`/health`)
- Endpoints for creating and retrieving messages (`/messages`)
- Persistent storage of messages in DynamoDB

## Deployment

The application is deployed using GitHub Actions workflows:

1. The `deploy.yaml` workflow builds and deploys the application to AWS
2. The `e2e_tests.yaml` workflow runs end-to-end tests against the deployed API

## End-to-End Tests

The project includes end-to-end tests to verify that the API service is available and working correctly. These tests:

1. Check the health endpoint
2. Create a new message
3. Retrieve all messages and verify the created message is included

### Running the End-to-End Tests

You can run the end-to-end tests locally:

```bash
cd e2e_tests
./run_tests.sh --api-url https://your-api-url.com
```

Or using the GitHub Actions workflow:

1. Go to the "Actions" tab in GitHub
2. Select the "End-to-End Tests" workflow
3. Click "Run workflow"
4. Optionally provide an API URL (if not provided, it will be retrieved from CloudFormation)

See the [e2e_tests/README.md](e2e_tests/README.md) file for more details.

## Local Development

### API Service

The API service uses an in-memory store by default when run locally, making it easy to develop without needing to set up DynamoDB:

```bash
cd api
go run cmd/api/main.go
```

If you want to test with DynamoDB locally, you can set the environment variables:

```bash
cd api
USE_DYNAMODB=true DYNAMODB_TABLE_NAME=local-messages go run cmd/api/main.go
```

Note: For local DynamoDB testing, you'll need to have AWS credentials configured with DynamoDB permissions.

### Frontend

```bash
cd frontend
npm install
npm start
```

## Infrastructure

The infrastructure is defined using CloudFormation templates in the `iac/` directory:

- `networking.yaml` - VPC, subnets, and other networking resources
- `ecr.yaml` - ECR repository for the API container image
- `ecs.yaml` - ECS cluster
- `ecs-task.yaml` - ECS task definition
- `ecs-service.yaml` - ECS service
- `alb.yaml` - Application Load Balancer
- `cloudfront.yaml` - CloudFront distribution for the frontend
- `logging.yaml` - CloudWatch logging resources
- `dynamodb.yaml` - DynamoDB table for message storage

## Distributed System Considerations

The API service is deployed with multiple instances for high availability. To ensure data consistency across instances, the application uses:

1. DynamoDB for persistent storage in AWS deployments (controlled by USE_DYNAMODB=true)
2. In-memory storage for local development (default behavior)
3. Appropriate IAM permissions for the ECS tasks to access DynamoDB
4. Configuration through environment variables to control storage behavior

This dual-storage approach provides:
- Simple local development without external dependencies
- Robust distributed storage for production deployments
- Flexibility to choose the appropriate storage mechanism based on the environment

The end-to-end tests include a small delay after creating messages to account for eventual consistency in the distributed system.
