#!/usr/bin/env bash
# Example: Create a service with fixed number of replicas
# This demonstrates using a fixed replica count instead of autoscaling
# The service will always run exactly 3 replicas

iai services create my-fixed-service \
  --project my-project \
  --port 80 \
  --image-type external \
  --image-repository kennethreitz \
  --image-name httpbin \
  --image-tag latest \
  --memory 128Mi \
  --cpu 50m \
  --replicas 3 \
  --endpoint \
  --env APP_ENV=production \
  --env LOG_LEVEL=info
