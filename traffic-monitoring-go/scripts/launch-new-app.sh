#!/bin/bash

# Script to launch the new refactored V2X SIEM application

set -e

echo "🚀 Launching V2X SIEM API (Refactored Version)"

# Check if we're in the correct directory
if [ ! -f "go.mod" ] || [ ! -d "cmd/api" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Set default environment variables
export PORT=${PORT:-8080}
export GO_ENV=${GO_ENV:-development}
export LOG_LEVEL=${LOG_LEVEL:-info}
export JWT_SECRET=${JWT_SECRET:-your-super-secret-jwt-key-change-this-in-production}
export JWT_DURATION=${JWT_DURATION:-24h}

# Database configuration
export DSN=${DSN:-"host=db-go user=go_user password=go_pass dbname=go_db port=5432 sslmode=disable TimeZone=UTC"}

echo "📋 Configuration:"
echo "  PORT: $PORT"
echo "  GO_ENV: $GO_ENV"
echo "  LOG_LEVEL: $LOG_LEVEL"
echo "  DSN: $DSN"
echo ""

# Check if Docker is running for database
echo "🔍 Checking database connection..."
if ! docker ps | grep -q "traffic_db_go\|postgres"; then
    echo "⚠️  Warning: No PostgreSQL container found running."
    echo "   Make sure your database is running. You can start it with:"
    echo "   docker-compose up -d db-go"
    echo ""
    echo "   Or start the entire stack with:"
    echo "   docker-compose up -d"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Apply database migrations
echo "📊 Applying database migrations..."
if command -v goose &> /dev/null; then
    echo "Running migrations with goose..."
    cd migrations
    goose postgres "$DSN" up
    cd ..
    echo "✅ Migrations applied successfully"
else
    echo "⚠️  Warning: goose not found. Install it with:"
    echo "   go install github.com/pressly/goose/v3/cmd/goose@latest"
    echo ""
    echo "   Or apply migrations manually using psql"
    echo ""
    read -p "Continue without migrations? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Download dependencies
echo "📦 Installing dependencies..."
go mod download
go mod tidy

# Build the application
echo "🔨 Building application..."
go build -o bin/siem-api ./cmd/api

# Check if build was successful
if [ ! -f "bin/siem-api" ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Start the application
echo "🚀 Starting V2X SIEM API server..."
echo "   Server will be available at: http://localhost:$PORT"
echo "   Health check: http://localhost:$PORT/health"
echo "   API documentation will be available at: http://localhost:$PORT/api/v1"
echo ""
echo "📋 Default admin credentials:"
echo "   Email: admin@v2x-siem.com"
echo "   Password: admin123"
echo ""
echo "🔧 Use these endpoints to get started:"
echo "   POST /api/v1/auth/login - Login"
echo "   GET  /api/v1/auth/profile - Get current user"
echo "   GET  /api/v1/rules - List rules"
echo "   GET  /api/v1/alerts - List alerts"
echo ""
echo "Press Ctrl+C to stop the server"
echo "----------------------------------------"

./bin/siem-api