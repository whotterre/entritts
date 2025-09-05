#!/bin/sh 
# wait-for-it.sh: Wait until a host:port is available

set -e

host="$1"
port="$2"
shift 2

while ! nc -z "$host" "$port"; do
  echo "Waiting for $host:$port..."
  sleep 1
done

echo "$host:$port is available!"
exec "$@"
