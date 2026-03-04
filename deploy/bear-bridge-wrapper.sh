#!/bin/sh
set -e

ENV_FILE="${HOME}/.config/bear-bridge/.env.bridge"
if [ ! -f "$ENV_FILE" ]; then
	echo "Error: $ENV_FILE not found" >&2
	exit 1
fi

set -a
. "$ENV_FILE"
set +a

# Rotate logs if larger than 5 MB (keeps one backup)
LOG_DIR="${HOME}/Library/Logs/bear-bridge"
MAX_SIZE=5242880
for log in "$LOG_DIR/stdout.log" "$LOG_DIR/stderr.log"; do
	if [ -f "$log" ]; then
		size=$(stat -f%z "$log" 2>/dev/null || echo 0)
		if [ "$size" -gt "$MAX_SIZE" ]; then
			mv "$log" "${log}.1"
		fi
	fi
done

exec "$(dirname "$0")/bear-bridge" "$@"
