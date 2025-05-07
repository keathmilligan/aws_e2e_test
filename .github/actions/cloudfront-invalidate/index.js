const { CloudFrontClient, CreateInvalidationCommand } = require('@aws-sdk/client-cloudfront');

async function run() {
  try {
    // Get inputs from environment variables
    const distributionId = process.env.INPUT_DISTRIBUTION_ID;
    const pathsInput = process.env.INPUT_PATHS || '/*';
    let callerReference = process.env.INPUT_CALLER_REFERENCE;
    const awsRegion = process.env.INPUT_AWS_REGION || 'us-east-1';

    if (!distributionId) {
      console.error('Error: distribution-id is required');
      process.exit(1);
    }

    // If no caller reference is provided, generate one based on the current timestamp
    if (!callerReference) {
      callerReference = `github-action-${Date.now()}`;
    }

    // Parse paths (comma-separated string to array)
    const paths = pathsInput.split(',').map(path => path.trim());

    // Create CloudFront client
    const client = new CloudFrontClient({ region: awsRegion });

    // Log the invalidation details
    console.log(`Creating CloudFront invalidation for distribution: ${distributionId}`);
    console.log(`Paths to invalidate: ${paths.join(', ')}`);
    console.log(`Caller reference: ${callerReference}`);

    // Create invalidation command
    const command = new CreateInvalidationCommand({
      DistributionId: distributionId,
      InvalidationBatch: {
        Paths: {
          Quantity: paths.length,
          Items: paths
        },
        CallerReference: callerReference
      }
    });

    // Execute the command
    const response = await client.send(command);

    // Output the invalidation ID
    const invalidationId = response.Invalidation.Id;
    console.log(`Invalidation created successfully. Invalidation ID: ${invalidationId}`);
    
    // Set output for GitHub Actions
    console.log(`::set-output name=invalidation-id::${invalidationId}`);

  } catch (error) {
    console.error(`Action failed with error: ${error.message}`);
    if (error.stack) {
      console.error(error.stack);
    }
    process.exit(1);
  }
}

run();