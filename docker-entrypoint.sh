#!/bin/sh
set -e

echo "Starting ASAM Backend..."
echo "PORT=${PORT:-8080}"
echo "ENVIRONMENT=${ENVIRONMENT:-production}"

# Execute the main application
exec "$@"
