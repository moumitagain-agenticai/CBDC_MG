curl -X POST http://localhost:8080/api/v1/cbdc/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "500.00",
    "currency": "INR",
    "source_wallet": "wallet-123",
    "destination_wallet": "wallet-456",
    "reference_id": "REF-002"
  }'