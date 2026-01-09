#!/bin/sh
set -e

# Try to increase inotify limits (best effort - may fail in restricted containers)
if [ -w /proc/sys/fs/inotify/max_user_watches ]; then
  echo 524288 > /proc/sys/fs/inotify/max_user_watches 2>/dev/null || true
fi

# Execute the command passed to this script
exec "$@"
