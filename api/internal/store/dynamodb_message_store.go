package store

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/awse2e/backend/internal/model"
)

// DynamoDBMessageStore is a DynamoDB-based implementation of message store
type DynamoDBMessageStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBMessageStore creates a new DynamoDB-based message store
func NewDynamoDBMessageStore(tableName string) (*DynamoDBMessageStore, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

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
	// Check if table exists
	_, err := s.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})

	// If table exists, return
	if err == nil {
		return nil
	}

	// Create table if it doesn't exist
	_, err = s.client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
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
	})

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(s.client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	}, 5*60)

	if err != nil {
		return fmt.Errorf("failed to wait for table to be created: %w", err)
	}

	log.Printf("Created DynamoDB table: %s", s.tableName)
	return nil
}

// GetAll returns all messages
func (s *DynamoDBMessageStore) GetAll() []*model.Message {
	// Scan the table to get all items
	result, err := s.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})

	if err != nil {
		log.Printf("Failed to scan table: %v", err)
		return []*model.Message{}
	}

	// Unmarshal items into messages
	messages := make([]*model.Message, 0, len(result.Items))
	for _, item := range result.Items {
		var message model.Message
		err := attributevalue.UnmarshalMap(item, &message)
		if err != nil {
			log.Printf("Failed to unmarshal item: %v", err)
			continue
		}
		messages = append(messages, &message)
	}

	return messages
}

// Add adds a new message to the store
func (s *DynamoDBMessageStore) Add(message *model.Message) {
	// Marshal message to DynamoDB item
	item, err := attributevalue.MarshalMap(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	// Put item in table
	_, err = s.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		log.Printf("Failed to put item: %v", err)
	}
}
