#!/usr/bin/env bash
# Example: Create a service with an HTTP healthcheck
# The platform will probe /health after an initial 10-second delay

iai services create my-healthcheck-service \
  --project demo-1 \
  --port 80 \
  --image-type external \
  --image-repository kennethreitz \
  --image-name httpbin \
  --image-tag latest \
  --replicas 1 \
  --memory 128M \
  --cpu 50m \
  --healthcheck-enabled \
  --healthcheck-path /health \
  --healthcheck-initial-delay 10 \
  --endpoint
