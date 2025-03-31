package database_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/database"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestProcessTransaction(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	manager := database.NewDynamoDBManager("PendingTransactions", mockClient)

	validTx := types.Transaction{
		TransactionID: "tx123",
		CardHash:      "card123",
		Timestamp:     time.Now().Unix(),
		ExpiryTime:    time.Now().Unix() + 300,
	}

	expiredTx := types.Transaction{
		TransactionID: "tx456",
		CardHash:      "card456",
		Timestamp:     time.Now().Unix() - 600,
		ExpiryTime:    time.Now().Unix() - 300,
	}

	validExistingTx := types.Transaction{
		TransactionID: "tx789",
		CardHash:      "card789",
		Timestamp:     time.Now().Unix() - 60,
		ExpiryTime:    time.Now().Unix() + 240,
	}

	invalidTx := types.Transaction{
		TransactionID: "",
		CardHash:      "card999",
	}

	t.Run("Behavior4_ValidationFails", func(t *testing.T) {

		result, err := manager.ProcessTransaction(invalidTx)

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction validation failed")

		mockClient.AssertNotCalled(t, "Query")
		mockClient.AssertNotCalled(t, "GetItem")
		mockClient.AssertNotCalled(t, "PutItem")
		mockClient.AssertNotCalled(t, "DeleteItem")
	})

	t.Run("Behavior1_HashExists_TTLExpired", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(expiredTx.TransactionID)},
					"CardHash":      {S: aws.String(expiredTx.CardHash)},
				},
			},
		}

		expiredItem, _ := dynamodbattribute.MarshalMap(expiredTx)
		getItemOutput := &dynamodb.GetItemOutput{
			Item: expiredItem,
		}

		deleteItemOutput := &dynamodb.DeleteItemOutput{}

		mockClient.On("Query", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.IndexName) == "CardHashIndex" &&
				aws.StringValue(input.KeyConditionExpression) == "CardHash = :cardHash"
		})).Return(queryOutput, nil).Once()

		mockClient.On("GetItem", mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == expiredTx.TransactionID
		})).Return(getItemOutput, nil).Once()

		mockClient.On("DeleteItem", mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == expiredTx.TransactionID
		})).Return(deleteItemOutput, nil).Once()

		result, err := manager.ProcessTransaction(types.Transaction{CardHash: expiredTx.CardHash, TransactionID: "newTx123"})

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction expired")
		mockClient.AssertExpectations(t)
	})

	t.Run("Behavior2_HashExists_TTLValid", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(validExistingTx.TransactionID)},
					"CardHash":      {S: aws.String(validExistingTx.CardHash)},
				},
			},
		}

		validItem, _ := dynamodbattribute.MarshalMap(validExistingTx)
		getItemOutput := &dynamodb.GetItemOutput{
			Item: validItem,
		}

		deleteItemOutput := &dynamodb.DeleteItemOutput{}

		mockClient.On("Query", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.IndexName) == "CardHashIndex"
		})).Return(queryOutput, nil).Once()

		mockClient.On("GetItem", mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == validExistingTx.TransactionID
		})).Return(getItemOutput, nil).Once()

		mockClient.On("DeleteItem", mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == validExistingTx.TransactionID
		})).Return(deleteItemOutput, nil).Once()

		result, err := manager.ProcessTransaction(types.Transaction{CardHash: validExistingTx.CardHash, TransactionID: "newTx456"})

		assert.Equal(t, "ACCEPT", result)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Behavior3_HashDoesNotExist", func(t *testing.T) {
		emptyQueryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}

		putItemOutput := &dynamodb.PutItemOutput{}

		mockClient.On("Query", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.IndexName) == "CardHashIndex"
		})).Return(emptyQueryOutput, nil).Once()

		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions"
		})).Return(putItemOutput, nil).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "ACCEPT", result)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_QueryFails", func(t *testing.T) {
		mockClient.On("Query", mock.Anything).Return(&dynamodb.QueryOutput{}, errors.New("query error")).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error checking existing transactions")
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_DeleteExpiredFails", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(expiredTx.TransactionID)},
					"CardHash":      {S: aws.String(expiredTx.CardHash)},
				},
			},
		}

		expiredItem, _ := dynamodbattribute.MarshalMap(expiredTx)
		getItemOutput := &dynamodb.GetItemOutput{
			Item: expiredItem,
		}

		mockClient.On("Query", mock.Anything).Return(queryOutput, nil).Once()
		mockClient.On("GetItem", mock.Anything).Return(getItemOutput, nil).Once()
		mockClient.On("DeleteItem", mock.Anything).Return(&dynamodb.DeleteItemOutput{}, errors.New("delete error")).Once()

		result, err := manager.ProcessTransaction(types.Transaction{CardHash: expiredTx.CardHash, TransactionID: "newTx123"})

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error deleting expired transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_DeleteValidFails", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(validExistingTx.TransactionID)},
					"CardHash":      {S: aws.String(validExistingTx.CardHash)},
				},
			},
		}

		validItem, _ := dynamodbattribute.MarshalMap(validExistingTx)
		getItemOutput := &dynamodb.GetItemOutput{
			Item: validItem,
		}

		mockClient.On("Query", mock.Anything).Return(queryOutput, nil).Once()
		mockClient.On("GetItem", mock.Anything).Return(getItemOutput, nil).Once()
		mockClient.On("DeleteItem", mock.Anything).Return(&dynamodb.DeleteItemOutput{}, errors.New("delete error")).Once()

		result, err := manager.ProcessTransaction(types.Transaction{CardHash: validExistingTx.CardHash, TransactionID: "newTx456"})

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error deleting transaction during accept")
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_CreateFails", func(t *testing.T) {
		emptyQueryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}

		mockClient.On("Query", mock.Anything).Return(emptyQueryOutput, nil).Once()
		mockClient.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("put item error")).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_GetItemFails", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(validExistingTx.TransactionID)},
					"CardHash":      {S: aws.String(validExistingTx.CardHash)},
				},
			},
		}

		mockClient.On("Query", mock.Anything).Return(queryOutput, nil).Once()
		mockClient.On("GetItem", mock.Anything).Return(&dynamodb.GetItemOutput{}, errors.New("get item error")).Once()

		result, err := manager.ProcessTransaction(types.Transaction{CardHash: validExistingTx.CardHash, TransactionID: "newTx456"})

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error checking existing transactions")
		mockClient.AssertExpectations(t)
	})
}
