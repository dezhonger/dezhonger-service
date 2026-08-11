#!/bin/sh
set -eu

backup_directory="${BACKUP_DIRECTORY:-./backups}"
mkdir -p "$backup_directory"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="$backup_directory/dezhonger-$timestamp.sql.gz"

docker compose exec -T db pg_dump -U dezhonger -d dezhonger \
  | gzip -9 > "$target"
chmod 600 "$target"
printf 'database backup created: %s\n' "$target"
