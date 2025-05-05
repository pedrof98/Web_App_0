#!/bin/bash

# Stop on first error
set -e

echo "Starting V2X SIEM integration tests..."

# Create a temporary network for testing
echo "Creating test network..."
NETWORK_NAME="siem-test-network"
docker network create $NETWORK_NAME || true

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker run --rm -d \
  --name siem-test-db \
  --network $NETWORK_NAME \
  -e POSTGRES_USER=test_user \
  -e POSTGRES_PASSWORD=test_pass \
  -e POSTGRES_DB=test_db \
  postgres:15

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if docker exec siem-test-db pg_isready -U test_user -d test_db; then
    echo "PostgreSQL is ready"
    break
  fi
  
  if [ $i -eq 30 ]; then
    echo "Error: PostgreSQL not ready after 30 seconds"
    exit 1
  fi
  
  echo "Waiting for PostgreSQL to be ready... ${i}/30"
  sleep 1
done

# Start Elasticsearch container
echo "Starting Elasticsearch container..."
docker run --rm -d \
  --name siem-test-es \
  --network $NETWORK_NAME \
  -e "discovery.type=single-node" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  docker.elastic.co/elasticsearch/elasticsearch:7.17.0

# Wait for Elasticsearch to be ready
echo "Waiting for Elasticsearch to be ready..."
for i in {1..60}; do
  if docker exec siem-test-es curl -s 'http://localhost:9200/_cluster/health' | grep -q '"status":"green\|yellow"'; then
    echo "Elasticsearch is ready"
    break
  fi
  
  if [ $i -eq 60 ]; then
    echo "Error: Elasticsearch not ready after 60 seconds"
    exit 1
  fi
  
  echo "Waiting for Elasticsearch to be ready... ${i}/60"
  sleep 1
done

# Build the SIEM app for testing
echo "Building SIEM app for testing..."
docker build -t siem-test-app .

# Run integration tests
echo "Running integration tests..."
docker run --rm \
  --name siem-integration-tests \
  --network $NETWORK_NAME \
  -e DSN="host=siem-test-db user=test_user password=test_pass dbname=test_db port=5432 sslmode=disable TimeZone=UTC" \
  -e ELASTICSEARCH_URL="http://siem-test-es:9200" \
  -e GO_ENV=test \
  siem-test-app \
  go test -v ./tests/integration/...

# Cleanup
echo "Cleaning up..."
docker stop siem-test-db siem-test-es

# Remove the network
docker network rm $NETWORK_NAME || true

echo "Integration tests completed successfully"