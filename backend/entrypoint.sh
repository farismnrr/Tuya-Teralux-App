#!/bin/bash
set -e

echo "🚀 Starting Teralux Backend..."

if [ "$AUTO_MIGRATE" = "true" ]; then
  # Wait for Database to be ready
  echo "⏳ Waiting for Database to be ready..."
  # Check if using MySQL or Postgres (defaulting to MySQL for checks if needed, but here using pg_isready is specific to postgres!)
  # The user moved to MySQL. pg_isready is WRONG for MySQL.
  # I need to use mysqladmin ping or similar, OR just rely on the app/migrate tool to retry?
  # The Dockerfile installs postgresql-client, I should install default-mysql-client or similar.
  # The migrate tool handles connection, but waiting is good.
  # Let's fix the wait logic too.
  
  until mysqladmin ping -h "$DB_HOST" -P "$DB_PORT" --silent; do
      echo "Database is unavailable - sleeping"
      sleep 2
  done

  echo "✅ Database is ready!"

  # Run database migrations
  echo "🔄 Running database migrations..."
  migrate -path ./migrations -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME?charset=utf8mb4&parseTime=True&loc=Local" up

  if [ $? -eq 0 ]; then
    echo "✅ Migrations completed successfully!"
  else
    echo "❌ Migration failed!"
    exit 1
  fi
fi

# Start the application
echo "🚀 Starting application..."
exec ./main
