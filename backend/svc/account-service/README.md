# Account Service

A production-ready Go microservice for managing user accounts in a distributed banking system.

## Features

- **Account Management**: Create and retrieve user accounts
- **Bank Integration**: Support for multiple banks
- **Payment Validation**: Asynchronous payment request validation
- **Kafka Integration**: Event-driven architecture with Kafka messaging
- **Graceful Shutdown**: Proper resource cleanup and shutdown handling
- **Health Checks**: Built-in health monitoring endpoints
- **Comprehensive Testing**: Unit tests with mocks and benchmarks

## Architecture

### Components

- **Service**: Main business logic layer
- **HTTPServer**: HTTP server with middleware and routing
- **AppEnv**: Application environment with dependencies
- **Config**: Configuration management
- **Validators**: Payment validation logic

### Key Improvements

1. **Separation of Concerns**: Clear separation between HTTP handling, business logic, and infrastructure
2. **Proper Error Handling**: Comprehensive error handling with structured logging
3. **Configuration Management**: Centralized configuration with validation
4. **Graceful Shutdown**: Proper resource cleanup and signal handling
5. **Testability**: Dependency injection and mock interfaces for testing
6. **Idiomatic Go**: Following Go best practices and conventions

## API Endpoints

### Health Check
```
GET /health
```
Returns service health status.

### Get Banks
```
GET /banks
```
Returns list of available banks. Requires authentication.

### Get User Accounts
```
GET /myaccounts
```
Returns accounts for the authenticated user.

### Create Account
```
POST /new
Content-Type: application/json

{
  "name": "My Account",
  "sourceFundsAccountId": 123,  // Optional, for additional accounts
  "initialBalance": 1000        // Optional, for additional accounts
}
```
Creates a new account for the authenticated user.

## Configuration

The service is configured via environment variables:

- `SERVE_PORT`: HTTP server port (default: 8080)
- `KAFKA_BROKER`: Kafka broker address (required)
- `DATABASE_URL`: Database connection string (handled by common package)

## Running the Service

### Prerequisites

- Go 1.21+
- Kafka cluster
- PostgreSQL database

### Local Development

```bash
# Set environment variables
export KAFKA_BROKER=localhost:9092
export SERVE_PORT=8080

# Run the service
go run .
```

### Docker

```bash
# Build image
docker build -t account-service .

# Run container
docker run -p 8080:8080 \
  -e KAFKA_BROKER=kafka:9092 \
  account-service
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Benchmarks
```bash
go test -bench=.
```

### Coverage
```bash
go test -cover ./...
```

## Payment Validation Flow

1. **Payment Request**: Received via Kafka from payment service
2. **Concurrent Validation**: 
   - Balance check: Verify source account has sufficient funds
   - Target account check: Verify target account exists
3. **Result Processing**: 
   - Success: Initiate transaction via Kafka
   - Failure: Send failure notification via Kafka
4. **Timeout Handling**: Automatic timeout after 4.5 seconds

## Error Handling

The service implements comprehensive error handling:

- **HTTP Errors**: Proper HTTP status codes with JSON error responses
- **Validation Errors**: Input validation with descriptive error messages
- **Database Errors**: Graceful handling of database connection issues
- **Kafka Errors**: Retry logic and error logging for message processing
- **Context Cancellation**: Proper handling of request cancellation

## Logging

Structured logging is implemented throughout the service:

- Request/response logging
- Error logging with context
- Performance metrics
- Business event logging

## Security Considerations

- **Authentication**: User ID validation via middleware
- **Authorization**: User-specific resource access
- **Input Validation**: Comprehensive request validation
- **Error Information**: Sanitized error responses

## Performance Optimizations

- **Connection Pooling**: Reused HTTP clients and database connections
- **Concurrent Processing**: Parallel validation checks
- **Efficient Serialization**: Optimized Kafka message serialization
- **Resource Management**: Proper cleanup and resource limits

## Monitoring and Observability

- Health check endpoint for load balancer integration
- Structured logging for log aggregation
- Error metrics and alerting hooks
- Request tracing support

## Future Enhancements

- [ ] Metrics collection (Prometheus)
- [ ] Distributed tracing (Jaeger/Zipkin)
- [ ] Circuit breaker pattern
- [ ] Rate limiting
- [ ] API versioning
- [ ] Database migrations
- [ ] TLS/HTTPS support
- [ ] JWT authentication