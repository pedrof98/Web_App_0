# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
- **Development with hot reload**: `docker-compose up` - Starts the full stack including PostgreSQL, Elasticsearch, Kibana, and the main application
- **Build production binary**: `go build -o traffic-monitoring-go ./app/main.go`
- **Run integration tests**: `./scripts/run-integration-tests.sh`

### Test Commands
- **Run tests**: `go test ./tests/integration/...` (requires PostgreSQL and Elasticsearch running)
- **Run with test environment**: Use the integration test script which sets up isolated containers

### Database
- **Connection**: PostgreSQL on port 5420 (development), credentials in docker-compose.yml
- **Migrations**: Located in `migrations/` directory, run automatically on startup
- **Test database**: Separate containers created by integration test script

## Architecture Overview

This is a **Vehicle-to-Everything (V2X) Security Information and Event Management (SIEM) system** built in Go. The project is implemented in three phases as documented in `SIEM_PHASES.md`.

### Core Components

1. **Main Application** (`app/main.go`):
   - Gin HTTP server on port 8080
   - Initializes database, Elasticsearch, and V2X collectors
   - DSRC collector on port 5001, CV2X collector on port 5002

2. **Data Models** (`app/models/`):
   - **Traffic models**: Station, Sensor, TrafficMeasurement, UserEvent
   - **Security models**: SecurityEvent, Alert, Rule, LogSource, User
   - **V2X models**: BSM (Basic Safety Messages), PLMNInfo, V2XMessage

3. **SIEM Core** (`app/siem/`):
   - **Rule Engine**: Evaluates security events against detection rules
   - **Collectors**: DSRC and CV2X protocol handlers with enhanced security detection
   - **Elasticsearch Integration**: Event storage and search capabilities
   - **V2X Security**: Specialized automotive threat detection

4. **API Routes** (`app/routes/routes.go`):
   - RESTful endpoints for all entities
   - Security event ingestion at `/ingest`
   - Dashboard endpoints for visualization
   - V2X-specific dashboards and testing endpoints

### Key Architecture Patterns

- **Multi-protocol Support**: Handles both DSRC and Cellular V2X communications
- **Real-time Processing**: UDP collectors process V2X messages as they arrive
- **Rule-based Detection**: Configurable security rules with severity levels
- **Elasticsearch Integration**: For log storage, search, and analytics
- **Phase-based Development**: Progressive enhancement from basic SIEM to V2X-specialized security

### V2X-Specific Features

- **Protocol Parsers**: J2735 and CV2X message parsing
- **Security Rules**: V2X-specific threat detection (message replay, spoofing, anomalies)
- **Geographic Correlation**: Location-based security analysis
- **Automotive Compliance**: Supports automotive security standards

### Microservices

- **data-generator**: Generates synthetic V2X events and attack scenarios
- **v2x-simulator**: Simulates V2X communication patterns for testing

### External Dependencies

- **PostgreSQL**: Primary database for structured data
- **Elasticsearch**: Log storage and analytics
- **Kibana**: Visualization and dashboards
- **Docker**: Containerization and orchestration

## Development Notes

- The system processes real-time V2X messages from vehicles
- Security events are evaluated against customizable detection rules
- Elasticsearch integration provides advanced search and analytics
- The collector system is designed for high-throughput message processing
- Geographic and temporal correlation enables sophisticated threat detection