#!/bin/sh

if [ $# -eq 0 ]; then
  exec ${SERVICE_NAME}
fi

exec "$@"
