package database

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
)

type DynamoDBClient interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

type DynamoDBManager struct {
	tableName string
	client    DynamoDBClient
}

func NewDynamoDBManager(tableName string, client DynamoDBClient) *DynamoDBManager {
	return &DynamoDBManager{
		tableName: tableName,
		client:    client,
	}
}

func (d *DynamoDBManager) ProcessTransaction(transaction types.Transaction) (string, error) {
	if !IsValidTransaction(transaction) {
		return "REJECT", errors.New("transaction validation failed")
	}
	processingTime := 800 + rand.Intn(800)
	transaction.ExpiryTime = time.Now().Unix() + 300
	transaction.Timestamp = time.Now().Unix()

	item, err := dynamodbattribute.MarshalMap(transaction)
	if err != nil {
		return "REJECT", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(CardHash) OR ExpiryTime < :now"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":now": {N: aws.String(fmt.Sprintf("%d", time.Now().Unix()))},
		},
	}

	_, err = d.client.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return "REJECT", errors.New("transaction in progress")
		}
		return "REJECT", fmt.Errorf("error creating transaction: %w", err)
	}

	time.Sleep(time.Duration(processingTime) * time.Millisecond)

	err = d.DeleteTransaction(transaction.TransactionID)
	if err != nil {
		fmt.Printf("Error deleting transaction: %v\n", err)
	}

	return "ACCEPT", nil
}

func IsValidTransaction(transaction types.Transaction) bool {
	return transaction.TransactionID != "" && transaction.CardHash != ""
}

func (d *DynamoDBManager) GetTransactionByCardHash(cardHash string) (*types.Transaction, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("CardHashIndex"),
		KeyConditionExpression: aws.String("CardHash = :cardHash"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cardHash": {S: aws.String(cardHash)},
		},
	}

	result, err := d.client.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query by CardHash: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	transactionID := *result.Items[0]["TransactionID"].S

	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"TransactionID": {S: aws.String(transactionID)},
		},
	}

	getResult, err := d.client.GetItem(getInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if getResult.Item == nil {
		return nil, nil
	}

	var transaction types.Transaction
	err = dynamodbattribute.UnmarshalMap(getResult.Item, &transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &transaction, nil
}

func (d *DynamoDBManager) CreateTransaction(transaction types.Transaction) error {
	if transaction.ExpiryTime == 0 {
		transaction.ExpiryTime = time.Now().Unix() + 300
	}
	if transaction.Timestamp == 0 {
		transaction.Timestamp = time.Now().Unix()
	}

	item, err := dynamodbattribute.MarshalMap(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	_, err = d.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (d *DynamoDBManager) DeleteTransaction(transactionID string) error {
	_, err := d.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"TransactionID": {S: aws.String(transactionID)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}
