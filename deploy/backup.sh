#!/usr/bin/env bash
set -euo pipefail
umask 077

cd "$(dirname "$0")"
mkdir -p backups

file="backups/tatusash-$(date +%F-%H%M).sql.gz"

docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB"' | gzip > "$file.tmp"
mv "$file.tmp" "$file"

find backups -name 'tatusash-*.sql.gz' -mtime +14 -delete