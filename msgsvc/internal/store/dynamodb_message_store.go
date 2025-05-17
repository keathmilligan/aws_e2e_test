package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws_e2e_test/msgsvc/internal/model"
)

// DynamoDBMessageStore is a DynamoDB-based implementation of message store
type DynamoDBMessageStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBMessageStore creates a new DynamoDB-based message store
func NewDynamoDBMessageStore(tableName string) (*DynamoDBMessageStore, error) {
	log.Printf("Initializing DynamoDB message store with table name: %s", tableName)

	// Validate table name
	if tableName == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	// Load AWS configuration with explicit region
	// First try to get region from environment variable
	region := os.Getenv("AWS_REGION")
	if region == "" {
		// Default to us-east-1 if not specified
		region = "us-east-1"
		log.Printf("AWS_REGION not set, defaulting to %s", region)
	}

	// Load AWS configuration
	log.Printf("Loading AWS configuration for region: %s", region)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	log.Printf("Initialized DynamoDB client in region: %s", region)

	// Create the store
	store := &DynamoDBMessageStore{
		client:    client,
		tableName: tableName,
	}

	// Ensure the table exists
	err = store.ensureTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure table exists: %w", err)
	}

	return store, nil
}

// ensureTableExists creates the DynamoDB table if it doesn't exist
func (s *DynamoDBMessageStore) ensureTableExists() error {
	log.Printf("Checking if DynamoDB table %s exists...", s.tableName)

	// Check if table exists
	describeInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	}
	log.Printf("Describing table with input: %+v", describeInput)

	describeOutput, err := s.client.DescribeTable(context.TODO(), describeInput)

	// If table exists, return
	if err == nil {
		log.Printf("DynamoDB table %s already exists with status: %s",
			s.tableName, describeOutput.Table.TableStatus)

		// Log the table ARN to help with debugging IAM permissions
		log.Printf("DynamoDB table ARN: %s", *describeOutput.Table.TableArn)
		return nil
	}

	// Check if the error is because the table doesn't exist or something else
	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		log.Printf("ERROR: Failed to describe table %s: %v", s.tableName, err)
		return fmt.Errorf("failed to describe table: %w", err)
	}

	log.Printf("DynamoDB table %s does not exist, creating it now...", s.tableName)

	// Create table if it doesn't exist
	createInput := &dynamodb.CreateTableInput{
		TableName: aws.String(s.tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	log.Printf("Creating table with input: %+v", createInput)

	_, err = s.client.CreateTable(context.TODO(), createInput)

	if err != nil {
		log.Printf("Failed to create table %s: %v", s.tableName, err)
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Table %s created, waiting for it to become active...", s.tableName)

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(s.client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	}, 5*60)

	if err != nil {
		log.Printf("Failed to wait for table %s to be created: %v", s.tableName, err)
		return fmt.Errorf("failed to wait for table to be created: %w", err)
	}

	log.Printf("Successfully created DynamoDB table: %s", s.tableName)
	return nil
}

// GetAll returns all messages
func (s *DynamoDBMessageStore) GetAll() ([]*model.Message, error) {
	log.Printf("Getting all messages from DynamoDB table %s", s.tableName)

	// Scan the table to get all items
	scanInput := &dynamodb.ScanInput{
		TableName:      aws.String(s.tableName),
		ConsistentRead: aws.Bool(true), // Use strongly consistent reads
	}

	log.Printf("Scanning table with input: %+v", scanInput)
	result, err := s.client.Scan(context.TODO(), scanInput)

	if err != nil {
		log.Printf("Failed to scan table %s: %v", s.tableName, err)
		return []*model.Message{}, fmt.Errorf("failed to scan table: %w", err)
	}

	log.Printf("Scan returned %d items from table %s", len(result.Items), s.tableName)

	// Unmarshal items into messages
	messages := make([]*model.Message, 0, len(result.Items))
	for i, item := range result.Items {
		log.Printf("Processing item %d: %+v", i, item)
		var message model.Message
		err := attributevalue.UnmarshalMap(item, &message)
		if err != nil {
			log.Printf("Failed to unmarshal item %d: %v", i, err)
			continue
		}
		log.Printf("Successfully unmarshalled item to message: %+v", message)
		messages = append(messages, &message)
	}

	log.Printf("Returning %d messages from table %s", len(messages), s.tableName)
	return messages, nil
}

// Add adds a new message to the store
func (s *DynamoDBMessageStore) Add(message *model.Message) error {
	log.Printf("Adding message with ID %s to DynamoDB table %s", message.ID, s.tableName)

	// Double-check that the table exists before trying to write to it
	describeInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	}

	_, err := s.client.DescribeTable(context.TODO(), describeInput)
	if err != nil {
		log.Printf("ERROR: Table %s does not exist or cannot be accessed: %v", s.tableName, err)
		log.Printf("ERROR: Attempting to create the table before writing...")

		// Try to create the table
		err = s.ensureTableExists()
		if err != nil {
			log.Printf("ERROR: Failed to create table %s: %v", s.tableName, err)
			return fmt.Errorf("failed to ensure table exists before writing: %w", err)
		}
	}

	// Marshal message to DynamoDB item
	item, err := attributevalue.MarshalMap(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("Marshalled message to DynamoDB item: %+v", item)

	// Put item in table
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
		// Add a condition to ensure the item doesn't already exist (optional)
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	}
	log.Printf("Putting item in table %s with input: %+v", s.tableName, input)

	_, err = s.client.PutItem(context.TODO(), input)

	if err != nil {
		// Check if the error is because the condition failed (item already exists)
		var conditionFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailedErr) {
			log.Printf("WARNING: Item with ID %s already exists in table %s", message.ID, s.tableName)
			// This is not considered an error, as the item already exists
			return nil
		}

		log.Printf("ERROR: Failed to put item in table %s: %v", s.tableName, err)
		log.Printf("ERROR: Check IAM permissions for dynamodb:PutItem on table %s", s.tableName)
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	log.Printf("Successfully added message with ID %s to DynamoDB table %s", message.ID, s.tableName)

	// Verify the item was written by trying to get it back
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: message.ID},
		},
		ConsistentRead: aws.Bool(true),
	}

	log.Printf("Verifying item was written by getting it back...")
	getOutput, err := s.client.GetItem(context.TODO(), getInput)

	if err != nil {
		log.Printf("WARNING: Failed to verify item was written: %v", err)
		log.Printf("WARNING: Check IAM permissions for dynamodb:GetItem")
		// Don't return an error here, as the item might have been written successfully
	} else if getOutput.Item == nil || len(getOutput.Item) == 0 {
		log.Printf("WARNING: Item was not found after writing it")
		log.Printf("WARNING: This may indicate a problem with the DynamoDB table or permissions")
	} else {
		log.Printf("Successfully verified item was written to DynamoDB")
	}

	return nil
}
