package database_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

	invalidTx := types.Transaction{
		TransactionID: "",
		CardHash:      "card999",
	}

	t.Run("Behavior4_ValidationFails", func(t *testing.T) {
		result, err := manager.ProcessTransaction(invalidTx)

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction validation failed")

		mockClient.AssertNotCalled(t, "PutItem")
		mockClient.AssertNotCalled(t, "DeleteItem")
	})

	t.Run("Behavior1_HashExists_TTLExpired", func(t *testing.T) {
		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.ConditionExpression) == "attribute_not_exists(CardHash) OR ExpiryTime < :now"
		})).Return(&dynamodb.PutItemOutput{}, awserr.New(
			dynamodb.ErrCodeConditionalCheckFailedException,
			"conditional check failed",
			errors.New("conditional check failed"),
		)).Once()

		result, err := manager.ProcessTransaction(types.Transaction{
			CardHash:      "card456",
			TransactionID: "newTx123",
		})

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction in progress")
		mockClient.AssertExpectations(t)
	})

	t.Run("Behavior2_HashExists_TTLValid", func(t *testing.T) {
		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.ConditionExpression) == "attribute_not_exists(CardHash) OR ExpiryTime < :now"
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		mockClient.On("DeleteItem", mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == "newTx456"
		})).Return(&dynamodb.DeleteItemOutput{}, nil).Once()

		result, err := manager.ProcessTransaction(types.Transaction{
			CardHash:      "card789",
			TransactionID: "newTx456",
		})

		assert.Equal(t, "ACCEPT", result)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Behavior3_HashDoesNotExist", func(t *testing.T) {
		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.ConditionExpression) == "attribute_not_exists(CardHash) OR ExpiryTime < :now"
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		mockClient.On("DeleteItem", mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == "tx123"
		})).Return(&dynamodb.DeleteItemOutput{}, nil).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "ACCEPT", result)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_PutItemFails", func(t *testing.T) {
		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.ConditionExpression) == "attribute_not_exists(CardHash) OR ExpiryTime < :now"
		})).Return(&dynamodb.PutItemOutput{}, errors.New("put item error")).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "REJECT", result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Error_DeleteFailsAfterAccept", func(t *testing.T) {
		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.ConditionExpression) == "attribute_not_exists(CardHash) OR ExpiryTime < :now"
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		mockClient.On("DeleteItem", mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == "tx123"
		})).Return(&dynamodb.DeleteItemOutput{}, errors.New("delete error")).Once()

		result, err := manager.ProcessTransaction(validTx)

		assert.Equal(t, "ACCEPT", result)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_GetTransactionByCardHash_Success", func(t *testing.T) {
		testTx := types.Transaction{
			TransactionID: "test123",
			CardHash:      "card123",
			Timestamp:     time.Now().Unix(),
			ExpiryTime:    time.Now().Unix() + 300,
		}

		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String(testTx.TransactionID)},
					"CardHash":      {S: aws.String(testTx.CardHash)},
				},
			},
		}

		item, _ := dynamodbattribute.MarshalMap(testTx)
		getItemOutput := &dynamodb.GetItemOutput{
			Item: item,
		}

		mockClient.On("Query", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.IndexName) == "CardHashIndex" &&
				aws.StringValue(input.KeyConditionExpression) == "CardHash = :cardHash"
		})).Return(queryOutput, nil).Once()

		mockClient.On("GetItem", mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions" &&
				aws.StringValue(input.Key["TransactionID"].S) == testTx.TransactionID
		})).Return(getItemOutput, nil).Once()

		result, err := manager.GetTransactionByCardHash(testTx.CardHash)

		assert.NoError(t, err)
		assert.Equal(t, testTx.TransactionID, result.TransactionID)
		assert.Equal(t, testTx.CardHash, result.CardHash)
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_GetTransactionByCardHash_NotFound", func(t *testing.T) {
		mockClient.On("Query", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}, nil).Once()

		result, err := manager.GetTransactionByCardHash("nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_GetTransactionByCardHash_QueryError", func(t *testing.T) {
		mockClient.On("Query", mock.Anything).Return(&dynamodb.QueryOutput{}, errors.New("query error")).Once()

		result, err := manager.GetTransactionByCardHash("card123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to query by CardHash")
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_GetTransactionByCardHash_GetItemError", func(t *testing.T) {
		queryOutput := &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"TransactionID": {S: aws.String("test123")},
					"CardHash":      {S: aws.String("card123")},
				},
			},
		}

		mockClient.On("Query", mock.Anything).Return(queryOutput, nil).Once()
		mockClient.On("GetItem", mock.Anything).Return(&dynamodb.GetItemOutput{}, errors.New("get item error")).Once()

		result, err := manager.GetTransactionByCardHash("card123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_createTransaction_Success", func(t *testing.T) {
		testTx := types.Transaction{
			TransactionID: "test123",
			CardHash:      "card123",
		}

		mockClient.On("PutItem", mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			return aws.StringValue(input.TableName) == "PendingTransactions"
		})).Return(&dynamodb.PutItemOutput{}, nil).Once()

		err := manager.CreateTransaction(testTx)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Test_createTransaction_PutItemError", func(t *testing.T) {
		testTx := types.Transaction{
			TransactionID: "test123",
			CardHash:      "card123",
		}

		mockClient.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("put item error")).Once()

		err := manager.CreateTransaction(testTx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create transaction")
		mockClient.AssertExpectations(t)
	})

}
