#!/bin/bash

ETCD_HOST=etcd
ETCD_PORT=2379
ETCD_URL=http://$ETCD_HOST:$ETCD_PORT

echo ETCD_URL = $ETCD_URL
if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."
  agent consumer 20000 20001 $ETCD_URL 6
elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."
  agent provider 30000 20880 $ETCD_URL 1
elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."
  agent provider 30000 20880 $ETCD_URL 2
elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."
  agent provider 30000 20880 $ETCD_URL 3
else
  echo "Unrecognized arguments, exit."
  exit 1
fi
