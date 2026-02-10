#!/usr/bin/env bash
# Example: Create a service with a downtime schedule
# The service will scale down on weekends (America/New_York)

iai services create my-scheduled-service \
  --project demo-1 \
  --port 80 \
  --image-type external \
  --image-repository kennethreitz \
  --image-name httpbin \
  --image-tag latest \
  --replicas 1 \
  --memory 128M \
  --cpu 50m \
  --schedule-downtime "Tue-Wed 00:00-24:00" \
  --schedule-timezone Europe/Madrid \
  --endpoint
