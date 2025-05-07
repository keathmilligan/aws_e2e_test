# CloudFront Invalidate Action

This GitHub Action creates CloudFront invalidations using the AWS SDK for JavaScript.

## Features

- Creates CloudFront invalidations for specified paths
- Uses AWS SDK for JavaScript v3
- Supports custom caller references
- Outputs the invalidation ID for reference in subsequent steps
- Simple implementation that installs dependencies at runtime

## Usage

```yaml
- name: Invalidate CloudFront Cache
  id: invalidate
  uses: ./.github/actions/cloudfront-invalidate
  with:
    distribution-id: ${{ inputs.CloudFrontDistributionId }}
    paths: '/*'
    aws-region: 'us-east-1'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `distribution-id` | CloudFront distribution ID | Yes | N/A |
| `paths` | Paths to invalidate (comma-separated) | No | `/*` |
| `caller-reference` | A unique identifier for the invalidation batch | No | Auto-generated timestamp |
| `aws-region` | AWS region | No | `us-east-1` |

## Outputs

| Output | Description |
|--------|-------------|
| `invalidation-id` | The ID of the created invalidation |

## How It Works

This action:
1. Installs the required npm dependencies at runtime
2. Executes the JavaScript code to create a CloudFront invalidation
3. Returns the invalidation ID as an output

## Prerequisites

This action requires AWS credentials to be configured. You can use the [configure-aws-credentials](https://github.com/aws-actions/configure-aws-credentials) action to set up the credentials:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v3
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
    aws-region: us-east-1
```

## Example Workflow

```yaml
name: Deploy and Invalidate Cache

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v3
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      # Deploy your application...

      - name: Invalidate CloudFront Cache
        id: invalidate
        uses: ./.github/actions/cloudfront-invalidate
        with:
          distribution-id: ${{ secrets.CLOUDFRONT_DISTRIBUTION_ID }}
          paths: '/index.html,/assets/*'
          
      - name: Print invalidation ID
        run: echo "Invalidation ID: ${{ steps.invalidate.outputs.invalidation-id }}"
```

## License

MIT