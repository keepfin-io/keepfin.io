#!/bin/sh
set -eu
umask 077
if ! test -f .env || ! grep -q '^TYPESENSE_API_KEY=.' .env; then
  printf '\nTYPESENSE_API_KEY=%s\n' "$(openssl rand -hex 32)" >> .env
  echo 'Created a private development search key in .env.'
fi
