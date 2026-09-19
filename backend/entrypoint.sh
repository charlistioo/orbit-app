#!/bin/sh
set -e

echo "Running database migrations..."
/app/migrate

echo "Starting ORBIT backend..."
exec /app/orbit-backend
