#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log()   { echo -e "${GREEN}[mysql-init]${NC} $1"; }
warn()  { echo -e "${YELLOW}[mysql-init]${NC} $1"; }
err()   { echo -e "${RED}[mysql-init]${NC} $1"; }

wait_for_mysql() {
  local retries=30
  while [ $retries -gt 0 ]; do
    if mysqladmin ping -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" --silent 2>/dev/null; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 1
  done
  return 1
}

check_initialized() {
  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" -N -e \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$MYSQL_DATABASE' AND table_name = 'users'" 2>/dev/null || echo "0"
}

run_migrations() {
  log "Running migrations..."

  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "
    CREATE TABLE IF NOT EXISTS _migrations (
      id INT AUTO_INCREMENT PRIMARY KEY,
      filename VARCHAR(255) NOT NULL UNIQUE,
      applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
  " 2>/dev/null

  for migration_file in /docker-entrypoint-initdb.d/migrations/*.sql; do
    [ -f "$migration_file" ] || continue
    local migration_name
    migration_name=$(basename "$migration_file")

    local applied
    applied=$(mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -N -e \
      "SELECT COUNT(*) FROM _migrations WHERE filename = '${migration_name}'" 2>/dev/null || echo "0")

    if [ "$applied" -eq 0 ]; then
      log "Applying migration: $migration_name"
      if mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$migration_file" 2>&1; then
        mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e \
          "INSERT INTO _migrations (filename) VALUES ('${migration_name}')" 2>/dev/null
        log "Migration applied: $migration_name"
      else
        warn "Migration failed (may already be applied): $migration_name"
      fi
    fi
  done

  log "Migrations complete."
}

main() {
  log "Starting MySQL initialization..."

  if ! wait_for_mysql; then
    err "MySQL failed to start after 30 retries"
    exit 1
  fi
  log "MySQL is ready."

  local table_count
  table_count=$(check_initialized)

  if [ "$table_count" -gt 0 ]; then
    log "Database already initialized. Running pending migrations only."
    run_migrations
    return
  fi

  log "Database not initialized. Running full setup..."

  log "Step 1/5: Running schema.sql..."
  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /docker-entrypoint-initdb.d/01-schema.sql
  log "Schema created."

  log "Step 2/5: Running migrations..."
  run_migrations

  log "Step 3/5: Running seed-rbac.sql..."
  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /docker-entrypoint-initdb.d/03-seed-rbac.sql
  log "RBAC seed data inserted."

  log "Step 4/5: Creating admin user..."
  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "
    INSERT INTO users (username, email, password_hash, is_active) VALUES
      ('admin', 'admin@siakad.local', '\$2a\$10\$KPfieEgX7iYGe4V.nIQrou/r8fUcUnLK0Q.FTIcDKvUNgNNygGwPC', 1)
    ON DUPLICATE KEY UPDATE username = VALUES(username);
  "
  log "Admin user created (username: admin, password: Admin123!)."

  log "Step 5/5: Assigning admin role..."
  mysql -h localhost -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "
    INSERT INTO user_roles (user_id, role_id)
    SELECT u.id, r.id
    FROM users u, roles r
    WHERE u.username = 'admin' AND r.code = 'admin'
      AND u.deleted_at IS NULL AND r.deleted_at IS NULL
    ON DUPLICATE KEY UPDATE user_id = user_id;
  "
  log "Admin role assigned."

  log "=== Database initialization complete! ==="
}

main
