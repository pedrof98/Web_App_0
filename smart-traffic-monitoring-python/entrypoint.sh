#!/bin/bash
set -e

# If the first argument starts with 'alembic', execute it directly
if [[ "$1" == alembic* ]]; then
    exec "$@"
fi

# Otherwise, wait for Kafka and start the application
/wait-for-kafka.sh kafka 9092 python3 app/main.py
