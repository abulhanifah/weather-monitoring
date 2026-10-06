#!/bin/sh
set -e
# WITH_SEED=true -> teruskan flag --with-seed ke server
if [ "${WITH_SEED}" = "true" ]; then
  set -- "$@" --with-seed
fi
exec "$@"
