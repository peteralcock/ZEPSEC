#!/bin/sh

set -e

echo "==> Removing stale PID file..."
rm -f tmp/pids/server.pid

echo "==> Waiting for PostgreSQL at ${DATABASE_HOST:-localhost}:${DATABASE_PORT:-5432}..."
until pg_isready -h "${DATABASE_HOST:-localhost}" -p "${DATABASE_PORT:-5432}" -U "${DATABASE_USER:-rism}" -q; do
  echo "    PostgreSQL not ready yet, retrying in 2s..."
  sleep 2
done
echo "==> PostgreSQL is ready."

echo "==> Running database migrations..."
bundle exec rake db:migrate 2>/dev/null || bundle exec rake db:setup
echo "==> Database is up to date."

echo "==> Starting Rails server..."
exec bundle exec rails s -b 0.0.0.0
