## Key Considerations
### **PCI-DSS Compliance**:
- Never store raw card numbers/CVVs (only hashes)
- Expiry dates never stored (validation only)
- No permanent transaction history

## **Validation Matrix**:
| **Field**      | **Rules**                                                                                           |
|---------------|---------------------------------------------------------------------------------------------------|
| **Card Number** | - Must be a **Luhn-valid** 16-digit number (handles American Express variations).  <br> - **Test card numbers** from known BIN ranges must be rejected. |
| **CVV**       | - Must be a **3-digit numeric** value (leading zeros allowed). <br> - Cannot be all zeros (`000`) or sequential patterns (e.g., `123`). |
| **Expiry**    | - Must be a **future date** (month precision). <br> - Strict **ISO 8601 format** (`YYYY-MM`). <br> - Expiry year must be within **current year ≤ expiry year ≤ current year + 20**. |
| **Name**      | - Only **ASCII printable characters** allowed. <br> - Length must be between **2-40 characters**. <br> - No consecutive spaces. |

### **Repository Structre**:
```
/
├── architecture
│   ├── network-diagram.puml         # PlantUML architecture diagrams
│   └── decision-log.md              # Design rationale
├── internal
│   ├── validators                   # Validation subsystems
│   │   ├── card.go                  # Luhn + BIN checks
│   │   ├── crypto.go                # Secure hashing (HMAC-SHA256)
│   │   └── temporal.go              # Time validation logic
│   ├── transaction                  # Core business logic
│   │   ├── manager.go               # Transaction lifecycle
│   │   └── concurrency.go           # Goroutine coordination
│   └── store                        # Database abstraction
│       ├── dynamo.go                # DAO pattern implementation
│       └── lock.go                  # Distributed lock interface
├── test
│   ├── fixtures                     # Test data sets
│   │   ├── valid_cards.json
│   │   └── invalid_transactions.csv
│   ├── unit                         # Test packages
│   │   ├── validators_test.go
│   │   └── transaction_test.go
│   └── load                         # Load testing scenarios
├── go.mod                           # Go modules
└── cloudformation                   # Infrastructure definitions
    ├── payment-table.yaml           # DynamoDB with TTL config
    └── iam-roles.yaml               # Minimal permissions
```