# Configure AWS Credentials Action

This reusable GitHub Action configures AWS credentials for both GitHub Actions (using OIDC) and local development (using `act`).

## Features

- Automatically detects whether it's running in GitHub Actions or locally with `act`
- Uses OIDC role-based authentication for GitHub Actions
- Uses access key-based authentication for local development with `act`
- Simplifies workflow files by consolidating credential configuration

## Usage

```yaml
- name: Configure AWS credentials
  uses: ./.github/actions/configure-aws-credentials
  with:
    # Required for GitHub Actions OIDC authentication
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
    
    # Required for both authentication methods
    aws-region: ${{ env.AWS_REGION }}
    
    # Required only for local development with act
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-session-token: ${{ secrets.AWS_SESSION_TOKEN }}
```

## Inputs

| Input | Description | Required |
|-------|-------------|----------|
| `role-to-assume` | AWS IAM role to assume when running in GitHub Actions | No |
| `aws-region` | AWS region to use | Yes |
| `aws-access-key-id` | AWS access key ID for local development with act | No |
| `aws-secret-access-key` | AWS secret access key for local development with act | No |
| `aws-session-token` | AWS session token for local development with act | No |

## How It Works

The action uses the `env.ACT` environment variable to detect whether it's running locally with `act`:

- If `env.ACT` is not set (running in GitHub Actions), it uses OIDC authentication
- If `env.ACT` is set (running locally), it uses access key-based authentication

This allows the same workflow to work seamlessly in both environments.