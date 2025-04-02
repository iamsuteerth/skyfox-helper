# Payment Service

A containerized Golang application that processes secure payment transactions via a RESTful API.

## Overview

Payment Service is part of the Skyfox Helper repository. It provides a robust payment processing gateway with card validation, transaction processing, and secure API authentication. The service is designed using Test-Driven Development (TDD) principles to ensure reliability and correctness.

## Features

- RESTful API for processing payment transactions
- Comprehensive card validation (Luhn check, expiry validation)
- Simulated payment processing with realistic success/failure rates
- API key authentication for secure access
- Request and transaction tracking with unique IDs
- Structured JSON responses
- Health check endpoint
- Containerized for easy deployment

## API Endpoints

### Health Check
```
GET /health
```
Returns the health status of the service.

### Process Payment
```
POST /payment
```
Processes a payment transaction.

#### Request Format
```json
{
    "card_number": "4242424242424242",
    "cvv": "123",
    "expiry": "12/25",
    "name": "John Doe",
    "amount": 24.23,
    "timestamp":"-"
}
```

## API Responses

### Successful Transaction (200 OK)
```json
{
    "message": "Transaction processed successfully",
    "request_id": "78e0351f-e5e3-4fdc-ad01-9def80d4ddf1",
    "status": "SUCCESS",
    "transaction_id": "2cbcc1af-567c-496c-8b3c-34dcbe660ae4"
}
```

### Validation Errors (422 Unprocessable Entity)
```json
{
    "errors": [
        {
            "field": "card_number",
            "message": "Card number failed Luhn check"
        },
        {
            "field": "name",
            "message": "Name must contain only letters, spaces, apostrophes, and hyphens"
        }
    ],
    "request_id": "be72964f-22aa-4863-89d3-3db5e08e4e2f",
    "status": "REJECT"
}
```

### Authentication Errors (403 Forbidden)

#### Missing API Key
```json
{
    "message": "API key is required",
    "status": "FORBIDDEN"
}
```

#### Invalid API Key
```json
{
    "message": "Invalid API key",
    "status": "FORBIDDEN"
}
```

### Route Not Found (404 Not Found)
```json
{
    "status": "NOT_FOUND",
    "error": "There is nothing to do here! 404!"
}
```

## Validation Rules

The payment service implements strict validation rules:

- **Card Number**: Must be 16 digits and pass the Luhn algorithm check
- **CVV**: Must be exactly 3 digits between 001-999
- **Expiry**: Must be in MM/YY format with valid month (01-12) and not expired
- **Name**: Must be 2-40 characters, containing only letters, spaces, apostrophes, and hyphens
- **Amount**: Must be a positive numeric value

## Configuration

The service can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Port on which the server listens | 8082 |
| API_KEY | API key for authentication (if empty, authentication is disabled) | "" |
| LOG_LEVEL | Logging level (debug/info) | info |
| APP_VERSION | Application version for health check | "dev" |

## Project Structure

```
.
├── Dockerfile                # Container configuration
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── main.go                   # Application entry point
├── processor
│   └── processor.go          # Payment processing logic
├── types
│   └── types.go              # Data models and types
└── validator
    └── validator.go          # Input validation logic
```

## Running Locally

1. Clone the repository
2. Navigate to the payment_service directory
3. Run with Go:
   ```
   go run main.go
   ```

## Running with Docker

```bash
# Build the image
docker build -t payment-service .

# Run the container
docker run -p 8082:8082 -e API_KEY=your_api_key payment-service
```

## Authentication

Secure your API by setting the `API_KEY` environment variable. Clients must include this key in the `x-api-key` header when making requests.

## Development

This service was built using Test-Driven Development (TDD) principles, ensuring high code quality and reliability.

