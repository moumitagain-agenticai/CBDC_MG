curl -X POST http://localhost:8080/api/v1/cbdc/issue \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "1000.00",
    "currency": "INR",
    "source_wallet": "wallet-123",
    "reference_id": "REF-001"
  }'