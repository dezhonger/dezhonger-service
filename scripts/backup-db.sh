#!/bin/sh
set -eu
umask 077

backup_directory="${BACKUP_DIRECTORY:-./backups}"
mkdir -p "$backup_directory"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="$backup_directory/dezhonger-$timestamp.sql.gz"

trap 'rm -f "$target.partial" "$target.partial.gz"' EXIT HUP INT TERM
docker compose exec -T db pg_dump -U dezhonger -d dezhonger > "$target.partial"
gzip -9 -c "$target.partial" > "$target.partial.gz"
mv "$target.partial.gz" "$target"
rm -f "$target.partial"
chmod 600 "$target"
printf 'database backup created: %s\n' "$target"
