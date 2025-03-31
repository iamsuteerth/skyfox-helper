aws dynamodb create-table \
    --table-name PendingTransactions \
    --attribute-definitions AttributeName=TransactionID,AttributeType=S \
                            AttributeName=CardHash,AttributeType=S \
    --key-schema AttributeName=TransactionID,KeyType=HASH \
    --global-secondary-indexes "[
        {
            \"IndexName\": \"CardHashIndex\",
            \"KeySchema\": [{\"AttributeName\":\"CardHash\",\"KeyType\":\"HASH\"}],
            \"Projection\": {\"ProjectionType\":\"KEYS_ONLY\"}
        }
    ]" \
    --billing-mode PAY_PER_REQUEST \
    --region ap-south-1

aws dynamodb update-time-to-live \
    --table-name PendingTransactions \
    --time-to-live-specification "Enabled=true, AttributeName=ExpiryTime" \
    --region ap-south-1