#!/usr/bin/env bash
# Example: Create a service with autoscaling enabled
# This demonstrates using autoscaling instead of fixed replicas
# The service will scale between 2-10 replicas based on CPU and memory usage

iai services create my-autoscaling-service \
  --project demo-1 \
  --port 80 \
  --image-type external \
  --image-repository kennethreitz \
  --image-name httpbin \
  --image-tag latest \
  --memory 128Mi \
  --cpu 50m \
  --autoscaling-enabled \
  --autoscaling-min-replicas 2 \
  --autoscaling-max-replicas 10 \
  --autoscaling-cpu-percentage 80 \
  --autoscaling-memory-percentage 85 \
  --endpoint \
  --env APP_ENV=production \
  --env LOG_LEVEL=info
