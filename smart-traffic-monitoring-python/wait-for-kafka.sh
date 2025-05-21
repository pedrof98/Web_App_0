#!/bin/bash


set -e

host="$1"
port="$2"
shift 2
cmd="$@"

TIMEOUT=60
ATTEMPTS=0


until netcat -z -v -w5 "$host" "$port" > /dev/null 2>&1; do
	ATTEMPTS=$((ATTEMPTS+1))
  echo "Waiting for Kafka to be ready... ($ATTEMPTS attempts)"
  
  if [ $ATTEMPTS -gt $TIMEOUT ]; then
    echo "Timeout reached waiting for Kafka to start"
    exit 1
  fi
  
  sleep 2
done

echo "Kafka is up - executing command"
exec $cmd

