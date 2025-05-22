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
	"github.com/aws_e2e_test/usersvc/internal/model"
)

// DynamoDBUserStore is a DynamoDB-based implementation of user store
type DynamoDBUserStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBUserStore creates a new DynamoDB-based user store
func NewDynamoDBUserStore(tableName string) (*DynamoDBUserStore, error) {
	log.Printf("Initializing DynamoDB user store with table name: %s", tableName)

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
	store := &DynamoDBUserStore{
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
func (s *DynamoDBUserStore) ensureTableExists() error {
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
				AttributeName: aws.String("Email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("Email"),
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

// GetByEmail retrieves a user by email
func (s *DynamoDBUserStore) GetByEmail(email string) (*model.User, error) {
	log.Printf("Getting user with email %s from DynamoDB table %s", email, s.tableName)

	// Get item from DynamoDB
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"Email": &types.AttributeValueMemberS{Value: email},
		},
		ConsistentRead: aws.Bool(true),
	}

	log.Printf("Getting item with input: %+v", getInput)
	result, err := s.client.GetItem(context.TODO(), getInput)

	if err != nil {
		log.Printf("Failed to get item from table %s: %v", s.tableName, err)
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	// Check if item exists
	if result.Item == nil || len(result.Item) == 0 {
		log.Printf("User with email %s not found in table %s", email, s.tableName)
		return nil, nil
	}

	// Unmarshal item into user
	var user model.User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Printf("Failed to unmarshal item: %v", err)
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	log.Printf("Successfully retrieved user with email %s from table %s", email, s.tableName)
	return &user, nil
}

// GetAll retrieves all users
func (s *DynamoDBUserStore) GetAll() ([]*model.User, error) {
	log.Printf("Getting all users from DynamoDB table %s", s.tableName)

	// Scan the table to get all items
	scanInput := &dynamodb.ScanInput{
		TableName:      aws.String(s.tableName),
		ConsistentRead: aws.Bool(true), // Use strongly consistent reads
	}

	log.Printf("Scanning table with input: %+v", scanInput)
	result, err := s.client.Scan(context.TODO(), scanInput)

	if err != nil {
		log.Printf("Failed to scan table %s: %v", s.tableName, err)
		return []*model.User{}, fmt.Errorf("failed to scan table: %w", err)
	}

	log.Printf("Scan returned %d items from table %s", len(result.Items), s.tableName)

	// Unmarshal items into users
	users := make([]*model.User, 0, len(result.Items))
	for i, item := range result.Items {
		log.Printf("Processing item %d: %+v", i, item)
		var user model.User
		err := attributevalue.UnmarshalMap(item, &user)
		if err != nil {
			log.Printf("Failed to unmarshal item %d: %v", i, err)
			continue
		}
		log.Printf("Successfully unmarshalled item to user: %+v", user)
		users = append(users, &user)
	}

	log.Printf("Returning %d users from table %s", len(users), s.tableName)
	return users, nil
}

// Create creates a new user
func (s *DynamoDBUserStore) Create(user *model.User) error {
	log.Printf("Creating user with email %s in DynamoDB table %s", user.Email, s.tableName)

	// Marshal user to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Printf("Failed to marshal user: %v", err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	log.Printf("Marshalled user to DynamoDB item: %+v", item)

	// Put item in table
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
		// Add a condition to ensure the item doesn't already exist
		ConditionExpression: aws.String("attribute_not_exists(Email)"),
	}
	log.Printf("Putting item in table %s with input: %+v", s.tableName, input)

	_, err = s.client.PutItem(context.TODO(), input)

	if err != nil {
		// Check if the error is because the condition failed (item already exists)
		var conditionFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailedErr) {
			log.Printf("User with email %s already exists in table %s", user.Email, s.tableName)
			return fmt.Errorf("user with email %s already exists", user.Email)
		}

		log.Printf("ERROR: Failed to put item in table %s: %v", s.tableName, err)
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	log.Printf("Successfully created user with email %s in DynamoDB table %s", user.Email, s.tableName)
	return nil
}

// Update updates an existing user
func (s *DynamoDBUserStore) Update(user *model.User) error {
	log.Printf("Updating user with email %s in DynamoDB table %s", user.Email, s.tableName)

	// Marshal user to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Printf("Failed to marshal user: %v", err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	log.Printf("Marshalled user to DynamoDB item: %+v", item)

	// Put item in table
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
		// Add a condition to ensure the item already exists
		ConditionExpression: aws.String("attribute_exists(Email)"),
	}
	log.Printf("Putting item in table %s with input: %+v", s.tableName, input)

	_, err = s.client.PutItem(context.TODO(), input)

	if err != nil {
		// Check if the error is because the condition failed (item doesn't exist)
		var conditionFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailedErr) {
			log.Printf("User with email %s does not exist in table %s", user.Email, s.tableName)
			return fmt.Errorf("user with email %s does not exist", user.Email)
		}

		log.Printf("ERROR: Failed to put item in table %s: %v", s.tableName, err)
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	log.Printf("Successfully updated user with email %s in DynamoDB table %s", user.Email, s.tableName)
	return nil
}

// Delete deletes a user by email
func (s *DynamoDBUserStore) Delete(email string) error {
	log.Printf("Deleting user with email %s from DynamoDB table %s", email, s.tableName)

	// Delete item from table
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"Email": &types.AttributeValueMemberS{Value: email},
		},
	}
	log.Printf("Deleting item from table %s with input: %+v", s.tableName, input)

	_, err := s.client.DeleteItem(context.TODO(), input)

	if err != nil {
		log.Printf("ERROR: Failed to delete item from table %s: %v", s.tableName, err)
		return fmt.Errorf("failed to delete item from DynamoDB: %w", err)
	}

	log.Printf("Successfully deleted user with email %s from DynamoDB table %s", email, s.tableName)
	return nil
}
