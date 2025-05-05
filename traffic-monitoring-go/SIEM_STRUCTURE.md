# V2X SIEM Project Structure

```
├── cmd/
│   └── api/
│       └── main.go           # Application entry point, wire DI
│
├── internal/
│   ├── domain/               # Pure domain entities without framework dependencies
│   │   ├── rule.go           # Rule entity with business logic
│   │   ├── alert.go
│   │   ├── event.go
│   │   └── enums.go          # Common enums like Severity, Status, etc.
│   │
│   ├── dto/                  # Data Transfer Objects
│   │   ├── common.go         # Shared DTO structures (pagination, envelopes)
│   │   ├── rule.go           # Rule request/response DTOs
│   │   ├── alert.go
│   │   └── event.go
│   │
│   ├── repository/           # Data access layer
│   │   ├── interfaces.go     # Repository interfaces
│   │   ├── rule.go           # Rule repository implementation (PostgreSQL)
│   │   ├── alert.go
│   │   ├── event.go
│   │   └── elasticsearch/    # ES-specific repository implementations
│   │
│   ├── service/              # Business logic layer
│   │   ├── interfaces.go     # Service interfaces
│   │   ├── rule.go           # Rule service implementation
│   │   ├── alert.go
│   │   ├── event.go
│   │   └── decorators/       # Service decorators (ES, logging, etc.)
│   │
│   ├── api/                  # API layer
│   │   ├── middleware/       # Middleware components
│   │   │   ├── auth.go
│   │   │   ├── correlation.go
│   │   │   ├── logging.go
│   │   │   ├── errors.go
│   │   │   └── recovery.go
│   │   │
│   │   ├── handlers/         # HTTP handlers
│   │   │   ├── rule.go
│   │   │   ├── alert.go
│   │   │   └── event.go
│   │   │
│   │   └── router.go         # Router setup
│   │
│   └── pkg/                  # Internal shared packages
│       ├── validator/        # Request validation
│       └── respond/          # Response helpers
│
├── pkg/                      # Public shared packages
│   ├── logger/               # Logging utilities
│   ├── errors/               # Error handling utilities
│   └── config/               # Configuration utilities
│
├── docs/                     # Documentation
│   └── openapi/              # Generated OpenAPI specs
│
├── migrations/               # Database migrations
│
└── scripts/                  # Build and deployment scripts
```
