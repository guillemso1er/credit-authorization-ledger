package idempotency

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoDBStore struct {
	client    *dynamodb.DynamoDB
	tableName string
}

type idempotencyItem struct {
	Key        string `json:"key"`
	Response   string `json:"response"`
	TimeToLive int64  `json:"ttl"`
}

func NewDynamoDBStore(client *dynamodb.DynamoDB, tableName string) *DynamoDBStore {
	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}
}

func (s *DynamoDBStore) Get(ctx context.Context, key string) (string, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				S: aws.String(key),
			},
		},
	}

	result, err := s.client.GetItemWithContext(ctx, input)
	if err != nil {
		return "", err
	}

	if result.Item == nil {
		return "", nil // Not found
	}

	var item idempotencyItem
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return "", err
	}

	return item.Response, nil
}

func (s *DynamoDBStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	item := idempotencyItem{
		Key:        key,
		Response:   value,
		TimeToLive: time.Now().Add(ttl).Unix(),
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	}

	_, err = s.client.PutItemWithContext(ctx, input)
	return err
}