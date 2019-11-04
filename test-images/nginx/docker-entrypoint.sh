#!/bin/bash

TEAPOT="$(($RANDOM%2))"

if [ "${TEAPOT}" == 1 ]; then
  nginx -g 'daemon off;' -c /etc/nginx/teapot.conf
else
  nginx -g 'daemon off;' -c /etc/nginx/nginx.conf
fi
