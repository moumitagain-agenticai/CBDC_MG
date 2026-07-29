# Fineract CBDC India Connector

Production-ready Go microservice that connects Apache Fineract with India's CBDC (Digital Rupee/e₹) rails via the RBI CBDC Pilot API.

## Features

- ✅ Complete CBDC operations: Issue, Transfer, Lock, Burn, Redeem
- ✅ Multi-currency support (INR, USD, etc.)
- ✅ Production-grade error handling with retry and circuit breaker
- ✅ OpenAPI/Swagger documentation
- ✅ Prometheus metrics
- ✅ Structured logging (JSON/Console)
- ✅ Health and readiness endpoints
- ✅ Rate limiting
- ✅ CORS support
- ✅ Containerized (Docker)
- ✅ Comprehensive test coverage

## Architecture

The service follows a hexagonal/ports-and-adapters architecture:

┌─────────────────────────────────────────────────────────────┐
│ API Layer (HTTP) │
│ (Handlers, Middleware) │
└─────────────────────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────────────────────┐
│ Service Layer │
│ (Connector, Transaction, Health Services) │
└─────────────────────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────────────────────┐
│ Domain Layer │
│ (Transaction, Wallet, Error Models) │
└─────────────────────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────────────────────┐
│ Ports (Interfaces) │
│ CBDCClient, FineractClient, Repository │
└─────────────────────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────────────────────┐
│ Adapters (Implementations) │
│ CBDC Client, Fineract Client, Postgres Repository │
└─────────────────────────────────────────────────────────────┘
