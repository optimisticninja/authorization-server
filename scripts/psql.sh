#!/bin/sh

PORT=${PORT:-5432}
HOST=${HOST:-localhost}
DB=${DB:-auth}
DB_USER=${DB_USER:-root}

if [ -z "$1" ]; then
	psql "postgres://${DB_USER}@${HOST}:${PORT}/${DB}"
else
	psql "postgres://${DB_USER}@${HOST}:${PORT}/${DB}" -f "$1"
fi
