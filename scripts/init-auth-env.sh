#!/bin/sh
set -eu
umask 077
if [ ! -f .env.auth.local ]; then
  secret=$(openssl rand -hex 32)
  printf 'BETTER_AUTH_SECRET=%s\nGOOGLE_CLIENT_ID=\nGOOGLE_CLIENT_SECRET=\n' "$secret" > .env.auth.local
  echo 'Created .env.auth.local (development only).'
fi
mkdir -p .auth-mail
