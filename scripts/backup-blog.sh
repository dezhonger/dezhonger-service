#!/bin/sh
# Run from the service checkout. Files and the database are both required to restore posts.
set -eu
umask 077

backup_directory="${BACKUP_DIRECTORY:-./backups}"
mkdir -p "$backup_directory"
./scripts/backup-db.sh
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="$backup_directory/blog-media-$timestamp.tar.gz"
trap 'rm -f "$target.partial"' EXIT HUP INT TERM
docker compose exec -T api tar -C /var/lib/dezhonger/blog-media -czf - . > "$target.partial"
mv "$target.partial" "$target"
chmod 600 "$target"
printf 'blog media backup created: %s\n' "$target"
